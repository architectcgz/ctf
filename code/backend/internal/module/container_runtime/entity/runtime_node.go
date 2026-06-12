package entity

import "time"

const (
	RuntimeNodeHealthUnknown  = "unknown"
	RuntimeNodeHealthReady    = "ready"
	RuntimeNodeHealthDegraded = "degraded"
	RuntimeNodeHealthOffline  = "offline"
)

type RuntimeNode struct {
	ID               int64      `gorm:"column:id;primaryKey"`
	Name             string     `gorm:"column:name;size:128;not null;uniqueIndex"`
	Endpoint         string     `gorm:"column:endpoint;size:255;not null;default:''"`
	TLSIdentity      string     `gorm:"column:tls_identity;size:255;not null;default:''"`
	Schedulable      bool       `gorm:"column:schedulable;not null;default:true"`
	Labels           string     `gorm:"column:labels;type:jsonb;not null;default:'{}'"`
	HealthStatus     string     `gorm:"column:health_status;size:32;not null;default:'unknown'"`
	CapacitySnapshot string     `gorm:"column:capacity_snapshot;type:jsonb;not null;default:'{}'"`
	LastSeenAt       *time.Time `gorm:"column:last_seen_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (RuntimeNode) TableName() string {
	return "runtime_nodes"
}
