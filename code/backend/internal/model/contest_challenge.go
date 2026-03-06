package model

import (
	"time"

	"gorm.io/gorm"
)

type ContestChallenge struct {
	ID          int64          `gorm:"column:id;primaryKey"`
	ContestID   int64          `gorm:"column:contest_id"`
	ChallengeID int64          `gorm:"column:challenge_id"`
	Points      int            `gorm:"column:points"`
	Order       int            `gorm:"column:order"`
	IsVisible   bool           `gorm:"column:is_visible"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (ContestChallenge) TableName() string {
	return "contest_challenges"
}
