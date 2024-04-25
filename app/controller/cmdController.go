package controller

import (
	"go_service/app/logic"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CmdController struct {
	logic *logic.Logic
}

func NewCmdController() *CmdController {
	return &CmdController{
		logic: logic.NewLogic(),
	}
}

func (s CmdController) Start(c *gin.Context) {
	id := c.Param("id")
	i, _ := strconv.Atoi(id)
	out, err := s.logic.Start(c, int64(i))
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "启动失败:" + err.Error(),
			"data": gin.H{},
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{"result": out},
	})
}

func (s CmdController) Stop(c *gin.Context) {
	id := c.Param("id")
	i, _ := strconv.Atoi(id)
	out, err := s.logic.Stop(c, int64(i))
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "启动失败:" + err.Error(),
			"data": gin.H{},
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{"result": out},
	})
}

func (s CmdController) Restart(c *gin.Context) {
	id := c.Param("id")
	i, _ := strconv.Atoi(id)
	out, err := s.logic.Restart(c, int64(i))
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "启动失败:" + err.Error(),
			"data": gin.H{},
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{"result": out},
	})
}

func (s CmdController) ForcedRestart(c *gin.Context) {
	id := c.Param("id")
	i, _ := strconv.Atoi(id)
	out, err := s.logic.ForcedRestart(c, int64(i))
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "启动失败:" + err.Error(),
			"data": gin.H{},
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{"result": out},
	})
}

func (s CmdController) Kill(c *gin.Context) {
	id := c.Param("id")
	i, _ := strconv.Atoi(id)
	out, err := s.logic.Kill(c, int64(i))
	if err != nil {
		c.JSON(200, gin.H{
			"code": 1,
			"msg":  "启动失败:" + err.Error(),
			"data": gin.H{},
		})
		return
	}
	c.JSON(200, gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{"result": out},
	})
}
