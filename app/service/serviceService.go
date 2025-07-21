package service

import (
	"context"
	"fmt"
	"go_service/app/common"
	"go_service/app/model"
	"go_service/pkg/utils"
	"strconv"
	"sync"
	"time"

	"gorm.io/gorm"
)

type ServiceService struct {
	db    *gorm.DB
	mutex sync.RWMutex // 读写锁保护并发操作
}

func NewServiceService(db *gorm.DB) *ServiceService {
	return &ServiceService{
		db: db,
	}
}

// CreateService 创建服务
func (s *ServiceService) CreateService(ctx context.Context, service *model.ServiceModel) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 验证服务数据
	if err := service.Validate(); err != nil {
		return common.WrapError(common.ErrCodeInvalidParam, "服务数据验证失败", err)
	}

	// 检查端口是否已被占用
	if err := s.checkPortAvailable(service.Port, 0); err != nil {
		return err
	}

	// 检查服务名是否已存在
	if err := s.checkNameAvailable(service.Name, 0); err != nil {
		return err
	}

	// 创建服务
	if err := s.db.WithContext(ctx).Create(service).Error; err != nil {
		return common.WrapError(common.ErrCodeDatabaseError, "创建服务失败", err)
	}

	return nil
}

// UpdateService 更新服务
func (s *ServiceService) UpdateService(ctx context.Context, service *model.ServiceModel) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if service.Id <= 0 {
		return common.NewBusinessError(common.ErrCodeInvalidParam, "无效的服务ID")
	}

	// 验证服务数据
	if err := service.Validate(); err != nil {
		return common.WrapError(common.ErrCodeInvalidParam, "服务数据验证失败", err)
	}

	// 检查服务是否存在
	var existing model.ServiceModel
	if err := s.db.WithContext(ctx).First(&existing, service.Id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return common.ErrServiceNotFound
		}
		return common.WrapError(common.ErrCodeDatabaseError, "查询服务失败", err)
	}

	// 检查端口是否被其他服务占用
	if err := s.checkPortAvailable(service.Port, service.Id); err != nil {
		return err
	}

	// 检查服务名是否被其他服务占用
	if err := s.checkNameAvailable(service.Name, service.Id); err != nil {
		return err
	}

	// 更新服务
	if err := s.db.WithContext(ctx).Model(&existing).Updates(service).Error; err != nil {
		return common.WrapError(common.ErrCodeDatabaseError, "更新服务失败", err)
	}

	return nil
}

// GetServiceById 根据ID获取服务
func (s *ServiceService) GetServiceById(ctx context.Context, id int64) (*model.ServiceModel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var service model.ServiceModel
	if err := s.db.WithContext(ctx).First(&service, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrServiceNotFound
		}
		return nil, common.WrapError(common.ErrCodeDatabaseError, "查询服务失败", err)
	}

	return &service, nil
}

// GetServiceByName 根据名称获取服务
func (s *ServiceService) GetServiceByName(ctx context.Context, name string) (*model.ServiceModel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var service model.ServiceModel
	if err := s.db.WithContext(ctx).Where("name = ?", name).First(&service).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrServiceNotFound
		}
		return nil, common.WrapError(common.ErrCodeDatabaseError, "查询服务失败", err)
	}

	return &service, nil
}

// DeleteService 删除服务
func (s *ServiceService) DeleteService(ctx context.Context, id int64) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 检查服务是否存在
	service, err := s.GetServiceById(ctx, id)
	if err != nil {
		return err
	}

	// 检查服务是否正在运行
	port := strconv.Itoa(int(service.Port))
	if isRunning, _ := utils.IsPortInUse(port); isRunning {
		return common.NewBusinessError(common.ErrCodeServiceRunning, "无法删除正在运行的服务，请先停止服务")
	}

	// 删除服务
	if err := s.db.WithContext(ctx).Delete(&model.ServiceModel{}, id).Error; err != nil {
		return common.WrapError(common.ErrCodeDatabaseError, "删除服务失败", err)
	}

	return nil
}

// ListServices 获取服务列表
func (s *ServiceService) ListServices(ctx context.Context, req *model.ServiceListRequest) (*model.ServiceListResponse, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// 设置默认分页参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	query := s.db.WithContext(ctx).Model(&model.ServiceModel{})

	// 按名称过滤
	if req.Name != "" {
		query = query.Where("name LIKE ?", "%"+req.Name+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, common.WrapError(common.ErrCodeDatabaseError, "查询服务总数失败", err)
	}

	// 获取服务列表
	var services []model.ServiceModel
	offset := (req.Page - 1) * req.PageSize
	if err := query.Offset(offset).Limit(req.PageSize).Find(&services).Error; err != nil {
		return nil, common.WrapError(common.ErrCodeDatabaseError, "查询服务列表失败", err)
	}

	// 获取端口状态信息
	portList, err := utils.GetPortList()
	if err != nil {
		return nil, common.WrapError(common.ErrCodeCommandFailed, "获取端口状态失败", err)
	}

	// 构建带状态的服务列表
	var serviceStatuses []model.ServiceStatusModel
	for _, service := range services {
		status := s.buildServiceStatus(service, portList)
		
		// 按状态过滤
		if req.Status != nil && status.Status != *req.Status {
			continue
		}
		
		serviceStatuses = append(serviceStatuses, status)
	}

	return &model.ServiceListResponse{
		List:  serviceStatuses,
		Total: total,
		Page:  req.Page,
		Size:  req.PageSize,
	}, nil
}

// GetAllServicesWithStatus 获取所有服务及其状态
func (s *ServiceService) GetAllServicesWithStatus(ctx context.Context) ([]model.ServiceStatusModel, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var services []model.ServiceModel
	if err := s.db.WithContext(ctx).Find(&services).Error; err != nil {
		return nil, common.WrapError(common.ErrCodeDatabaseError, "查询服务列表失败", err)
	}

	if len(services) == 0 {
		return []model.ServiceStatusModel{}, nil
	}

	// 获取端口状态信息
	portList, err := utils.GetPortList()
	if err != nil {
		return nil, common.WrapError(common.ErrCodeCommandFailed, "获取端口状态失败", err)
	}

	// 构建带状态的服务列表
	var serviceStatuses []model.ServiceStatusModel
	for _, service := range services {
		status := s.buildServiceStatus(service, portList)
		serviceStatuses = append(serviceStatuses, status)
	}

	return serviceStatuses, nil
}

// checkPortAvailable 检查端口是否可用
func (s *ServiceService) checkPortAvailable(port int64, excludeId int64) error {
	var existing model.ServiceModel
	query := s.db.Where("port = ?", port)
	if excludeId > 0 {
		query = query.Where("id != ?", excludeId)
	}
	
	if err := query.First(&existing).Error; err == nil {
		return common.NewBusinessError(common.ErrCodePortInUse, 
			fmt.Sprintf("端口 %d 已被服务 '%s' 占用", port, existing.Name))
	} else if err != gorm.ErrRecordNotFound {
		return common.WrapError(common.ErrCodeDatabaseError, "检查端口占用失败", err)
	}
	
	return nil
}

// checkNameAvailable 检查服务名是否可用
func (s *ServiceService) checkNameAvailable(name string, excludeId int64) error {
	var existing model.ServiceModel
	query := s.db.Where("name = ?", name)
	if excludeId > 0 {
		query = query.Where("id != ?", excludeId)
	}
	
	if err := query.First(&existing).Error; err == nil {
		return common.NewBusinessError(common.ErrCodeInvalidParam, 
			fmt.Sprintf("服务名称 '%s' 已存在", name))
	} else if err != gorm.ErrRecordNotFound {
		return common.WrapError(common.ErrCodeDatabaseError, "检查服务名称失败", err)
	}
	
	return nil
}

// buildServiceStatus 构建服务状态信息
func (s *ServiceService) buildServiceStatus(service model.ServiceModel, portList map[string]map[string]interface{}) model.ServiceStatusModel {
	status := model.ServiceStatusModel{
		ServiceModel: service,
		Status:       0, // 默认停止状态
		Pid:          "",
		Process:      "",
	}

	port := strconv.Itoa(int(service.Port))
	if portInfo, exists := portList[port]; exists {
		status.Status = 1 // 运行状态
		if pid, ok := portInfo["pid"].(string); ok {
			status.Pid = pid
		}
		if process, ok := portInfo["process"].(string); ok {
			status.Process = process
		}
	}

	return status
}

// HealthCheck 健康检查
func (s *ServiceService) HealthCheck(ctx context.Context) map[string]interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var total int64
	var running int64

	// 获取服务总数
	s.db.WithContext(ctx).Model(&model.ServiceModel{}).Count(&total)

	// 获取运行中的服务数量
	services, err := s.GetAllServicesWithStatus(ctx)
	if err == nil {
		for _, service := range services {
			if service.Status == 1 {
				running++
			}
		}
	}

	return map[string]interface{}{
		"total_services":   total,
		"running_services": running,
		"stopped_services": total - running,
		"timestamp":        time.Now().Unix(),
	}
}