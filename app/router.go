package app

import (
	"go_service/app/config"
	"go_service/app/controller"
	"go_service/app/middleware"

	"github.com/gin-gonic/gin"
)

func RunHttp() {
	config.InitConfig()
	config.InitDatabase()

	r := gin.Default()
	// 捕获异常
	r.Use(middleware.ExceptErr())
	//增加拦截器
	r.Use(middleware.HttpInterceptor())
	//解决跨域
	r.Use(middleware.Cors())
	// 服务上下文
	r.Use(middleware.ServerContextHandler())

	//静态文件
	r.Static("/static", "./static")
	//模板文件
	r.LoadHTMLGlob("template/*")
	r.GET("/", controller.NewServiceController().Index)

	//路由组
	serviceGroup := r.Group("/service")
	{
		serviceGroup.POST("/add", controller.NewServiceController().Add)
		serviceGroup.GET("/findById/:id", controller.NewServiceController().FindById)
		serviceGroup.GET("/findByName/:key", controller.NewServiceController().FindByName)
		serviceGroup.POST("/delete/:id", controller.NewServiceController().DeleteById)
		serviceGroup.GET("/all", controller.NewServiceController().FindAll)
		serviceGroup.POST("/update", controller.NewServiceController().Update)
	}

	cmdGroup := r.Group("/cmd")
	{
		cmdGroup.POST("/start/:id", controller.NewCmdController().Start)
		cmdGroup.POST("/stop/:id", controller.NewCmdController().Stop)
		cmdGroup.POST("/restart/:id", controller.NewCmdController().Restart)
		cmdGroup.POST("/forcedRestart/:id", controller.NewCmdController().ForcedRestart)
		cmdGroup.POST("/kill/:id", controller.NewCmdController().Kill)
	}

	r.Run("127.0.0.1:" + config.Config.Port)
}
