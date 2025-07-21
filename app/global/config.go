package global

import (
	"flag"
	"os"

	"go_service/app/config"
)

var ConfigFile string
var AppPath string

// 从config目录中读取yaml环境配置
func InitConfig() {
	AppPath, _ = os.Getwd()

	// 从启动命令里面读取-c参数指定的配置文件
	// go run main.go -c config.yml
	flag.StringVar(&ConfigFile, "c", "config.yml", "specify config file")
	flag.Parse()
	config.InitConfig(ConfigFile)
}
