package queries

import (
	"context"
	"testing"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	opsports "ctf-platform/internal/module/ops/ports"
)

type stubNotificationQueryRepository struct{}

func (r *stubNotificationQueryRepository) Create(_ context.Context, _ *model.Notification) error {
	return nil
}

func (r *stubNotificationQueryRepository) List(_ context.Context, _ opsports.NotificationListFilter) ([]model.Notification, int64, error) {
	return nil, 0, nil
}

func (r *stubNotificationQueryRepository) FindByID(_ context.Context, _, _ int64) (*model.Notification, error) {
	return nil, nil
}

func (r *stubNotificationQueryRepository) MarkAsRead(_ context.Context, _, _ int64, _ any) error {
	return nil
}

func TestNotificationServiceGetNotificationsNormalizesPagination(t *testing.T) {
	service := NewNotificationService(&stubNotificationQueryRepository{}, config.PaginationConfig{
		DefaultPageSize: 20,
		MaxPageSize:     100,
	}, zap.NewNop())

	items, total, page, pageSize, err := service.GetNotifications(context.Background(), 7, &dto.NotificationQuery{
		Page:     0,
		PageSize: 999,
	})
	if err != nil {
		t.Fatalf("GetNotifications() error = %v", err)
	}
	if len(items) != 0 || total != 0 {
		t.Fatalf("GetNotifications() items=%d total=%d, want empty result", len(items), total)
	}
	if page != 1 || pageSize != 100 {
		t.Fatalf("GetNotifications() page=%d pageSize=%d, want 1 and 100", page, pageSize)
	}
}
