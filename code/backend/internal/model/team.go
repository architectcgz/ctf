package model

import "time"

type Team struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	Name        string     `gorm:"column:name;size:64;not null"`
	ContestID   int64      `gorm:"column:contest_id;not null;index"`
	CaptainID   int64      `gorm:"column:captain_id;not null"`
	TotalScore  int        `gorm:"column:total_score;not null;default:0"`
	LastSolveAt *time.Time `gorm:"column:last_solve_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;not null"`
}

func (Team) TableName() string {
	return "teams"
}
