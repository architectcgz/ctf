package model

import "time"

type ChallengeHint struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	ChallengeID int64     `gorm:"column:challenge_id;index;uniqueIndex:uk_challenge_hint_level,priority:1"`
	Level       int       `gorm:"column:level;uniqueIndex:uk_challenge_hint_level,priority:2"`
	Title       string    `gorm:"column:title"`
	Content     string    `gorm:"column:content"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (ChallengeHint) TableName() string {
	return "challenge_hints"
}
