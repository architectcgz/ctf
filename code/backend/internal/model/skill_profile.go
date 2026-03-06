package model

import "time"

// 能力维度枚举
const (
	DimensionWeb       = "web"
	DimensionPwn       = "pwn"
	DimensionReverse   = "reverse"
	DimensionCrypto    = "crypto"
	DimensionMisc      = "misc"
	DimensionForensics = "forensics"
)

// SkillProfile 用户能力画像
type SkillProfile struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	UserID    int64     `gorm:"column:user_id;not null;index:idx_user_dimension"`
	Dimension string    `gorm:"column:dimension;size:20;not null;index:idx_user_dimension"`
	Score     float64   `gorm:"column:score;not null;default:0"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

func (SkillProfile) TableName() string {
	return "skill_profiles"
}
