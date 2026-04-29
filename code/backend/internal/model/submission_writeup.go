package model

import "time"

const (
	SubmissionWriteupStatusDraft     = "draft"
	SubmissionWriteupStatusPublished = "published"

	SubmissionWriteupVisibilityVisible = "visible"
	SubmissionWriteupVisibilityHidden  = "hidden"

	// Deprecated aliases kept temporarily to reduce cross-package churn during migration.
	SubmissionWriteupStatusSubmitted     = SubmissionWriteupStatusPublished
	SubmissionWriteupReviewPending       = SubmissionWriteupVisibilityVisible
	SubmissionWriteupReviewReviewed      = SubmissionWriteupVisibilityVisible
	SubmissionWriteupReviewExcellent     = SubmissionWriteupVisibilityVisible
	SubmissionWriteupReviewNeedsRevision = SubmissionWriteupVisibilityHidden
)

type SubmissionWriteup struct {
	ID               int64      `gorm:"column:id;primaryKey"`
	UserID           int64      `gorm:"column:user_id;uniqueIndex:uk_submission_writeups_user_challenge"`
	ChallengeID      int64      `gorm:"column:challenge_id;uniqueIndex:uk_submission_writeups_user_challenge;index:idx_submission_writeups_challenge"`
	ContestID        *int64     `gorm:"column:contest_id"`
	Title            string     `gorm:"column:title"`
	Content          string     `gorm:"column:content"`
	SubmissionStatus string     `gorm:"column:submission_status"`
	VisibilityStatus string     `gorm:"column:visibility_status;index:idx_submission_writeups_visibility_status"`
	IsRecommended    bool       `gorm:"column:is_recommended;index:idx_submission_writeups_recommended"`
	RecommendedAt    *time.Time `gorm:"column:recommended_at"`
	RecommendedBy    *int64     `gorm:"column:recommended_by"`
	PublishedAt      *time.Time `gorm:"column:published_at"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at"`
}

func (SubmissionWriteup) TableName() string {
	return "submission_writeups"
}
