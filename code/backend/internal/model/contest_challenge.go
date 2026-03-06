package model

import "time"

type ContestChallenge struct {
	ID           int64     `gorm:"column:id;primaryKey"`
	ContestID    int64     `gorm:"column:contest_id;not null;index"`
	ChallengeID  int64     `gorm:"column:challenge_id;not null"`
	ContestScore *int      `gorm:"column:contest_score"`
	FirstBloodBy *int64    `gorm:"column:first_blood_by"`
	CreatedAt    time.Time `gorm:"column:created_at;not null"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null"`
}

func (ContestChallenge) TableName() string {
	return "contest_challenges"
}
