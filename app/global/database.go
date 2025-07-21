package global

import (
	"fmt"
	"time"

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
		db, err := InitMysqlClient(v.Host, v.Port, v.User, v.Password, v.Name)
		if err != nil {
			panic(fmt.Errorf("init mysql" + k + " client failed:" + err.Error()))
		}
		DatabaseClient[k] = db
		sqlDB, err := db.DB()
		if err != nil {
			panic("get mysql" + k + " sqlDB failed:" + err.Error())
		}

		// 优化连接池配置
		if v.MaxIdle > 0 {
			sqlDB.SetMaxIdleConns(v.MaxIdle)
		} else {
			sqlDB.SetMaxIdleConns(10) // 默认最大空闲连接数
		}

		if v.MaxOpen > 0 {
			sqlDB.SetMaxOpenConns(v.MaxOpen)
		} else {
			sqlDB.SetMaxOpenConns(100) // 默认最大连接数
		}

		// 设置连接生存时间，避免长时间空闲连接被服务器关闭
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetConnMaxIdleTime(10 * time.Minute)

		// 测试连接
		if err := sqlDB.Ping(); err != nil {
			panic(fmt.Errorf("database %s ping failed: %v", k, err))
		}

		fmt.Printf("数据库 %s 连接成功\n", k)
	}
	return nil
}
