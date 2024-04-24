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
	r.LoadHTMLGlob("src/*")
	r.GET("/", controller.NewServiceController().Index)

	//路由组
	appInfoGroup := r.Group("/service")
	{
		appInfoGroup.POST("/add", controller.NewServiceController().Add)
		appInfoGroup.GET("/findById/:id", controller.NewServiceController().FindById)
		appInfoGroup.GET("/findByName/:key", controller.NewServiceController().FindByName)
		appInfoGroup.POST("/delete/:id", controller.NewServiceController().DeleteById)
		appInfoGroup.GET("/all", controller.NewServiceController().FindAll)
		appInfoGroup.POST("/update", controller.NewServiceController().Update)
	}
	r.Run("127.0.0.1:" + config.Config.Port)
}
