package model

import "time"

const (
	ContestRegistrationStatusPending  = "pending"
	ContestRegistrationStatusApproved = "approved"
	ContestRegistrationStatusRejected = "rejected"
)

type ContestRegistration struct {
	ID         int64      `gorm:"column:id;primaryKey"`
	ContestID  int64      `gorm:"column:contest_id;not null;uniqueIndex:uk_contest_reg_user"`
	UserID     int64      `gorm:"column:user_id;not null;uniqueIndex:uk_contest_reg_user"`
	TeamID     *int64     `gorm:"column:team_id"`
	Status     string     `gorm:"column:status;size:16;not null;default:'pending'"`
	ReviewedBy *int64     `gorm:"column:reviewed_by"`
	ReviewedAt *time.Time `gorm:"column:reviewed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;not null"`
}

func (ContestRegistration) TableName() string {
	return "contest_registrations"
}
