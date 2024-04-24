package controller

import (
	"go_service/app/logic"
	"strconv"

	"go_service/app/model"

	"github.com/gin-gonic/gin"
)

type ServiceController struct {
	logic *logic.ServiceLogic
}

func NewServiceController() *ServiceController {
	return &ServiceController{
		logic: logic.NewServiceLogic(),
	}
}

func (s ServiceController) Index(c *gin.Context) {
	// 渲染模板
	c.HTML(200, "index.html", gin.H{})
}

func (s ServiceController) Add(c *gin.Context) {
	info := model.ServiceModel{}
	err := c.ShouldBindJSON(&info)
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "参数异常:" + err.Error(),
			"data": gin.H{},
		})
		return
	}
	id, err := s.logic.Add(c, info)
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "添加失败:" + err.Error(),
			"data": gin.H{},
		})
		return
	}

	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"id": id,
		},
	})
}

func (s ServiceController) FindById(c *gin.Context) {
	id := c.Param("id")
	i, _ := strconv.Atoi(id)
	info := s.logic.GetById(c, int64(i))
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": info,
	})
}

func (s ServiceController) FindByName(c *gin.Context) {
	name := c.Param("name")
	info := s.logic.GetByName(c, name)
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": info,
	})
}

func (s ServiceController) FindAll(c *gin.Context) {
	infos, _ := s.logic.FindAll(c)
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": infos,
	})
}

func (s ServiceController) DeleteById(c *gin.Context) {
	id := c.Param("id")
	i, _ := strconv.Atoi(id)
	isOk := s.logic.DeleteById(c, int64(i))
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{"isOk": isOk},
	})
}

func (s ServiceController) Update(c *gin.Context) {
	info := model.ServiceModel{}
	err := c.ShouldBindJSON(&info)
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "参数异常:" + err.Error(),
			"data": gin.H{"id": info.Id},
		})
		return
	}
	s.logic.UpdateById(c, info)
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{"id": info.Id},
	})
}
