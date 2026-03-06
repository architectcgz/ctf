package model

import "time"

type Submission struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	UserID      int64     `gorm:"column:user_id;not null;index:idx_user_challenge"`
	ChallengeID int64     `gorm:"column:challenge_id;not null;index:idx_user_challenge"`
	Flag        string    `gorm:"column:flag;size:500"`
	IsCorrect   bool      `gorm:"column:is_correct;not null"`
	SubmittedAt time.Time `gorm:"column:submitted_at;not null;index"`
}

func (Submission) TableName() string {
	return "submissions"
}
