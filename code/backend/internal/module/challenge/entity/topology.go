package entity

import "time"

const (
	ChallengeTopologySourceTypeManual        = "platform_manual"
	ChallengeTopologySourceTypePackageImport = "package_import"

	ChallengeTopologySyncStatusClean   = "clean"
	ChallengeTopologySyncStatusDrifted = "drifted"

	WriteupVisibilityPrivate = "private"
	WriteupVisibilityPublic  = "public"
)

type ChallengeTopology struct {
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

func (ChallengeTopology) TableName() string {
	return "challenge_topologies"
}

type EnvironmentTemplate struct {
	ID           int64      `gorm:"column:id;primaryKey"`
	Name         string     `gorm:"column:name"`
	Description  string     `gorm:"column:description"`
	EntryNodeKey string     `gorm:"column:entry_node_key"`
	Spec         string     `gorm:"column:spec"`
	UsageCount   int        `gorm:"column:usage_count"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

func (EnvironmentTemplate) TableName() string {
	return "environment_templates"
}

type ChallengeWriteup struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	ChallengeID   int64      `gorm:"column:challenge_id;uniqueIndex"`
	Title         string     `gorm:"column:title"`
	Content       string     `gorm:"column:content"`
	Visibility    string     `gorm:"column:visibility"`
	CreatedBy     *int64     `gorm:"column:created_by"`
	IsRecommended bool       `gorm:"column:is_recommended;index:idx_challenge_writeups_recommended"`
	RecommendedAt *time.Time `gorm:"column:recommended_at"`
	RecommendedBy *int64     `gorm:"column:recommended_by"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (ChallengeWriteup) TableName() string {
	return "challenge_writeups"
}
