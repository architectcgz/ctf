package model

import "time"

// Instance 实例模型
type Instance struct {
	ID           int64     `gorm:"primaryKey"`
	UserID       int64     `gorm:"not null;index"`
	ChallengeID  int64     `gorm:"not null;index"`
	ContainerID  string    `gorm:"size:64;not null"`
	NetworkID    string    `gorm:"size:64"`
	Status       string    `gorm:"size:16;not null;index"`
	AccessURL    string    `gorm:"size:255"`
	Nonce        string    `gorm:"size:64"`
	ExpiresAt    time.Time `gorm:"not null;index"`
	ExtendCount  int       `gorm:"default:0"`
	MaxExtends   int       `gorm:"default:2"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// 状态常量
const (
	InstanceStatusCreating = "creating"
	InstanceStatusRunning  = "running"
	InstanceStatusStopped  = "stopped"
	InstanceStatusExpired  = "expired"
	InstanceStatusFailed   = "failed"
)
