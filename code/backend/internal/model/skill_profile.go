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

// ValidDimensions 合法维度集合
var ValidDimensions = map[string]bool{
	DimensionWeb:       true,
	DimensionPwn:       true,
	DimensionReverse:   true,
	DimensionCrypto:    true,
	DimensionMisc:      true,
	DimensionForensics: true,
}

// AllDimensions 所有维度列表
var AllDimensions = []string{
	DimensionWeb,
	DimensionPwn,
	DimensionReverse,
	DimensionCrypto,
	DimensionMisc,
	DimensionForensics,
}

// IsValidDimension 检查维度是否合法
func IsValidDimension(dimension string) bool {
	return ValidDimensions[dimension]
}

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
