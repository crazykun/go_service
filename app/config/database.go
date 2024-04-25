package config

import (
	"fmt"

	"gorm.io/gorm"
)

var DatabaseClient map[string]*gorm.DB

func InitDatabase() error {
	// 初始化数据库
	databaseConfig := Config.Database
	if len(databaseConfig) <= 0 {
		return nil
	}

	DatabaseClient = make(map[string]*gorm.DB, len(databaseConfig))
	for k, v := range databaseConfig {
		// 初始化数据库客户端
		db, err := InitMysqlClient(v.Host, v.Port, v.User, v.Pwd, v.Name)
		if err != nil {
			panic(fmt.Errorf("init mysql" + k + " client failed:" + err.Error()))
		}
		DatabaseClient[k] = db
		sqlDB, err := db.DB()
		if err != nil {
			panic("get mysql" + k + " sqlDB failed:" + err.Error())
		}
		// 设置连接池
		if v.MaxIdle > 0 {
			sqlDB.SetMaxIdleConns(v.MaxIdle)
		}
		if v.MaxOpen > 0 {
			sqlDB.SetMaxOpenConns(v.MaxOpen)
		}
	}
	return nil
}
