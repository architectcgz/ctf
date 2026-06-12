package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	identitycontracts "ctf-platform/internal/module/identity/contracts"
	opsentity "ctf-platform/internal/module/ops/entity"
	opsports "ctf-platform/internal/module/ops/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type NotificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) WithDB(db *gorm.DB) *NotificationRepository {
	return &NotificationRepository{db: db}
}

func (r *NotificationRepository) dbWithContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		return r.db
	}
	return r.db.WithContext(ctx)
}

func (r *NotificationRepository) Create(ctx context.Context, notification *opsentity.Notification) error {
	return r.dbWithContext(ctx).Create(notification).Error
}

func (r *NotificationRepository) CreateIfSourceEventAbsent(ctx context.Context, notification *opsentity.Notification) (bool, error) {
	if notification == nil {
		return false, nil
	}
	if notification.SourceEventKey == "" {
		return true, r.Create(ctx, notification)
	}
	result := r.dbWithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "source_event_key"}},
		TargetWhere: clause.Where{Exprs: []clause.Expression{
			clause.Expr{SQL: "source_event_key <> ''"},
		}},
		DoNothing: true,
	}).Create(notification)
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}

func (r *NotificationRepository) CreateBatch(ctx context.Context, batch *opsentity.NotificationBatch, notifications []*opsentity.Notification) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (r *NotificationRepository) WithinNotificationOutboxTx(ctx context.Context, fn func(txRepo opsports.NotificationOutboxTxRepository) error) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.WithDB(tx))
	})
}

func (r *NotificationRepository) EnqueueOutboxEvent(ctx context.Context, event platformevents.OutboxEvent) error {
	return platformevents.NewOutboxRepository(r.dbWithContext(ctx)).Enqueue(ctx, event)
}

func (r *NotificationRepository) List(ctx context.Context, filter opsports.NotificationListFilter) ([]opsentity.Notification, int64, error) {
	query := r.dbWithContext(ctx).Model(&opsentity.Notification{}).Where("user_id = ?", filter.UserID)
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	items := make([]opsentity.Notification, 0)
	if err := query.Order("created_at DESC").Offset(filter.Offset).Limit(filter.Limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *NotificationRepository) FindByID(ctx context.Context, notificationID, userID int64) (*opsentity.Notification, error) {
	var notification opsentity.Notification
	if err := r.dbWithContext(ctx).
		Where("id = ? AND user_id = ?", notificationID, userID).
		First(&notification).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, opsports.ErrNotificationNotFound
		}
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
	return r.dbWithContext(ctx).
		Model(&opsentity.Notification{}).
		Where("id = ? AND user_id = ?", notificationID, userID).
		Updates(map[string]any{"is_read": true, "read_at": readAt}).Error
}

func (r *NotificationRepository) listUserIDs(ctx context.Context, apply func(query *gorm.DB) *gorm.DB) ([]int64, error) {
	query := r.dbWithContext(ctx).Model(&identitycontracts.User{}).Where("deleted_at IS NULL")
	if apply != nil {
		query = apply(query)
	}
	userIDs := make([]int64, 0)
	if err := query.Order("id ASC").Pluck("id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}
