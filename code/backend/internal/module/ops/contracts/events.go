package contracts

import "time"

const (
	EventNotificationCreated        = "notification.created"
	EventNotificationRead           = "notification.read"
	EventNotificationPayloadVersion = 1
)

type NotificationInfo struct {
	ID        int64      `json:"id"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Content   *string    `json:"content,omitempty"`
	Unread    bool       `json:"unread"`
	Link      *string    `json:"link,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
}

type NotificationFanoutEvent struct {
	UserID       int64            `json:"user_id"`
	Notification NotificationInfo `json:"notification"`
	OccurredAt   time.Time        `json:"occurred_at"`
}
