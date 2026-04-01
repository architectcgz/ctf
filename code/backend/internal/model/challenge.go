package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	ChallengeDifficultyBeginner = "beginner"
	ChallengeDifficultyEasy     = "easy"
	ChallengeDifficultyMedium   = "medium"
	ChallengeDifficultyHard     = "hard"
	ChallengeDifficultyInsane   = "insane"

	ChallengeStatusDraft     = "draft"
	ChallengeStatusPublished = "published"
	ChallengeStatusArchived  = "archived"

	FlagTypeStatic  = "static"
	FlagTypeDynamic = "dynamic"
)

type Challenge struct {
	ID            int64          `gorm:"column:id;primaryKey"`
	PackageSlug   *string        `gorm:"column:package_slug;size:128;uniqueIndex:uq_challenges_package_slug"`
	Title         string         `gorm:"column:title"`
	Description   string         `gorm:"column:description"`
	Category      string         `gorm:"column:category"`
	Difficulty    string         `gorm:"column:difficulty"`
	Points        int            `gorm:"column:points"`
	ImageID       int64          `gorm:"column:image_id"`
	AttachmentURL string         `gorm:"column:attachment_url"`
	Status        string         `gorm:"column:status"`
	FlagType      string         `gorm:"column:flag_type;default:'static'"`
	FlagHash      string         `gorm:"column:flag_hash;size:128"`
	FlagSalt      string         `gorm:"column:flag_salt;size:64"`
	FlagPrefix    string         `gorm:"column:flag_prefix;size:32;default:'flag'"`
	CreatedBy     *int64         `gorm:"column:created_by"`
	CreatedAt     time.Time      `gorm:"column:created_at"`
	UpdatedAt     time.Time      `gorm:"column:updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Challenge) TableName() string {
	return "challenges"
}
