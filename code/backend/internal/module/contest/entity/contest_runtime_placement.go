package entity

import "time"

const (
	ContestRuntimePlacementStatusActive   = "active"
	ContestRuntimePlacementStatusReleased = "released"
)

type ContestRuntimePlacement struct {
	ID            int64  `gorm:"primaryKey"`
	ContestID     int64  `gorm:"column:contest_id;not null;index"`
	RuntimeNodeID int64  `gorm:"column:runtime_node_id;not null;index"`
	Status        string `gorm:"size:16;not null;default:'active'"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ReleasedAt    *time.Time `gorm:"column:released_at"`
}

func (ContestRuntimePlacement) TableName() string {
	return "contest_runtime_placements"
}
