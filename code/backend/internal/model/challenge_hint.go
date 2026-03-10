package model

import "time"

type ChallengeHint struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	ChallengeID int64     `gorm:"column:challenge_id;index;uniqueIndex:uk_challenge_hint_level,priority:1"`
	Level       int       `gorm:"column:level;uniqueIndex:uk_challenge_hint_level,priority:2"`
	Title       string    `gorm:"column:title"`
	CostPoints  int       `gorm:"column:cost_points;default:0"`
	Content     string    `gorm:"column:content"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (ChallengeHint) TableName() string {
	return "challenge_hints"
}

type ChallengeHintUnlock struct {
	ID              int64     `gorm:"column:id;primaryKey"`
	UserID          int64     `gorm:"column:user_id;index;uniqueIndex:uk_user_hint_unlock,priority:1"`
	ChallengeID     int64     `gorm:"column:challenge_id;index"`
	ChallengeHintID int64     `gorm:"column:challenge_hint_id;index;uniqueIndex:uk_user_hint_unlock,priority:2"`
	UnlockedAt      time.Time `gorm:"column:unlocked_at"`
}

func (ChallengeHintUnlock) TableName() string {
	return "challenge_hint_unlocks"
}
