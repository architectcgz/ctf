package model

import "time"

const (
	DimensionWeb      = "web"
	DimensionPwn      = "pwn"
	DimensionReverse  = "reverse"
	DimensionCrypto   = "crypto"
	DimensionMisc     = "misc"
	DimensionForensics = "forensics"
)

type SkillProfile struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	UserID    int64     `gorm:"column:user_id;not null;uniqueIndex:idx_user_dimension"`
	Dimension string    `gorm:"column:dimension;not null;uniqueIndex:idx_user_dimension"`
	Score     float64   `gorm:"column:score;not null;default:0"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (SkillProfile) TableName() string {
	return "skill_profiles"
}
