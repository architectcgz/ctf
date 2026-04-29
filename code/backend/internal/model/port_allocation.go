package model

import "time"

// PortAllocation 记录实例运行时占用的宿主机端口。
type PortAllocation struct {
	Port       int    `gorm:"primaryKey"`
	InstanceID *int64 `gorm:"column:instance_id;index"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
