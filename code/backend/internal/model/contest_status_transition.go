package model

import "time"

type ContestStatusTransition struct {
	ID               int64     `gorm:"column:id;primaryKey"`
	ContestID        int64     `gorm:"column:contest_id;uniqueIndex:uk_contest_status_transitions_contest_version"`
	StatusVersion    int64     `gorm:"column:status_version;uniqueIndex:uk_contest_status_transitions_contest_version"`
	FromStatus       string    `gorm:"column:from_status"`
	ToStatus         string    `gorm:"column:to_status"`
	Reason           string    `gorm:"column:reason"`
	AppliedBy        string    `gorm:"column:applied_by"`
	SideEffectStatus string    `gorm:"column:side_effect_status"`
	SideEffectError  string    `gorm:"column:side_effect_error"`
	OccurredAt       time.Time `gorm:"column:occurred_at;index:idx_contest_status_transitions_occurred_at"`
	CreatedAt        time.Time `gorm:"column:created_at"`
	UpdatedAt        time.Time `gorm:"column:updated_at"`
}

func (ContestStatusTransition) TableName() string {
	return "contest_status_transitions"
}
