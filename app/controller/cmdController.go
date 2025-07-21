package controller

import (
	"go_service/app/common"
	"go_service/app/global"
	"go_service/app/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CmdController struct {
	commandService *service.CommandService
}

func NewCmdController() *CmdController {
	return &CmdController{
		commandService: service.NewCommandService(global.GetDefaultDb()),
	}
}

func (s *CmdController) Start(c *gin.Context) {
	id := c.Param("id")
	serviceId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		common.Error(c, "无效的ID参数")
		return
	}
	
	output, err := s.commandService.StartService(c.Request.Context(), serviceId)
	if err != nil {
		common.HandleBusinessError(c, err)
		return
	}
	
	common.Success(c, gin.H{
		"service_id": serviceId,
		"operation":  "start",
		"output":     output,
		"message":    "启动成功",
	})
}

func (s *CmdController) Stop(c *gin.Context) {
	id := c.Param("id")
	serviceId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		common.Error(c, "无效的ID参数")
		return
	}
	
	output, err := s.commandService.StopService(c.Request.Context(), serviceId)
	if err != nil {
		common.HandleBusinessError(c, err)
		return
	}
	
	common.Success(c, gin.H{
		"service_id": serviceId,
		"operation":  "stop",
		"output":     output,
		"message":    "停止成功",
	})
}

func (s *CmdController) Restart(c *gin.Context) {
	id := c.Param("id")
	serviceId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		common.Error(c, "无效的ID参数")
		return
	}
	
	output, err := s.commandService.RestartService(c.Request.Context(), serviceId)
	if err != nil {
		common.HandleBusinessError(c, err)
		return
	}
	
	common.Success(c, gin.H{
		"service_id": serviceId,
		"operation":  "restart",
		"output":     output,
		"message":    "重启成功",
	})
}

func (s *CmdController) ForcedRestart(c *gin.Context) {
	id := c.Param("id")
	serviceId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		common.Error(c, "无效的ID参数")
		return
	}
	
	output, err := s.commandService.ForceRestartService(c.Request.Context(), serviceId)
	if err != nil {
		common.HandleBusinessError(c, err)
		return
	}
	
	common.Success(c, gin.H{
		"service_id": serviceId,
		"operation":  "force_restart",
		"output":     output,
		"message":    "强制重启成功",
	})
}

func (s *CmdController) Kill(c *gin.Context) {
	id := c.Param("id")
	serviceId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		common.Error(c, "无效的ID参数")
		return
	}
	
	output, err := s.commandService.KillService(c.Request.Context(), serviceId)
	if err != nil {
		common.HandleBusinessError(c, err)
		return
	}
	
	common.Success(c, gin.H{
		"service_id": serviceId,
		"operation":  "kill",
		"output":     output,
		"message":    "强制终止成功",
	})
}
