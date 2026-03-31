package model

import "time"

type NotificationBatch struct {
	ID             int64     `gorm:"column:id;primaryKey"`
	CreatedBy      int64     `gorm:"column:created_by"`
	Type           string    `gorm:"column:type"`
	Title          string    `gorm:"column:title"`
	Content        string    `gorm:"column:content"`
	Link           *string   `gorm:"column:link"`
	AudienceMode   string    `gorm:"column:audience_mode"`
	AudienceRules  string    `gorm:"column:audience_rules"`
	RecipientCount int       `gorm:"column:recipient_count"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	PublishedAt    time.Time `gorm:"column:published_at"`
}

func (NotificationBatch) TableName() string {
	return "notification_batches"
}
