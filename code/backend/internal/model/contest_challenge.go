package model

import (
	"time"

	"gorm.io/gorm"
)

type AWDCheckerType string

const (
	AWDCheckerTypeLegacyProbe  AWDCheckerType = "legacy_probe"
	AWDCheckerTypeHTTPStandard AWDCheckerType = "http_standard"
)

type ContestChallenge struct {
	ID               int64          `gorm:"column:id;primaryKey"`
	ContestID        int64          `gorm:"column:contest_id"`
	ChallengeID      int64          `gorm:"column:challenge_id"`
	Points           int            `gorm:"column:points"`
	ContestScore     *int           `gorm:"column:contest_score"`
	Order            int            `gorm:"column:order"`
	IsVisible        bool           `gorm:"column:is_visible"`
	AWDCheckerType   AWDCheckerType `gorm:"column:awd_checker_type;size:32;default:''"`
	AWDCheckerConfig string         `gorm:"column:awd_checker_config;type:text;default:'{}'"`
	AWDSLAScore      int            `gorm:"column:awd_sla_score;not null;default:0"`
	AWDDefenseScore  int            `gorm:"column:awd_defense_score;not null;default:0"`
	FirstBloodBy     *int64         `gorm:"column:first_blood_by"`
	CreatedAt        time.Time      `gorm:"column:created_at"`
	UpdatedAt        time.Time      `gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (ContestChallenge) TableName() string {
	return "contest_challenges"
}
