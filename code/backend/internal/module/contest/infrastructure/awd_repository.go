package infrastructure

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type AWDRepository struct {
	db *gorm.DB
}

func NewAWDRepository(db *gorm.DB) *AWDRepository {
	return &AWDRepository{db: db}
}

func (r *AWDRepository) WithDB(db *gorm.DB) *AWDRepository {
	return &AWDRepository{db: db}
}

func (r *AWDRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *AWDRepository) WithinTransaction(ctx context.Context, fn func(txRepo contestports.AWDRepository) error) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.WithDB(tx))
	})
}

func (r *AWDRepository) CreateContestAWDService(ctx context.Context, service *model.ContestAWDService) error {
	return r.dbWithContext(ctx).Create(service).Error
}

func (r *AWDRepository) UpdateContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64, updates map[string]any) error {
	if len(updates) == 0 {
		return nil
	}
	return r.dbWithContext(ctx).Model(&model.ContestAWDService{}).
		Where("contest_id = ? AND id = ?", contestID, serviceID).
		Updates(updates).Error
}

func (r *AWDRepository) FindContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error) {
	var item model.ContestAWDService
	if err := r.dbWithContext(ctx).
		Where("contest_id = ? AND id = ?", contestID, serviceID).
		First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *AWDRepository) ListContestAWDServicesByContest(ctx context.Context, contestID int64) ([]model.ContestAWDService, error) {
	var items []model.ContestAWDService
	err := r.dbWithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("\"order\" ASC, id ASC").
		Find(&items).Error
	return items, err
}

func (r *AWDRepository) DeleteContestAWDServiceByContestAndID(ctx context.Context, contestID, serviceID int64) error {
	return r.dbWithContext(ctx).
		Where("contest_id = ? AND id = ?", contestID, serviceID).
		Delete(&model.ContestAWDService{}).Error
}

func (r *AWDRepository) RecalculateContestTeamScores(ctx context.Context, contestID int64) error {
	return RecalculateAWDContestTeamScores(ctx, r.db, contestID)
}

func (r *AWDRepository) RebuildContestScoreboardCache(ctx context.Context, redis *redislib.Client, contestID int64) error {
	return RebuildContestScoreboardCache(ctx, r.db, redis, contestID)
}
