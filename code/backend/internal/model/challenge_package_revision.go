package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ChallengePackageRevisionSourceImported = "imported"
	ChallengePackageRevisionSourceExported = "exported"
)

type ChallengePackageRevision struct {
	ID                 int64          `gorm:"column:id;primaryKey"`
	ChallengeID        int64          `gorm:"column:challenge_id;index:idx_package_revisions_challenge_revision,priority:1"`
	RevisionNo         int            `gorm:"column:revision_no;index:idx_package_revisions_challenge_revision,priority:2"`
	SourceType         string         `gorm:"column:source_type"`
	ParentRevisionID   *int64         `gorm:"column:parent_revision_id"`
	PackageSlug        string         `gorm:"column:package_slug"`
	ArchivePath        string         `gorm:"column:archive_path"`
	SourceDir          string         `gorm:"column:source_dir"`
	ManifestSnapshot   string         `gorm:"column:manifest_snapshot"`
	TopologySourcePath string         `gorm:"column:topology_source_path"`
	TopologySnapshot   string         `gorm:"column:topology_snapshot"`
	CreatedBy          *int64         `gorm:"column:created_by"`
	CreatedAt          time.Time      `gorm:"column:created_at"`
	UpdatedAt          time.Time      `gorm:"column:updated_at"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (ChallengePackageRevision) TableName() string {
	return "challenge_package_revisions"
}
