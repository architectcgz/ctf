package model

import "time"

type Contest struct {
	ID          int64     `gorm:"primaryKey"`
	Title       string    `gorm:"not null"`
	Description string    `gorm:"type:text"`
	StartTime   time.Time `gorm:"not null"`
	EndTime     time.Time `gorm:"not null"`
	FreezeTime  *time.Time
	Status      string    `gorm:"not null;default:'draft'"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

const (
	ContestStatusDraft     = "draft"
	ContestStatusRunning   = "running"
	ContestStatusFinished  = "finished"
)
