package service

import (
	"context"
	"go_service/app/common"
	"go_service/app/model"
	"log"
	"sync"
	"time"

	"gorm.io/gorm"
)

type LogService struct {
	db          *gorm.DB
	mutex       sync.RWMutex
	logChannel  chan model.ServiceLog
	stopChannel chan struct{}
	wg          sync.WaitGroup
	
	// 性能优化配置
	batchSize     int
	flushInterval time.Duration
	channelSize   int
	
	// 统计信息
	stats struct {
		sync.RWMutex
		totalLogs    int64
		failedWrites int64
		lastFlush    time.Time
	}
}

func NewLogService(db *gorm.DB) *LogService {
	service := &LogService{
		db:            db,
		batchSize:     50,
		flushInterval: 5 * time.Second,
		channelSize:   1000,
		stopChannel:   make(chan struct{}),
	}
	
	service.logChannel = make(chan model.ServiceLog, service.channelSize)
	
	// 启动日志处理协程
	service.startLogProcessor()
	return service
}

// startLogProcessor 启动日志处理器
func (l *LogService) startLogProcessor() {
	l.wg.Add(1)
	go func() {
		defer l.wg.Done()
		
		// 批量处理日志
		batch := make([]model.ServiceLog, 0, 100)
		ticker := time.NewTicker(5 * time.Second) // 每5秒批量写入一次
		defer ticker.Stop()
		
		for {
			select {
			case logEntry := <-l.logChannel:
				batch = append(batch, logEntry)
				
				// 批量大小达到阈值时立即写入
				if len(batch) >= 50 {
					l.flushLogs(batch)
					batch = batch[:0] // 清空切片
				}
				
			case <-ticker.C:
				// 定时写入
				if len(batch) > 0 {
					l.flushLogs(batch)
					batch = batch[:0]
				}
				
			case <-l.stopChannel:
				// 停止前写入剩余日志
				if len(batch) > 0 {
					l.flushLogs(batch)
				}
				return
			}
		}
	}()
}

// flushLogs 批量写入日志
func (l *LogService) flushLogs(logs []model.ServiceLog) {
	if len(logs) == 0 {
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := l.db.WithContext(ctx).CreateInBatches(logs, l.batchSize).Error; err != nil {
		log.Printf("批量写入日志失败: %v", err)
		l.stats.Lock()
		l.stats.failedWrites++
		l.stats.Unlock()
	} else {
		l.stats.Lock()
		l.stats.totalLogs += int64(len(logs))
		l.stats.lastFlush = time.Now()
		l.stats.Unlock()
	}
}

// LogOperation 记录操作日志 - 异步优化版本
func (l *LogService) LogOperation(ctx context.Context, serviceId int64, operation, status, output, errorMsg string, duration time.Duration) {
	logEntry := model.ServiceLog{
		ServiceId: serviceId,
		Operation: operation,
		Status:    status,
		Output:    truncateString(output, 10000), // 限制输出长度
		Error:     truncateString(errorMsg, 5000), // 限制错误信息长度
		Duration:  duration.Milliseconds(),
	}
	
	// 非阻塞写入通道
	select {
	case l.logChannel <- logEntry:
		// 成功写入通道
	default:
		// 通道满了，直接写数据库（降级处理）
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := l.db.WithContext(ctx).Create(&logEntry).Error; err != nil {
				log.Printf("直接写入日志失败: %v", err)
			}
		}()
	}
}

// truncateString 截断字符串
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...[truncated]"
}

// RecordServiceOperation 记录服务操作日志
func (l *LogService) RecordServiceOperation(ctx context.Context, serviceId int64, operation, status, output, errorMsg string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	log := &model.ServiceLog{
		ServiceId: serviceId,
		Operation: operation,
		Status:    status,
		Output:    output,
		Error:     errorMsg,
	}

	if err := l.db.WithContext(ctx).Create(log).Error; err != nil {
		return common.WrapError(common.ErrCodeDatabaseError, "记录操作日志失败", err)
	}

	return nil
}

// GetServiceLogs 获取服务操作日志
func (l *LogService) GetServiceLogs(ctx context.Context, serviceId int64, limit int) ([]model.ServiceLog, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	var logs []model.ServiceLog
	query := l.db.WithContext(ctx).Where("service_id = ?", serviceId).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Find(&logs).Error; err != nil {
		return nil, common.WrapError(common.ErrCodeDatabaseError, "查询操作日志失败", err)
	}

	return logs, nil
}

// GetAllLogs 获取所有服务操作日志
func (l *LogService) GetAllLogs(ctx context.Context, page, pageSize int) (*model.LogListResponse, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}

	var total int64
	if err := l.db.WithContext(ctx).Model(&model.ServiceLog{}).Count(&total).Error; err != nil {
		return nil, common.WrapError(common.ErrCodeDatabaseError, "查询日志总数失败", err)
	}

	var logs []model.ServiceLog
	offset := (page - 1) * pageSize
	if err := l.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error; err != nil {
		return nil, common.WrapError(common.ErrCodeDatabaseError, "查询操作日志失败", err)
	}

	return &model.LogListResponse{
		List:  logs,
		Total: total,
		Page:  page,
		Size:  pageSize,
	}, nil
}

// CleanOldLogs 清理旧日志
func (l *LogService) CleanOldLogs(ctx context.Context, days int) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	if days <= 0 {
		days = 30 // 默认保留30天
	}

	cutoffTime := time.Now().AddDate(0, 0, -days)
	
	result := l.db.WithContext(ctx).Where("created_at < ?", cutoffTime).Delete(&model.ServiceLog{})
	if result.Error != nil {
		return common.WrapError(common.ErrCodeDatabaseError, "清理旧日志失败", result.Error)
	}

	return nil
}

// GetLogStats 获取日志统计信息
func (l *LogService) GetLogStats(ctx context.Context) (map[string]interface{}, error) {
	l.mutex.RLock()
	defer l.mutex.RUnlock()

	stats := make(map[string]interface{})

	// 总日志数
	var totalLogs int64
	l.db.WithContext(ctx).Model(&model.ServiceLog{}).Count(&totalLogs)
	stats["total_logs"] = totalLogs

	// 今日日志数
	today := time.Now().Format("2006-01-02")
	var todayLogs int64
	l.db.WithContext(ctx).Model(&model.ServiceLog{}).Where("DATE(created_at) = ?", today).Count(&todayLogs)
	stats["today_logs"] = todayLogs

	// 成功/失败统计
	var successLogs, failedLogs int64
	l.db.WithContext(ctx).Model(&model.ServiceLog{}).Where("status = ?", "success").Count(&successLogs)
	l.db.WithContext(ctx).Model(&model.ServiceLog{}).Where("status = ?", "failed").Count(&failedLogs)
	stats["success_logs"] = successLogs
	stats["failed_logs"] = failedLogs

	// 操作类型统计
	var operationStats []map[string]interface{}
	l.db.WithContext(ctx).Model(&model.ServiceLog{}).
		Select("operation, COUNT(*) as count").
		Group("operation").
		Scan(&operationStats)
	stats["operation_stats"] = operationStats

	return stats, nil
}

// Close 关闭日志服务
func (l *LogService) Close() {
	close(l.stopChannel)
	l.wg.Wait()
	close(l.logChannel)
}

// GetServiceStats 获取日志服务统计信息
func (l *LogService) GetServiceStats() map[string]interface{} {
	l.stats.RLock()
	defer l.stats.RUnlock()
	
	return map[string]interface{}{
		"total_logs":     l.stats.totalLogs,
		"failed_writes":  l.stats.failedWrites,
		"last_flush":     l.stats.lastFlush,
		"channel_size":   l.channelSize,
		"batch_size":     l.batchSize,
		"flush_interval": l.flushInterval.String(),
		"channel_usage":  len(l.logChannel),
	}
}

// SetBatchSize 设置批量大小
func (l *LogService) SetBatchSize(size int) {
	if size > 0 && size <= 1000 {
		l.batchSize = size
	}
}

// SetFlushInterval 设置刷新间隔
func (l *LogService) SetFlushInterval(interval time.Duration) {
	if interval >= time.Second && interval <= time.Minute {
		l.flushInterval = interval
	}
}