package model

import "time"

// UserScore 用户得分模型
type UserScore struct {
	UserID      int64     `gorm:"column:user_id;primaryKey"`
	TotalScore  int       `gorm:"column:total_score;not null;default:0"`
	SolvedCount int       `gorm:"column:solved_count;not null;default:0"`
	Rank        int       `gorm:"column:rank;not null;default:0;index"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (UserScore) TableName() string {
	return "user_scores"
}
