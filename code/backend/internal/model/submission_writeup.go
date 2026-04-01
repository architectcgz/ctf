package model

import "time"

const (
	SubmissionWriteupStatusDraft     = "draft"
	SubmissionWriteupStatusSubmitted = "submitted"

	SubmissionWriteupReviewPending       = "pending"
	SubmissionWriteupReviewReviewed      = "reviewed"
	SubmissionWriteupReviewExcellent     = "excellent"
	SubmissionWriteupReviewNeedsRevision = "needs_revision"
)

type SubmissionWriteup struct {
	ID               int64      `gorm:"column:id;primaryKey"`
	UserID           int64      `gorm:"column:user_id;uniqueIndex:uk_submission_writeups_user_challenge"`
	ChallengeID      int64      `gorm:"column:challenge_id;uniqueIndex:uk_submission_writeups_user_challenge;index:idx_submission_writeups_challenge"`
	ContestID        *int64     `gorm:"column:contest_id"`
	Title            string     `gorm:"column:title"`
	Content          string     `gorm:"column:content"`
	SubmissionStatus string     `gorm:"column:submission_status"`
	ReviewStatus     string     `gorm:"column:review_status;index:idx_submission_writeups_review_status"`
	SubmittedAt      *time.Time `gorm:"column:submitted_at"`
	ReviewedBy       *int64     `gorm:"column:reviewed_by"`
	ReviewedAt       *time.Time `gorm:"column:reviewed_at"`
	ReviewComment    string     `gorm:"column:review_comment"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (SubmissionWriteup) TableName() string {
	return "submission_writeups"
}
