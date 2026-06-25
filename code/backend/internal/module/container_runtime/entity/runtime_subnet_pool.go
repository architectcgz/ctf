package entity

import "time"

const (
	RuntimeSubnetPoolKindSingleContainer = "single_container"
	RuntimeSubnetPoolKindTopology        = "topology"
)

type RuntimeSubnetPool struct {
	RuntimeNodeID int64      `gorm:"column:runtime_node_id;primaryKey"`
	PoolKind      string     `gorm:"column:pool_kind;size:32;not null;index"`
	Subnet        string     `gorm:"column:subnet;type:text;primaryKey"`
	Status        string     `gorm:"column:status;size:16;not null;default:'available';index"`
	InstanceID    *int64     `gorm:"column:instance_id;index"`
	NetworkKey    string     `gorm:"column:network_key;size:128;not null;default:''"`
	ReservedAt    *time.Time `gorm:"column:reserved_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (RuntimeSubnetPool) TableName() string {
	return "runtime_subnet_pool"
}
