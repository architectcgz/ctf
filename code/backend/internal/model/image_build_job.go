package model

import "time"

const (
	ImageBuildJobStatusPending   = "pending"
	ImageBuildJobStatusBuilding  = "building"
	ImageBuildJobStatusPushed    = "pushed"
	ImageBuildJobStatusVerifying = "verifying"
	ImageBuildJobStatusAvailable = "available"
	ImageBuildJobStatusFailed    = "failed"
)

type ImageBuildJob struct {
	ID             int64      `gorm:"column:id;primaryKey"`
	SourceType     string     `gorm:"column:source_type"`
	ChallengeMode  string     `gorm:"column:challenge_mode"`
	PackageSlug    string     `gorm:"column:package_slug"`
	SourceDir      string     `gorm:"column:source_dir"`
	DockerfilePath string     `gorm:"column:dockerfile_path"`
	ContextPath    string     `gorm:"column:context_path"`
	TargetRef      string     `gorm:"column:target_ref"`
	TargetDigest   string     `gorm:"column:target_digest"`
	Status         string     `gorm:"column:status"`
	LogPath        string     `gorm:"column:log_path"`
	ErrorSummary   string     `gorm:"column:error_summary"`
	CreatedBy      *int64     `gorm:"column:created_by"`
	StartedAt      *time.Time `gorm:"column:started_at"`
	FinishedAt     *time.Time `gorm:"column:finished_at"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedAt      time.Time  `gorm:"column:updated_at"`
}

func (ImageBuildJob) TableName() string {
	return "image_build_jobs"
}
