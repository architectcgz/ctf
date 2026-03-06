package model

import "time"

type Team struct {
	ID        int64     `gorm:"primaryKey"`
	ContestID int64     `gorm:"not null;index"`
	Name      string    `gorm:"not null"`
	UserID    int64     `gorm:"not null;index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
