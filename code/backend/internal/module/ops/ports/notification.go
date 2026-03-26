package ports

import (
	"context"

	"ctf-platform/internal/model"
	ctfws "ctf-platform/pkg/websocket"
)

type NotificationListFilter struct {
	UserID int64
	Type   string
	Offset int
	Limit  int
}

type NotificationRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	List(ctx context.Context, filter NotificationListFilter) ([]model.Notification, int64, error)
	FindByID(ctx context.Context, notificationID, userID int64) (*model.Notification, error)
	MarkAsRead(ctx context.Context, notificationID, userID int64, readAt any) error
}

type NotificationBroadcaster interface {
	SendToUser(userID int64, message ctfws.Envelope) int
	Broadcast(message ctfws.Envelope) int
}
