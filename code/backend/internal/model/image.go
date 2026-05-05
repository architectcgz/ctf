package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ImageStatusPending   = "pending"
	ImageStatusBuilding  = "building"
	ImageStatusPushed    = "pushed"
	ImageStatusVerifying = "verifying"
	ImageStatusAvailable = "available"
	ImageStatusFailed    = "failed"
)

const (
	ImageSourceTypeManual        = "manual"
	ImageSourceTypePlatformBuild = "platform_build"
	ImageSourceTypeExternalRef   = "external_ref"
)

type Image struct {
	ID          int64          `gorm:"column:id;primaryKey"`
	Name        string         `gorm:"column:name;uniqueIndex:idx_name_tag"`
	Tag         string         `gorm:"column:tag;uniqueIndex:idx_name_tag"`
	Description string         `gorm:"column:description"`
	Size        int64          `gorm:"column:size"` // 字节
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

func (Image) TableName() string {
	return "images"
}
