package model

import "time"

const (
	SubmissionReviewStatusNotRequired = "not_required"
	SubmissionReviewStatusPending     = "pending"
	SubmissionReviewStatusApproved    = "approved"
	SubmissionReviewStatusRejected    = "rejected"
)

type Submission struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	UserID        int64      `gorm:"column:user_id;not null;index:idx_user_challenge"`
	ChallengeID   int64      `gorm:"column:challenge_id;not null;index:idx_user_challenge"`
	ContestID     *int64     `gorm:"column:contest_id;index"`
	TeamID        *int64     `gorm:"column:team_id"`
	Flag          string     `gorm:"column:flag;size:500"`
	IsCorrect     bool       `gorm:"column:is_correct;not null"`
	ReviewStatus  string     `gorm:"column:review_status;size:32;default:'not_required';index:idx_submissions_review_status"`
	ReviewedBy    *int64     `gorm:"column:reviewed_by"`
	ReviewedAt    *time.Time `gorm:"column:reviewed_at"`
	ReviewComment string     `gorm:"column:review_comment"`
	Score         int        `gorm:"column:score;default:0"`
	SubmittedAt   time.Time  `gorm:"column:submitted_at;not null;index"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null"`
}

func (Submission) TableName() string {
	return "submissions"
}
