package model

import (
	"time"

	"gorm.io/gorm"
)

type AWDServiceType string

const (
	AWDServiceTypeWebHTTP        AWDServiceType = "web_http"
	AWDServiceTypeBinaryTCP      AWDServiceType = "binary_tcp"
	AWDServiceTypeMultiContainer AWDServiceType = "multi_container"
)

type AWDDeploymentMode string

const (
	AWDDeploymentModeSingleContainer AWDDeploymentMode = "single_container"
	AWDDeploymentModeTopology        AWDDeploymentMode = "topology"
)

type AWDChallengeStatus string

const (
	AWDChallengeStatusDraft     AWDChallengeStatus = "draft"
	AWDChallengeStatusPublished AWDChallengeStatus = "published"
	AWDChallengeStatusArchived  AWDChallengeStatus = "archived"
)

type AWDReadinessStatus string

const (
	AWDReadinessStatusPending AWDReadinessStatus = "pending"
	AWDReadinessStatusPassed  AWDReadinessStatus = "passed"
	AWDReadinessStatusFailed  AWDReadinessStatus = "failed"
)

type AWDChallenge struct {
	ID               int64              `gorm:"column:id;primaryKey"`
	Name             string             `gorm:"column:name"`
	Slug             string             `gorm:"column:slug"`
	Category         string             `gorm:"column:category"`
	Difficulty       string             `gorm:"column:difficulty"`
	Description      string             `gorm:"column:description"`
	ServiceType      AWDServiceType     `gorm:"column:service_type"`
	DeploymentMode   AWDDeploymentMode  `gorm:"column:deployment_mode"`
	Version          string             `gorm:"column:version"`
	Status           AWDChallengeStatus `gorm:"column:status"`
	CheckerType      AWDCheckerType     `gorm:"column:checker_type"`
	CheckerConfig    string             `gorm:"column:checker_config"`
	FlagMode         string             `gorm:"column:flag_mode"`
	FlagConfig       string             `gorm:"column:flag_config"`
	DefenseEntryMode string             `gorm:"column:defense_entry_mode"`
	AccessConfig     string             `gorm:"column:access_config"`
	RuntimeConfig    string             `gorm:"column:runtime_config"`
	ReadinessStatus  AWDReadinessStatus `gorm:"column:readiness_status"`
	ReadinessReport  string             `gorm:"column:readiness_report"`
	LastVerifiedAt   *time.Time         `gorm:"column:last_verified_at"`
	LastVerifiedBy   *int64             `gorm:"column:last_verified_by"`
	CreatedBy        *int64             `gorm:"column:created_by"`
	CreatedAt        time.Time          `gorm:"column:created_at"`
	UpdatedAt        time.Time          `gorm:"column:updated_at"`
	DeletedAt        gorm.DeletedAt     `gorm:"column:deleted_at"`
}

func (AWDChallenge) TableName() string {
	return "awd_challenges"
}
