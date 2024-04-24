package config

import (
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type Mysql struct {
}

func InitMysqlClient(host string, port int, user, pwd, name string) (*gorm.DB, error) {
	// 初始化mysql客户端
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local", user, pwd, host, port, name)
	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN: dsn,
	}), &gorm.Config{})

	if err != nil {
		return nil, err
	}
	return db, nil
}

func (m *Mysql) GetMysqlClient(name string) *gorm.DB {
	// 获取mysql客户端
	return DatabaseClient[name]
}

func (m *Mysql) GetMysqlClientByDefault() *gorm.DB {
	// 获取默认mysql客户端
	return DatabaseClient["default"]
}

func GetDefaultDb() *gorm.DB {
	// 获取默认数据库
	db := DatabaseClient["default"]
	return db
}
