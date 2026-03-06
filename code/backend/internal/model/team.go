package model

import (
	"time"

	"gorm.io/gorm"
)

type Team struct {
	ID          int64          `gorm:"column:id;primaryKey"`
	ContestID   int64          `gorm:"column:contest_id;index"`
	Name        string         `gorm:"column:name"`
	CaptainID   int64          `gorm:"column:captain_id"`
	InviteCode  string         `gorm:"column:invite_code;uniqueIndex"`
	MaxMembers  int            `gorm:"column:max_members;default:4"`
	TotalScore  int            `gorm:"column:total_score;default:0"`
	LastSolveAt *time.Time     `gorm:"column:last_solve_at"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Team) TableName() string {
	return "teams"
}

type TeamMember struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	ContestID int64     `gorm:"column:contest_id"`
	TeamID    int64     `gorm:"column:team_id;index:idx_team_user"`
	UserID    int64     `gorm:"column:user_id;index:idx_team_user"`
	JoinedAt  time.Time `gorm:"column:joined_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (TeamMember) TableName() string {
	return "team_members"
}
