package entity

import "time"

const (
	RuntimeResourceStatusAvailable   = "available"
	RuntimeResourceStatusReserved    = "reserved"
	RuntimeResourceStatusBound       = "bound"
	RuntimeResourceStatusQuarantined = "quarantined"
)

type RuntimePortPool struct {
	RuntimeNodeID int64      `gorm:"column:runtime_node_id;primaryKey"`
	Port          int        `gorm:"column:port;primaryKey"`
	Status        string     `gorm:"column:status;size:16;not null;default:'available';index"`
	InstanceID    *int64     `gorm:"column:instance_id;index"`
	ReservedAt    *time.Time `gorm:"column:reserved_at"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (RuntimePortPool) TableName() string {
	return "runtime_port_pool"
}
