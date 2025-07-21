package model

import "time"

// ServiceLog 服务日志模型
type ServiceLog struct {
	Id        int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	ServiceId int64     `json:"service_id" gorm:"not null;index"`
	Operation string    `json:"operation" gorm:"type:varchar(50);not null"` // start, stop, restart, kill
	Status    string    `json:"status" gorm:"type:varchar(20);not null"`    // success, failed
	Output    string    `json:"output" gorm:"type:text"`
	Error     string    `json:"error" gorm:"type:text"`
	Duration  int64     `json:"duration" gorm:"default:0"` // 执行时长(毫秒)
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

func (s ServiceLog) TableName() string {
	return "service_log"
}
