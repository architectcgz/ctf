package model

import (
	"time"

	"gorm.io/gorm"
)

type ContestChallenge struct {
	ID          int64          `gorm:"column:id;primaryKey"`
	ContestID   int64          `gorm:"column:contest_id;index:idx_contest_challenge,unique"`
	ChallengeID int64          `gorm:"column:challenge_id;index:idx_contest_challenge,unique"`
	Points      int            `gorm:"column:points"`
	Order       int            `gorm:"column:order"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (ContestChallenge) TableName() string {
	return "contest_challenges"
}
