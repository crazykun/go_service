package model

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

type ServiceModel struct {
	Id              int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Name            string    `json:"name" gorm:"type:varchar(100);not null;uniqueIndex" binding:"required"`
	Title           string    `json:"title" gorm:"type:varchar(200)"`
	Dir             string    `json:"dir" gorm:"type:varchar(500)" binding:"required"`
	CmdStart        string    `json:"cmd_start" gorm:"type:text;not null" binding:"required"`
	CmdStop         string    `json:"cmd_stop" gorm:"type:text"`
	CmdRestart      string    `json:"cmd_restart" gorm:"type:text"`
	Port            int64     `json:"port" gorm:"not null;uniqueIndex" binding:"required,min=1,max=65535"`
	HealthCheckUrl  string    `json:"health_check_url" gorm:"type:varchar(500)"` // 健康检查URL
	AutoRestart     bool      `json:"auto_restart" gorm:"default:false"`         // 是否自动重启
	MaxRestartCount int       `json:"max_restart_count" gorm:"default:3"`        // 最大重启次数
	RestartInterval int       `json:"restart_interval" gorm:"default:30"`        // 重启间隔(秒)
	Remark          string    `json:"remark" gorm:"type:text"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ServiceStatusModel struct {
	ServiceModel
	Status  int    `json:"status"`  // 0: 停止, 1: 运行中
	Pid     string `json:"pid"`     // 进程ID
	Process string `json:"process"` // 进程名称
}

func (s ServiceModel) TableName() string {
	return "service"
}

// Validate 验证服务模型数据
func (s *ServiceModel) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("服务名称不能为空")
	}
	if s.Dir == "" {
		return fmt.Errorf("工作目录不能为空")
	}
	if s.CmdStart == "" {
		return fmt.Errorf("启动命令不能为空")
	}
	if s.Port <= 0 || s.Port > 65535 {
		return fmt.Errorf("端口号必须在1-65535之间")
	}
	return nil
}

// 自动添加时间
func (s *ServiceModel) BeforeCreate(tx *gorm.DB) (err error) {
	s.CreatedAt = time.Now()
	s.UpdatedAt = time.Now()
	return
}
