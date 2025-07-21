package service

import (
	"context"
	"fmt"
	"go_service/app/common"
	"go_service/pkg/utils"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"gorm.io/gorm"
)

type CommandService struct {
	db             *gorm.DB
	serviceService *ServiceService
	logService     *LogService
	mutex          sync.RWMutex
	commandTimeout time.Duration
}

func NewCommandService(db *gorm.DB) *CommandService {
	return &CommandService{
		db:             db,
		serviceService: NewServiceService(db),
		logService:     NewLogService(db),
		commandTimeout: 5 * time.Second, // 默认5秒超时
	}
}

// SetCommandTimeout 设置命令执行超时时间
func (c *CommandService) SetCommandTimeout(timeout time.Duration) {
	c.commandTimeout = timeout
}

// StartService 启动服务
func (c *CommandService) StartService(ctx context.Context, serviceId int64) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	startTime := time.Now()

	// 获取服务信息
	service, err := c.serviceService.GetServiceById(ctx, serviceId)
	if err != nil {
		c.logService.LogOperation(ctx, serviceId, "start", "failed", "", err.Error(), time.Since(startTime))
		return "", err
	}

	if service.CmdStart == "" {
		err := common.NewBusinessError(common.ErrCodeInvalidParam, "启动命令未配置")
		c.logService.LogOperation(ctx, serviceId, "start", "failed", "", err.Error(), time.Since(startTime))
		return "", err
	}

	// 检查服务是否已在运行
	port := strconv.Itoa(int(service.Port))
	if isRunning, _ := utils.IsPortInUse(port); isRunning {
		err := common.NewBusinessError(common.ErrCodeServiceRunning, "服务已在运行")
		c.logService.LogOperation(ctx, serviceId, "start", "failed", "", err.Error(), time.Since(startTime))
		return "", err
	}

	// 执行启动命令
	output, err := c.executeCommand(ctx, service.CmdStart, service.Dir)
	if err != nil {
		c.logService.LogOperation(ctx, serviceId, "start", "failed", output, err.Error(), time.Since(startTime))
		return output, common.WrapError(common.ErrCodeCommandFailed, "启动服务失败", err)
	}

	// 等待服务启动完成
	if err := c.waitForServiceStart(port, 3*time.Second); err != nil {
		c.logService.LogOperation(ctx, serviceId, "start", "failed", output, err.Error(), time.Since(startTime))
		return output, common.WrapError(common.ErrCodeCommandFailed, "服务启动超时", err)
	}

	// 记录成功日志
	c.logService.LogOperation(ctx, serviceId, "start", "success", output, "", time.Since(startTime))
	return output, nil
}

// StopService 停止服务
func (c *CommandService) StopService(ctx context.Context, serviceId int64) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	startTime := time.Now()

	// 获取服务信息
	service, err := c.serviceService.GetServiceById(ctx, serviceId)
	if err != nil {
		c.logService.LogOperation(ctx, serviceId, "stop", "failed", "", err.Error(), time.Since(startTime))
		return "", err
	}

	port := strconv.Itoa(int(service.Port))
	if isRunning, _ := utils.IsPortInUse(port); !isRunning {
		err := common.NewBusinessError(common.ErrCodeServiceStopped, "服务未运行")
		c.logService.LogOperation(ctx, serviceId, "stop", "failed", "", err.Error(), time.Since(startTime))
		return "", err
	}

	var output string
	var finalErr error

	// 优先使用停止命令
	if service.CmdStop != "" {
		output, err = c.executeCommand(ctx, service.CmdStop, service.Dir)
		if err != nil {
			// 停止命令失败，尝试强制终止
			killOutput, killErr := utils.Kill(port)
			if killErr != nil {
				finalErr = common.WrapError(common.ErrCodeCommandFailed, "停止服务失败", err)
				c.logService.LogOperation(ctx, serviceId, "stop", "failed", output, finalErr.Error(), time.Since(startTime))
				return output, finalErr
			}
			output = fmt.Sprintf("停止命令失败，已强制终止:\n%s\n强制终止输出:\n%s", output, killOutput)
		}
	} else {
		// 没有停止命令，直接强制终止
		output, err = utils.Kill(port)
		if err != nil {
			finalErr = common.WrapError(common.ErrCodeCommandFailed, "强制终止服务失败", err)
			c.logService.LogOperation(ctx, serviceId, "stop", "failed", output, finalErr.Error(), time.Since(startTime))
			return output, finalErr
		}
	}

	// 等待服务停止完成
	if err := c.waitForServiceStop(port, 3*time.Second); err != nil {
		finalErr = common.WrapError(common.ErrCodeCommandFailed, "服务停止超时", err)
		c.logService.LogOperation(ctx, serviceId, "stop", "failed", output, finalErr.Error(), time.Since(startTime))
		return output, finalErr
	}

	// 记录成功日志
	c.logService.LogOperation(ctx, serviceId, "stop", "success", output, "", time.Since(startTime))
	return output, nil
}

// RestartService 重启服务
func (c *CommandService) RestartService(ctx context.Context, serviceId int64) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 获取服务信息
	service, err := c.serviceService.GetServiceById(ctx, serviceId)
	if err != nil {
		return "", err
	}

	port := strconv.Itoa(int(service.Port))
	if isRunning, _ := utils.IsPortInUse(port); !isRunning {
		return "", common.NewBusinessError(common.ErrCodeServiceStopped, "服务未运行")
	}

	var output string
	// 优先使用重启命令
	if service.CmdRestart != "" {
		output, err = c.executeCommand(ctx, service.CmdRestart, service.Dir)
		if err != nil {
			return output, common.WrapError(common.ErrCodeCommandFailed, "重启服务失败", err)
		}
	} else {
		// 没有重启命令，先停止再启动
		c.mutex.Unlock() // 临时释放锁避免死锁
		stopOutput, err := c.StopService(ctx, serviceId)
		if err != nil {
			c.mutex.Lock()
			return stopOutput, err
		}

		// 等待端口释放
		if err := utils.WaitForPortFree(port, 5*time.Second); err != nil {
			c.mutex.Lock()
			return stopOutput, common.WrapError(common.ErrCodeCommandFailed, "等待端口释放失败", err)
		}

		startOutput, err := c.StartService(ctx, serviceId)
		c.mutex.Lock()
		if err != nil {
			return startOutput, err
		}
		output = fmt.Sprintf("停止输出:\n%s\n启动输出:\n%s", stopOutput, startOutput)
	}

	return output, nil
}

// ForceRestartService 强制重启服务
func (c *CommandService) ForceRestartService(ctx context.Context, serviceId int64) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 获取服务信息
	service, err := c.serviceService.GetServiceById(ctx, serviceId)
	if err != nil {
		return "", err
	}

	if service.CmdStart == "" {
		return "", common.NewBusinessError(common.ErrCodeInvalidParam, "启动命令未配置")
	}

	port := strconv.Itoa(int(service.Port))
	var stopOutput string

	// 如果服务正在运行，强制终止
	if isRunning, _ := utils.IsPortInUse(port); isRunning {
		stopOutput, err = utils.Kill(port)
		if err != nil {
			return stopOutput, common.WrapError(common.ErrCodeCommandFailed, "强制终止服务失败", err)
		}

		// 等待端口释放
		if err := utils.WaitForPortFree(port, 5*time.Second); err != nil {
			return stopOutput, common.WrapError(common.ErrCodeCommandFailed, "等待端口释放失败", err)
		}
	}

	// 启动服务
	startOutput, err := c.executeCommand(ctx, service.CmdStart, service.Dir)
	if err != nil {
		return startOutput, common.WrapError(common.ErrCodeCommandFailed, "启动服务失败", err)
	}

	// 等待服务启动完成
	if err := c.waitForServiceStart(port, 3*time.Second); err != nil {
		return startOutput, common.WrapError(common.ErrCodeCommandFailed, "服务启动超时", err)
	}

	if stopOutput != "" {
		return fmt.Sprintf("强制终止输出:\n%s\n启动输出:\n%s", stopOutput, startOutput), nil
	}
	return startOutput, nil
}

// KillService 强制终止服务
func (c *CommandService) KillService(ctx context.Context, serviceId int64) (string, error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 获取服务信息
	service, err := c.serviceService.GetServiceById(ctx, serviceId)
	if err != nil {
		return "", err
	}

	port := strconv.Itoa(int(service.Port))
	if isRunning, _ := utils.IsPortInUse(port); !isRunning {
		return "", common.NewBusinessError(common.ErrCodeServiceStopped, "服务未运行")
	}

	// 强制终止进程
	output, err := utils.Kill(port)
	if err != nil {
		return output, common.WrapError(common.ErrCodeCommandFailed, "强制终止服务失败", err)
	}

	return output, nil
}

// BatchOperation 批量操作服务 - 优化版本
func (c *CommandService) BatchOperation(ctx context.Context, serviceIds []int64, operation string) []map[string]interface{} {
	results := make([]map[string]interface{}, len(serviceIds))

	// 限制并发数量，避免系统过载
	maxConcurrency := 5
	if len(serviceIds) < maxConcurrency {
		maxConcurrency = len(serviceIds)
	}

	semaphore := make(chan struct{}, maxConcurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i, serviceId := range serviceIds {
		wg.Add(1)
		go func(index int, id int64) {
			defer wg.Done()

			// 获取信号量
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			result := map[string]interface{}{
				"service_id": id,
				"success":    false,
				"message":    "",
				"output":     "",
				"duration":   0,
			}

			startTime := time.Now()
			var output string
			var err error

			// 为每个操作创建带超时的上下文
			opCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
			defer cancel()

			switch operation {
			case "start":
				output, err = c.StartService(opCtx, id)
			case "stop":
				output, err = c.StopService(opCtx, id)
			case "restart":
				output, err = c.RestartService(opCtx, id)
			case "kill":
				output, err = c.KillService(opCtx, id)
			default:
				err = fmt.Errorf("不支持的操作: %s", operation)
			}

			duration := time.Since(startTime)
			result["duration"] = duration.Milliseconds()

			if err != nil {
				if bizErr, ok := err.(*common.BusinessError); ok {
					result["message"] = bizErr.Message
				} else {
					result["message"] = err.Error()
				}
			} else {
				result["success"] = true
				result["message"] = "操作成功"
				result["output"] = output
			}

			mu.Lock()
			results[index] = result
			mu.Unlock()
		}(i, serviceId)
	}

	wg.Wait()
	return results
}

// executeCommand 执行命令 - 安全优化版本
func (c *CommandService) executeCommand(ctx context.Context, command, workDir string) (string, error) {
	if command == "" {
		return "", fmt.Errorf("命令不能为空")
	}

	// 验证命令安全性
	if err := utils.ValidateCommand(command); err != nil {
		return "", fmt.Errorf("命令安全验证失败: %v", err)
	}

	// 验证工作目录
	if workDir != "" {
		if err := utils.IsValidDirectory(workDir); err != nil {
			return "", fmt.Errorf("工作目录验证失败: %v", err)
		}
	}

	// 清理命令
	command = utils.SanitizeCommand(command)

	// 创建带超时的上下文
	cmdCtx, cancel := context.WithTimeout(ctx, c.commandTimeout)
	defer cancel()

	var cmd *exec.Cmd

	// 根据命令类型选择执行方式
	if strings.Contains(command, "&&") || strings.Contains(command, "||") || strings.Contains(command, ";") {
		// 复杂命令，使用bash执行
		cmd = exec.CommandContext(cmdCtx, "bash", "-c", command)
	} else {
		// 简单命令，直接分割参数
		args := strings.Fields(command)
		if len(args) == 0 {
			return "", fmt.Errorf("无效的命令")
		}
		cmd = exec.CommandContext(cmdCtx, args[0], args[1:]...)
	}

	// 设置工作目录
	if workDir != "" {
		cmd.Dir = workDir
	}

	// 设置子进程独立于父进程
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true, // 创建新进程组
	}

	// 设置环境变量限制
	cmd.Env = append(os.Environ(),
		"PATH=/usr/local/bin:/usr/bin:/bin", // 限制PATH
		"SHELL=/bin/bash",                   // 固定shell
	)

	// 执行命令
	output, err := cmd.CombinedOutput()
	if err != nil {
		// 检查是否是超时错误
		if cmdCtx.Err() == context.DeadlineExceeded {
			return string(output), fmt.Errorf("命令执行超时: %v", err)
		}
		return string(output), fmt.Errorf("命令执行失败: %v", err)
	}

	return string(output), nil
}

// waitForServiceStart 等待服务启动
func (c *CommandService) waitForServiceStart(port string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		if isRunning, _ := utils.IsPortInUse(port); isRunning {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("等待服务启动超时")
}

// waitForServiceStop 等待服务停止
func (c *CommandService) waitForServiceStop(port string, timeout time.Duration) error {
	start := time.Now()
	for time.Since(start) < timeout {
		if isRunning, _ := utils.IsPortInUse(port); !isRunning {
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("等待服务停止超时")
}
