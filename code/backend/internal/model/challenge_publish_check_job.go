package model

import "time"

const (
	ChallengePublishCheckStatusPending = "pending"
	ChallengePublishCheckStatusRunning = "running"
	ChallengePublishCheckStatusPassed  = "passed"
	ChallengePublishCheckStatusFailed  = "failed"
)

type ChallengePublishCheckJob struct {
	ID             int64      `gorm:"column:id;primaryKey"`
	ChallengeID    int64      `gorm:"column:challenge_id;index:idx_cp_jobs_challenge_created"`
	RequestedBy    int64      `gorm:"column:requested_by"`
	Status         string     `gorm:"column:status;size:32;index:idx_cp_jobs_status_created"`
	RequestSource  string     `gorm:"column:request_source;size:32"`
	ResultJSON     string     `gorm:"column:result_json;type:text"`
	FailureSummary string     `gorm:"column:failure_summary;size:512"`
	PublishedAt    *time.Time `gorm:"column:published_at"`
	StartedAt      *time.Time `gorm:"column:started_at"`
	FinishedAt     *time.Time `gorm:"column:finished_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (ChallengePublishCheckJob) TableName() string {
	return "challenge_publish_check_jobs"
}
