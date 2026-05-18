package entity

import "time"

// SkillProfile 用户能力画像持久化实体。
type SkillProfile struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	UserID    int64     `gorm:"column:user_id;not null;uniqueIndex:uk_skill_profiles_user_dimension,priority:1;index:idx_skill_profiles_user_id"`
	Dimension string    `gorm:"column:dimension;size:20;not null;uniqueIndex:uk_skill_profiles_user_dimension,priority:2"`
	Score     float64   `gorm:"column:score;not null;default:0"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;index:idx_skill_profiles_updated_at"`
}

func (SkillProfile) TableName() string {
	return "skill_profiles"
}
