package infrastructure

import (
	"context"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	opsports "ctf-platform/internal/module/ops/ports"
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

func (r *NotificationRepository) CreateBatch(ctx context.Context, batch *model.NotificationBatch, notifications []*model.Notification) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(batch).Error; err != nil {
			return err
		}
		if len(notifications) == 0 {
			return nil
		}
		for _, item := range notifications {
			item.BatchID = &batch.ID
		}
		return tx.Create(notifications).Error
	})
}

func (r *NotificationRepository) List(ctx context.Context, filter opsports.NotificationListFilter) ([]model.Notification, int64, error) {
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

func (r *NotificationRepository) ListAllUserIDs(ctx context.Context) ([]int64, error) {
	return r.listUserIDs(ctx, nil)
}

func (r *NotificationRepository) ListUserIDsByRoles(ctx context.Context, roles []string) ([]int64, error) {
	if len(roles) == 0 {
		return []int64{}, nil
	}
	return r.listUserIDs(ctx, func(query *gorm.DB) *gorm.DB {
		return query.Where("role IN ?", roles)
	})
}

func (r *NotificationRepository) ListUserIDsByClasses(ctx context.Context, classNames []string) ([]int64, error) {
	if len(classNames) == 0 {
		return []int64{}, nil
	}
	return r.listUserIDs(ctx, func(query *gorm.DB) *gorm.DB {
		return query.Where("class_name IN ?", classNames)
	})
}

func (r *NotificationRepository) ListExistingUserIDs(ctx context.Context, userIDs []int64) ([]int64, error) {
	if len(userIDs) == 0 {
		return []int64{}, nil
	}
	return r.listUserIDs(ctx, func(query *gorm.DB) *gorm.DB {
		return query.Where("id IN ?", userIDs)
	})
}

func (r *NotificationRepository) MarkAsRead(ctx context.Context, notificationID, userID int64, readAt any) error {
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]any{"is_read": true, "read_at": readAt}).Error
}

func (r *NotificationRepository) listUserIDs(ctx context.Context, apply func(query *gorm.DB) *gorm.DB) ([]int64, error) {
	query := r.db.WithContext(ctx).Model(&model.User{}).Where("deleted_at IS NULL")
	if apply != nil {
		query = apply(query)
	}
	userIDs := make([]int64, 0)
	if err := query.Order("id ASC").Pluck("id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}
