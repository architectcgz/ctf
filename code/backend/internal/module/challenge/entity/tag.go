package entity

import "time"

type Tag struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Type        string    `gorm:"column:type"`
	Description string    `gorm:"column:description"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (Tag) TableName() string {
	return "tags"
}

type ChallengeTag struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	ChallengeID int64     `gorm:"column:challenge_id"`
	TagID       int64     `gorm:"column:tag_id"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (ChallengeTag) TableName() string {
	return "challenge_tags"
}
