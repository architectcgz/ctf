package model

import "time"

const (
	ContestStatusDraft      = "draft"
	ContestStatusRunning    = "running"
	ContestStatusEnded      = "ended"
)

type Contest struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	Title     string    `gorm:"column:title;size:128;not null"`
	Status    string    `gorm:"column:status;size:16;not null;default:'draft'"`
	TeamMode  bool      `gorm:"column:team_mode;not null;default:false"`
	StartAt   time.Time `gorm:"column:start_at;not null"`
	EndAt     time.Time `gorm:"column:end_at;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (Contest) TableName() string {
	return "contests"
}
