package model

import (
	"time"

	"gorm.io/gorm"
)

type AWDCheckerType string

const (
	AWDCheckerTypeLegacyProbe  AWDCheckerType = "legacy_probe"
	AWDCheckerTypeHTTPStandard AWDCheckerType = "http_standard"
	AWDCheckerTypeTCPStandard  AWDCheckerType = "tcp_standard"
	AWDCheckerTypeScript       AWDCheckerType = "script_checker"
)

type AWDCheckerValidationState string

const (
	AWDCheckerValidationStatePending AWDCheckerValidationState = "pending"
	AWDCheckerValidationStatePassed  AWDCheckerValidationState = "passed"
	AWDCheckerValidationStateFailed  AWDCheckerValidationState = "failed"
	AWDCheckerValidationStateStale   AWDCheckerValidationState = "stale"
)

type ContestChallenge struct {
	ID           int64          `gorm:"column:id;primaryKey"`
	ContestID    int64          `gorm:"column:contest_id"`
	ChallengeID  int64          `gorm:"column:challenge_id"`
	Points       int            `gorm:"column:points"`
	ContestScore *int           `gorm:"column:contest_score"`
	Order        int            `gorm:"column:order"`
	IsVisible    bool           `gorm:"column:is_visible"`
	FirstBloodBy *int64         `gorm:"column:first_blood_by"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (ContestChallenge) TableName() string {
	return "contest_challenges"
}
