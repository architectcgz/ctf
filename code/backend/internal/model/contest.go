package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ContestStatusPending = "pending"
	ContestStatusRunning = "running"
	ContestStatusEnded   = "ended"
)

type Contest struct {
	ID        int64          `gorm:"column:id;primaryKey"`
	Name      string         `gorm:"column:name"`
	Status    string         `gorm:"column:status"`
	StartTime time.Time      `gorm:"column:start_time"`
	EndTime   time.Time      `gorm:"column:end_time"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Contest) TableName() string {
	return "contests"
}
