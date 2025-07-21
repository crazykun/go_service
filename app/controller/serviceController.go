package controller

import (
	"go_service/app/common"
	"go_service/app/global"
	"go_service/app/model"
	"go_service/app/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ServiceController struct {
	serviceService *service.ServiceService
}

func NewServiceController() *ServiceController {
	return &ServiceController{
		serviceService: service.NewServiceService(global.GetDefaultDb()),
	}
}

func (s *ServiceController) Index(c *gin.Context) {
	c.HTML(200, "index.html", gin.H{})
}

func (s *ServiceController) Add(c *gin.Context) {
	var serviceModel model.ServiceModel
	if err := c.ShouldBindJSON(&serviceModel); err != nil {
		common.Error(c, "参数异常: "+err.Error())
		return
	}
	
	if err := s.serviceService.CreateService(c.Request.Context(), &serviceModel); err != nil {
		common.HandleBusinessError(c, err)
		return
	}

	common.Success(c, gin.H{"id": serviceModel.Id})
}

func (s *ServiceController) FindById(c *gin.Context) {
	id := c.Param("id")
	serviceId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		common.Error(c, "无效的ID参数")
		return
	}
	
	serviceModel, err := s.serviceService.GetServiceById(c.Request.Context(), serviceId)
	if err != nil {
		common.HandleBusinessError(c, err)
		return
	}
	
	common.Success(c, serviceModel)
}

func (s *ServiceController) FindByName(c *gin.Context) {
	name := c.Param("key")
	if name == "" {
		common.Error(c, "服务名称不能为空")
		return
	}
	
	serviceModel, err := s.serviceService.GetServiceByName(c.Request.Context(), name)
	if err != nil {
		common.HandleBusinessError(c, err)
		return
	}
	
	common.Success(c, serviceModel)
}

func (s *ServiceController) FindAll(c *gin.Context) {
	// 支持分页查询
	var req model.ServiceListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// 如果绑定失败，使用默认参数
		req = model.ServiceListRequest{Page: 1, PageSize: 100}
	}

	// 如果没有分页参数，获取所有服务
	if req.Page == 0 && req.PageSize == 0 {
		services, err := s.serviceService.GetAllServicesWithStatus(c.Request.Context())
		if err != nil {
			common.HandleBusinessError(c, err)
			return
		}
		common.Success(c, services)
		return
	}

	// 分页查询
	response, err := s.serviceService.ListServices(c.Request.Context(), &req)
	if err != nil {
		common.HandleBusinessError(c, err)
		return
	}

	common.Paginate(c, response.List, response.Total, response.Page, response.Size)
}

func (s *ServiceController) DeleteById(c *gin.Context) {
	id := c.Param("id")
	serviceId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		common.Error(c, "无效的ID参数")
		return
	}
	
	if err := s.serviceService.DeleteService(c.Request.Context(), serviceId); err != nil {
		common.HandleBusinessError(c, err)
		return
	}
	
	common.Success(c, gin.H{"deleted": true})
}

func (s *ServiceController) Update(c *gin.Context) {
	var serviceModel model.ServiceModel
	if err := c.ShouldBindJSON(&serviceModel); err != nil {
		common.Error(c, "参数异常: "+err.Error())
		return
	}
	
	if serviceModel.Id <= 0 {
		common.Error(c, "无效的服务ID")
		return
	}
	
	if err := s.serviceService.UpdateService(c.Request.Context(), &serviceModel); err != nil {
		common.HandleBusinessError(c, err)
		return
	}
	
	common.Success(c, gin.H{"id": serviceModel.Id})
}
