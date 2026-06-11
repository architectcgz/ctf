package entity

import (
	"time"

	"gorm.io/gorm"
)

const (
	ChallengeStatusPublished = "published"

	FlagTypeStatic       = "static"
	FlagTypeDynamic      = "dynamic"
	FlagTypeRegex        = "regex"
	FlagTypeManualReview = "manual_review"

	ChallengeTargetProtocolHTTP = "http"
	ChallengeTargetProtocolTCP  = "tcp"

	InstanceSharingPerUser = "per_user"
	InstanceSharingPerTeam = "per_team"
	InstanceSharingShared  = "shared"
)

// Challenge is a practice-facing challenge projection used by scoring/runtime flows.
type Challenge struct {
	ID              int64   `gorm:"column:id;primaryKey"`
	PackageSlug     *string `gorm:"column:package_slug"`
	Title           string  `gorm:"column:title"`
	Description     string  `gorm:"column:description"`
	Category        string  `gorm:"column:category"`
	Difficulty      string  `gorm:"column:difficulty"`
	Points          int     `gorm:"column:points"`
	ImageID         *int64  `gorm:"column:image_id"`
	AttachmentURL   string  `gorm:"column:attachment_url"`
	Status          string  `gorm:"column:status"`
	FlagType        string  `gorm:"column:flag_type"`
	FlagHash        string  `gorm:"column:flag_hash"`
	FlagSalt        string  `gorm:"column:flag_salt"`
	FlagRegex       string  `gorm:"column:flag_regex"`
	FlagPrefix      string  `gorm:"column:flag_prefix"`
	InstanceSharing string  `gorm:"column:instance_sharing"`
	TargetProtocol  string  `gorm:"column:target_protocol"`
	TargetPort      int     `gorm:"column:target_port"`
	CreatedBy       *int64  `gorm:"column:created_by"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Challenge) TableName() string {
	return "challenges"
}
