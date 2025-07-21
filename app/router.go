package app

import (
	"fmt"
	"go_service/app/config"
	"go_service/app/controller"
	"go_service/app/global"
	"go_service/app/middleware"
	"log"

	"github.com/gin-gonic/gin"
)

func RunHttp() {
	global.InitConfig()
	global.InitDatabase()

	// 设置Gin模式
	if config.GlobalConfig.App.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	// 创建性能监控中间件

	// 中间件
	r.Use(gin.Recovery())
	r.Use(middleware.ExceptErr())
	r.Use(middleware.HttpInterceptor())
	r.Use(middleware.Cors())
	r.Use(middleware.ServerContextHandler())
	r.Use(middleware.Logger())

	// 日志
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("[GIN] %s | %3d | %13v | %15s | %-7s  %s\n",
			param.TimeStamp.Format("2006-01-02 15:04:05"),
			param.StatusCode,
			param.Latency,
			param.ClientIP,
			param.Method,
			param.Path,
		)
	}))

	// 静态文件和模板
	r.Static("/static", "./static")
	r.LoadHTMLGlob("template/*")

	// 首页
	r.GET("/", controller.NewServiceController().Index)

	// API路由组
	api := r.Group("/api/v1")
	{
		// 服务管理
		services := api.Group("/service")
		{
			serviceController := controller.NewServiceController()
			services.POST("/add", serviceController.Add)
			services.GET("/findById/:id", serviceController.FindById)
			services.GET("/findByName/:key", serviceController.FindByName)
			services.POST("/delete/:id", serviceController.DeleteById)
			services.GET("/all", serviceController.FindAll)
			services.POST("/update", serviceController.Update)
		}

		// 服务操作
		cmd := api.Group("/cmd")
		{
			cmdController := controller.NewCmdController()
			cmd.POST("/start/:id", cmdController.Start)
			cmd.POST("/stop/:id", cmdController.Stop)
			cmd.POST("/restart/:id", cmdController.Restart)
			cmd.POST("/force-restart/:id", cmdController.ForcedRestart)
			cmd.POST("/kill/:id", cmdController.Kill)
		}

		// 批量操作
		batch := api.Group("/batch")
		{
			batchController := controller.NewBatchController()
			batch.POST("/operation", batchController.BatchOperation)
			batch.POST("/start-all", batchController.StartAll)
			batch.POST("/stop-all", batchController.StopAll)
		}

	}

	// 启动服务器
	addr := "127.0.0.1:" + config.GlobalConfig.Server.Port
	log.Printf("服务管理工具启动成功，访问地址: http://%s", addr)

	r.Run(addr)
}
