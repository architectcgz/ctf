package model

import "time"

type ContestAnnouncement struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	ContestID int64     `gorm:"column:contest_id;index"`
	Title     string    `gorm:"column:title"`
	Content   string    `gorm:"column:content"`
	CreatedBy *int64    `gorm:"column:created_by"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (ContestAnnouncement) TableName() string {
	return "contest_announcements"
}
