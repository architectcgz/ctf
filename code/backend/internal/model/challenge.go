package model

import (
	"time"

	"gorm.io/gorm"
)

type ChallengeStatus string

const (
	ChallengeDifficultyBeginner = "beginner"
	ChallengeDifficultyEasy     = "easy"
	ChallengeDifficultyMedium   = "medium"
	ChallengeDifficultyHard     = "hard"
	ChallengeDifficultyInsane   = "insane"

	ChallengeStatusDraft     ChallengeStatus = "draft"
	ChallengeStatusPublished ChallengeStatus = "published"
	ChallengeStatusArchived  ChallengeStatus = "archived"

	FlagTypeStatic       = "static"
	FlagTypeDynamic      = "dynamic"
	FlagTypeRegex        = "regex"
	FlagTypeManualReview = "manual_review"

	ChallengeTargetProtocolHTTP = "http"
	ChallengeTargetProtocolTCP  = "tcp"
)

type InstanceSharing string

type ShareScope = InstanceSharing

const (
	InstanceSharingPerUser InstanceSharing = "per_user"
	InstanceSharingPerTeam InstanceSharing = "per_team"
	InstanceSharingShared  InstanceSharing = "shared"

	ShareScopePerUser = InstanceSharingPerUser
	ShareScopePerTeam = InstanceSharingPerTeam
	ShareScopeShared  = InstanceSharingShared
)

type Challenge struct {
	ID              int64           `gorm:"column:id;primaryKey"`
	PackageSlug     *string         `gorm:"column:package_slug;size:128;uniqueIndex:uq_challenges_package_slug"`
	Title           string          `gorm:"column:title"`
	Description     string          `gorm:"column:description"`
	Category        string          `gorm:"column:category"`
	Difficulty      string          `gorm:"column:difficulty"`
	Points          int             `gorm:"column:points"`
	ImageID         int64           `gorm:"column:image_id"`
	AttachmentURL   string          `gorm:"column:attachment_url"`
	Status          ChallengeStatus `gorm:"column:status"`
	FlagType        string          `gorm:"column:flag_type;default:'static'"`
	FlagHash        string          `gorm:"column:flag_hash;size:128"`
	FlagSalt        string          `gorm:"column:flag_salt;size:64"`
	FlagRegex       string          `gorm:"column:flag_regex;size:512"`
	FlagPrefix      string          `gorm:"column:flag_prefix;size:32;default:'flag'"`
	InstanceSharing InstanceSharing `gorm:"column:instance_sharing;size:16;default:'per_user'"`
	TargetProtocol  string          `gorm:"column:target_protocol;size:16;default:'http'"`
	TargetPort      int             `gorm:"column:target_port;default:0"`
	CreatedBy       *int64          `gorm:"column:created_by"`
	CreatedAt       time.Time       `gorm:"column:created_at"`
	UpdatedAt       time.Time       `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt  `gorm:"column:deleted_at"`

	RecommendationDimension string `gorm:"column:recommendation_dimension;->;-:migration"`
}

func (Challenge) TableName() string {
	return "challenges"
}
