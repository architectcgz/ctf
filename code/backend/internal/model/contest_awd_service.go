package model

import (
	"time"

	"gorm.io/gorm"
)

type ContestAWDService struct {
	ID                int64                     `gorm:"column:id;primaryKey"`
	ContestID         int64                     `gorm:"column:contest_id"`
	ChallengeID       int64                     `gorm:"column:challenge_id"`
	TemplateID        *int64                    `gorm:"column:template_id"`
	DisplayName       string                    `gorm:"column:display_name"`
	ServiceSnapshot   string                    `gorm:"column:service_snapshot;type:text;default:'{}'"`
	Order             int                       `gorm:"column:order"`
	IsVisible         bool                      `gorm:"column:is_visible"`
	ScoreConfig       string                    `gorm:"column:score_config;type:text;default:'{}'"`
	RuntimeConfig     string                    `gorm:"column:runtime_config;type:text;default:'{}'"`
	ValidationState   AWDCheckerValidationState `gorm:"column:awd_checker_validation_state;size:24;not null;default:'pending'"`
	LastPreviewAt     *time.Time                `gorm:"column:awd_checker_last_preview_at"`
	LastPreviewResult string                    `gorm:"column:awd_checker_last_preview_result;type:text;default:''"`
	CreatedAt         time.Time                 `gorm:"column:created_at"`
	UpdatedAt         time.Time                 `gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt            `gorm:"column:deleted_at"`
}

func (ContestAWDService) TableName() string {
	return "contest_awd_services"
}
