package infrastructure

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

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
	if ctx == nil {
		ctx = context.Background()
	}
	return r.db.WithContext(ctx)
}

func (r *AWDRepository) WithinTransaction(ctx context.Context, fn func(txRepo contestports.AWDRepository) error) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.WithDB(tx))
	})
}

func (r *AWDRepository) RecalculateContestTeamScores(ctx context.Context, contestID int64) error {
	return RecalculateAWDContestTeamScores(ctx, r.db, contestID)
}

func (r *AWDRepository) RebuildContestScoreboardCache(ctx context.Context, redis *redislib.Client, contestID int64) error {
	return RebuildContestScoreboardCache(ctx, r.db, redis, contestID)
}
