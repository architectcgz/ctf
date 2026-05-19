package commands

import (
	"time"

	"gorm.io/gorm"
)

type practiceCommandImageRow struct {
	ID          int64          `gorm:"column:id;primaryKey"`
	Name        string         `gorm:"column:name;uniqueIndex:idx_name_tag"`
	Tag         string         `gorm:"column:tag;uniqueIndex:idx_name_tag"`
	Description string         `gorm:"column:description"`
	Size        int64          `gorm:"column:size"`
	Status      string         `gorm:"column:status"`
	Digest      string         `gorm:"column:digest"`
	SourceType  string         `gorm:"column:source_type"`
	BuildJobID  *int64         `gorm:"column:build_job_id"`
	LastError   string         `gorm:"column:last_error"`
	VerifiedAt  *time.Time     `gorm:"column:verified_at"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (practiceCommandImageRow) TableName() string {
	return "images"
}

type practiceCommandChallengeRow struct {
	ID              int64          `gorm:"column:id;primaryKey"`
	PackageSlug     *string        `gorm:"column:package_slug;size:128;uniqueIndex:uq_challenges_package_slug"`
	Title           string         `gorm:"column:title"`
	Description     string         `gorm:"column:description"`
	Category        string         `gorm:"column:category"`
	Difficulty      string         `gorm:"column:difficulty"`
	Points          int            `gorm:"column:points"`
	ImageID         int64          `gorm:"column:image_id"`
	AttachmentURL   string         `gorm:"column:attachment_url"`
	Status          string         `gorm:"column:status"`
	FlagType        string         `gorm:"column:flag_type;default:'static'"`
	FlagHash        string         `gorm:"column:flag_hash;size:128"`
	FlagSalt        string         `gorm:"column:flag_salt;size:64"`
	FlagRegex       string         `gorm:"column:flag_regex;size:512"`
	FlagPrefix      string         `gorm:"column:flag_prefix;size:32;default:'flag'"`
	InstanceSharing string         `gorm:"column:instance_sharing;size:16;default:'per_user'"`
	TargetProtocol  string         `gorm:"column:target_protocol;size:16;default:'http'"`
	TargetPort      int            `gorm:"column:target_port;default:0"`
	CreatedBy       *int64         `gorm:"column:created_by"`
	CreatedAt       time.Time      `gorm:"column:created_at"`
	UpdatedAt       time.Time      `gorm:"column:updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (practiceCommandChallengeRow) TableName() string {
	return "challenges"
}

type practiceCommandChallengeTopologyRow struct {
	ID                   int64      `gorm:"column:id;primaryKey"`
	ChallengeID          int64      `gorm:"column:challenge_id;uniqueIndex"`
	TemplateID           *int64     `gorm:"column:template_id"`
	EntryNodeKey         string     `gorm:"column:entry_node_key"`
	Spec                 string     `gorm:"column:spec"`
	SourceType           string     `gorm:"column:source_type"`
	SourcePath           string     `gorm:"column:source_path"`
	PackageRevisionID    *int64     `gorm:"column:package_revision_id"`
	PackageBaselineSpec  string     `gorm:"column:package_baseline_spec"`
	SyncStatus           string     `gorm:"column:sync_status"`
	LastExportRevisionID *int64     `gorm:"column:last_export_revision_id"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	UpdatedAt            time.Time  `gorm:"column:updated_at"`
	DeletedAt            *time.Time `gorm:"column:deleted_at"`
}

func (practiceCommandChallengeTopologyRow) TableName() string {
	return "challenge_topologies"
}
