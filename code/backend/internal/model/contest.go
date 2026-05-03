package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ContestModeJeopardy = "jeopardy"
	ContestModeAWD      = "awd"

	ContestStatusDraft        = "draft"
	ContestStatusRegistration = "registration"
	ContestStatusRunning      = "running"
	ContestStatusFrozen       = "frozen"
	ContestStatusEnded        = "ended"
)

type Contest struct {
	ID            int64          `gorm:"column:id;primaryKey"`
	Title         string         `gorm:"column:title"`
	Description   string         `gorm:"column:description;type:text"`
	Mode          string         `gorm:"column:mode"`
	StartTime     time.Time      `gorm:"column:start_time"`
	EndTime       time.Time      `gorm:"column:end_time"`
	FreezeTime    *time.Time     `gorm:"column:freeze_time"`
	Status        string         `gorm:"column:status"`
	StatusVersion int64          `gorm:"column:status_version"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Contest) TableName() string {
	return "contests"
}
