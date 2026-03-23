package infrastructure

import (
	"context"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	opsapp "ctf-platform/internal/module/ops/application"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) Create(ctx context.Context, notification *model.Notification) error {
	return r.db.WithContext(ctx).Create(notification).Error
}

func (r *NotificationRepository) List(ctx context.Context, filter opsapp.NotificationListFilter) ([]model.Notification, int64, error) {
	query := r.db.WithContext(ctx).Model(&model.Notification{}).Where("user_id = ?", filter.UserID)
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]model.Notification, 0)
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *NotificationRepository) FindByID(ctx context.Context, notificationID, userID int64) (*model.Notification, error) {
	var notification model.Notification
	if err := r.db.WithContext(ctx).
		Where("id = ? AND user_id = ?", notificationID, userID).
		First(&notification).Error; err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, notificationID, userID int64, readAt any) error {
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]any{"is_read": true, "read_at": readAt}).Error
}
