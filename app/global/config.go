package global

import (
	"flag"
	"fmt"
	"os"

	"go_service/app/config"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var RunMode string
var AppPath string
var Config config.Config

// 从config目录中读取yaml环境配置
func InitConfig() {
	AppPath, _ = os.Getwd()

	// 从启动命令里面读取-c参数指定的配置文件
	// go run main.go -c config.yml
	flag.StringVar(&RunMode, "c", "config.yml", "specify config file")
	flag.Parse()
	// 读取配置文件
	if RunMode == "" {
		RunMode = "config.yml"
	}
	viper.SetConfigName(RunMode)       //配置文件名
	viper.SetConfigType("yml")         //配置文件类型
	viper.AddConfigPath(AppPath + "/") //执行go run对应的路径配置
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("read config file error: %s", err.Error()))
	}
	// 将配置信息反序列化到结构体中
	if err := viper.Unmarshal(&Config); err != nil {
		panic(fmt.Errorf("unmarshal config error: %s", err.Error()))
	}
	fmt.Printf("dir:%s\nmode:%s\nconf:%+v\n", AppPath, RunMode, Config)
	// 注册每次配置文件发生变更后都会调用的回调函数
	viper.OnConfigChange(func(e fsnotify.Event) {
		// 每次配置文件发生变化，需要重新将其反序列化到结构体中
		if err := viper.Unmarshal(&Config); err != nil {
			panic(fmt.Errorf("unmarshal config error: %s", err.Error()))
		}
		fmt.Printf("Config file update:%s Op:%s\n", e.Name, e.Op)
	})

	// 监控配置文件变化
	viper.WatchConfig()

}
