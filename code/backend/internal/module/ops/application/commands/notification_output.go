package commands

import "time"

const (
	NotificationAudienceTypeAll   = "all"
	NotificationAudienceTypeRole  = "role"
	NotificationAudienceTypeClass = "class"
	NotificationAudienceTypeUser  = "user"
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

type AdminNotificationPublishResp struct {
	BatchID        int64 `json:"batch_id"`
	RecipientCount int   `json:"recipient_count"`
}
