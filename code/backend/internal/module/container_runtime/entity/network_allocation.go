package entity

import "time"

// NetworkAllocation 记录实例运行时占用的 Jeopardy 网络子网。
type NetworkAllocation struct {
	Subnet     string `gorm:"column:subnet;primaryKey"`
	InstanceID *int64 `gorm:"column:instance_id;index;uniqueIndex:uk_network_allocations_owner_key"`
	NetworkKey string `gorm:"column:network_key;size:128;not null;default:'';uniqueIndex:uk_network_allocations_owner_key"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (NetworkAllocation) TableName() string {
	return "network_allocations"
}
