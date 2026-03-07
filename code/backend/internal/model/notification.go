package model

import "time"

const (
	NotificationTypeSystem    = "system"
	NotificationTypeContest   = "contest"
	NotificationTypeChallenge = "challenge"
	NotificationTypeTeam      = "team"
)

type Notification struct {
	ID        int64      `gorm:"column:id;primaryKey"`
	UserID    int64      `gorm:"column:user_id"`
	Type      string     `gorm:"column:type"`
	Title     string     `gorm:"column:title"`
	Content   string     `gorm:"column:content"`
	IsRead    bool       `gorm:"column:is_read"`
	Link      *string    `gorm:"column:link"`
	CreatedAt time.Time  `gorm:"column:created_at"`
	ReadAt    *time.Time `gorm:"column:read_at"`
}

func (Notification) TableName() string {
	return "notifications"
}
