package ports

import (
	"context"
	"errors"

	"ctf-platform/internal/model"
	ctfws "ctf-platform/pkg/websocket"
)

var ErrNotificationNotFound = errors.New("ops notification not found")

type NotificationListFilter struct {
	UserID int64
	Type   string
	Offset int
	Limit  int
}

type NotificationCommandRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	CreateBatch(ctx context.Context, batch *model.NotificationBatch, notifications []*model.Notification) error
	FindByID(ctx context.Context, notificationID, userID int64) (*model.Notification, error)
	MarkAsRead(ctx context.Context, notificationID, userID int64, readAt any) error
	ListAllUserIDs(ctx context.Context) ([]int64, error)
	ListUserIDsByRoles(ctx context.Context, roles []string) ([]int64, error)
	ListUserIDsByClasses(ctx context.Context, classNames []string) ([]int64, error)
	ListExistingUserIDs(ctx context.Context, userIDs []int64) ([]int64, error)
}

type NotificationQueryRepository interface {
	List(ctx context.Context, filter NotificationListFilter) ([]model.Notification, int64, error)
}

type NotificationBroadcaster interface {
	SendToUser(userID int64, message ctfws.Envelope) int
	Broadcast(message ctfws.Envelope) int
}
