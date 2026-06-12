package ports

import (
	"context"
	"errors"

	opsentity "ctf-platform/internal/module/ops/entity"
	platformevents "ctf-platform/internal/platform/events"
	ctfws "ctf-platform/internal/websocket"
)

var ErrNotificationNotFound = errors.New("ops notification not found")

type NotificationListFilter struct {
	UserID int64
	Type   string
	Offset int
	Limit  int
}

type NotificationCommandRepository interface {
	Create(ctx context.Context, notification *opsentity.Notification) error
	CreateIfSourceEventAbsent(ctx context.Context, notification *opsentity.Notification) (bool, error)
	CreateBatch(ctx context.Context, batch *opsentity.NotificationBatch, notifications []*opsentity.Notification) error
	FindByID(ctx context.Context, notificationID, userID int64) (*opsentity.Notification, error)
	MarkAsRead(ctx context.Context, notificationID, userID int64, readAt any) error
	ListAllUserIDs(ctx context.Context) ([]int64, error)
	ListUserIDsByRoles(ctx context.Context, roles []string) ([]int64, error)
	ListUserIDsByClasses(ctx context.Context, classNames []string) ([]int64, error)
	ListExistingUserIDs(ctx context.Context, userIDs []int64) ([]int64, error)
}

type NotificationOutboxTxRepository interface {
	Create(ctx context.Context, notification *opsentity.Notification) error
	CreateIfSourceEventAbsent(ctx context.Context, notification *opsentity.Notification) (bool, error)
	CreateBatch(ctx context.Context, batch *opsentity.NotificationBatch, notifications []*opsentity.Notification) error
	MarkAsRead(ctx context.Context, notificationID, userID int64, readAt any) error
	platformevents.OutboxEventEnqueuer
}

type NotificationOutboxTxManager interface {
	WithinNotificationOutboxTx(ctx context.Context, fn func(txRepo NotificationOutboxTxRepository) error) error
}

type NotificationQueryRepository interface {
	List(ctx context.Context, filter NotificationListFilter) ([]opsentity.Notification, int64, error)
}

type NotificationBroadcaster interface {
	SendToUser(userID int64, message ctfws.Envelope) int
	Broadcast(message ctfws.Envelope) int
}
