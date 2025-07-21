package controller

import (
	"go_service/app/common"
	"go_service/app/global"
	"go_service/app/service"

	"github.com/gin-gonic/gin"
)

type BatchController struct {
	commandService *service.CommandService
	serviceService *service.ServiceService
}

func NewBatchController() *BatchController {
	db := global.GetDefaultDb()
	return &BatchController{
		commandService: service.NewCommandService(db),
		serviceService: service.NewServiceService(db),
	}
}

// BatchOperationRequest 批量操作请求
type BatchOperationRequest struct {
	ServiceIds []int64 `json:"service_ids" binding:"required"`
	Operation  string  `json:"operation" binding:"required,oneof=start stop restart force_restart kill"`
}

// BatchOperation 批量操作服务
func (b *BatchController) BatchOperation(c *gin.Context) {
	var req BatchOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Error(c, "参数错误: "+err.Error())
		return
	}

	if len(req.ServiceIds) == 0 {
		common.Error(c, "服务ID列表不能为空")
		return
	}

	// 使用服务层的批量操作方法
	results := b.commandService.BatchOperation(c.Request.Context(), req.ServiceIds, req.Operation)

	// 统计成功数量
	successCount := 0
	for _, result := range results {
		if success, ok := result["success"].(bool); ok && success {
			successCount++
		}
	}

	common.Success(c, gin.H{
		"operation":     req.Operation,
		"total":         len(req.ServiceIds),
		"success_count": successCount,
		"results":       results,
	})
}

// StartAll 启动所有服务
func (b *BatchController) StartAll(c *gin.Context) {
	// 获取所有服务及其状态
	services, err := b.serviceService.GetAllServicesWithStatus(c.Request.Context())
	if err != nil {
		common.HandleBusinessError(c, err)
		return
	}

	if len(services) == 0 {
		common.Success(c, gin.H{
			"total":         0,
			"success_count": 0,
			"results":       []interface{}{},
			"message":       "没有可启动的服务",
		})
		return
	}

	// 筛选出需要启动的服务ID
	var serviceIds []int64
	for _, service := range services {
		if service.Status == 0 { // 只启动停止状态的服务
			serviceIds = append(serviceIds, service.Id)
		}
	}

	if len(serviceIds) == 0 {
		common.Success(c, gin.H{
			"total":         len(services),
			"success_count": len(services),
			"results":       []interface{}{},
			"message":       "所有服务都已在运行",
		})
		return
	}

	// 批量启动服务
	results := b.commandService.BatchOperation(c.Request.Context(), serviceIds, "start")

	// 为已运行的服务添加跳过记录
	for _, service := range services {
		if service.Status == 1 {
			results = append(results, map[string]interface{}{
				"service_id":   service.Id,
				"service_name": service.Name,
				"success":      true,
				"message":      "服务已在运行",
				"skipped":      true,
			})
		}
	}

	// 统计成功数量
	successCount := 0
	for _, result := range results {
		if success, ok := result["success"].(bool); ok && success {
			successCount++
		}
	}

	common.Success(c, gin.H{
		"operation":     "start_all",
		"total":         len(services),
		"success_count": successCount,
		"results":       results,
	})
}

// StopAll 停止所有服务
func (b *BatchController) StopAll(c *gin.Context) {
	// 获取所有服务及其状态
	services, err := b.serviceService.GetAllServicesWithStatus(c.Request.Context())
	if err != nil {
		common.HandleBusinessError(c, err)
		return
	}

	if len(services) == 0 {
		common.Success(c, gin.H{
			"total":         0,
			"success_count": 0,
			"results":       []interface{}{},
			"message":       "没有可停止的服务",
		})
		return
	}

	// 筛选出需要停止的服务ID
	var serviceIds []int64
	for _, service := range services {
		if service.Status == 1 { // 只停止运行状态的服务
			serviceIds = append(serviceIds, service.Id)
		}
	}

	if len(serviceIds) == 0 {
		common.Success(c, gin.H{
			"total":         len(services),
			"success_count": len(services),
			"results":       []interface{}{},
			"message":       "所有服务都已停止",
		})
		return
	}

	// 批量停止服务
	results := b.commandService.BatchOperation(c.Request.Context(), serviceIds, "stop")

	// 为已停止的服务添加跳过记录
	for _, service := range services {
		if service.Status == 0 {
			results = append(results, map[string]interface{}{
				"service_id":   service.Id,
				"service_name": service.Name,
				"success":      true,
				"message":      "服务已停止",
				"skipped":      true,
			})
		}
	}

	// 统计成功数量
	successCount := 0
	for _, result := range results {
		if success, ok := result["success"].(bool); ok && success {
			successCount++
		}
	}

	common.Success(c, gin.H{
		"operation":     "stop_all",
		"total":         len(services),
		"success_count": successCount,
		"results":       results,
	})
}

// RestartAll 重启所有运行中的服务
func (b *BatchController) RestartAll(c *gin.Context) {
	// 获取所有服务及其状态
	services, err := b.serviceService.GetAllServicesWithStatus(c.Request.Context())
	if err != nil {
		common.HandleBusinessError(c, err)
		return
	}

	if len(services) == 0 {
		common.Success(c, gin.H{
			"total":         0,
			"success_count": 0,
			"results":       []interface{}{},
			"message":       "没有可重启的服务",
		})
		return
	}

	// 筛选出需要重启的服务ID（只重启运行中的服务）
	var serviceIds []int64
	for _, service := range services {
		if service.Status == 1 {
			serviceIds = append(serviceIds, service.Id)
		}
	}

	if len(serviceIds) == 0 {
		common.Success(c, gin.H{
			"total":         len(services),
			"success_count": 0,
			"results":       []interface{}{},
			"message":       "没有运行中的服务需要重启",
		})
		return
	}

	// 批量重启服务
	results := b.commandService.BatchOperation(c.Request.Context(), serviceIds, "restart")

	// 统计成功数量
	successCount := 0
	for _, result := range results {
		if success, ok := result["success"].(bool); ok && success {
			successCount++
		}
	}

	common.Success(c, gin.H{
		"operation":     "restart_all",
		"total":         len(serviceIds),
		"success_count": successCount,
		"results":       results,
	})
}