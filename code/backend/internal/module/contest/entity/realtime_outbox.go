package entity

import "time"

type ContestRealtimeOutbox struct {
	ID              int64      `gorm:"column:id;primaryKey"`
	EventName       string     `gorm:"column:event_name"`
	Delivery        string     `gorm:"column:delivery"`
	Channel         string     `gorm:"column:channel"`
	RecipientUserID *int64     `gorm:"column:recipient_user_id"`
	MessageType     string     `gorm:"column:message_type"`
	Payload         string     `gorm:"column:payload;type:jsonb"`
	DedupeKey       string     `gorm:"column:dedupe_key;uniqueIndex"`
	Status          string     `gorm:"column:status;index"`
	AttemptCount    int        `gorm:"column:attempt_count"`
	EventOccurredAt time.Time  `gorm:"column:event_occurred_at"`
	NextAttemptAt   time.Time  `gorm:"column:next_attempt_at;index"`
	LastError       string     `gorm:"column:last_error"`
	StreamMessageID string     `gorm:"column:stream_message_id"`
	SentAt          *time.Time `gorm:"column:sent_at"`
	CreatedAt       time.Time  `gorm:"column:created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at"`
}

func (ContestRealtimeOutbox) TableName() string {
	return "contest_realtime_outbox"
}
