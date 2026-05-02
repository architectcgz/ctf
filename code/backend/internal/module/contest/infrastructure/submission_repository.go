package infrastructure

import (
	"context"

	contestports "ctf-platform/internal/module/contest/ports"
	"gorm.io/gorm"
)

type SubmissionRepository struct {
	db *gorm.DB
}

func NewSubmissionRepository(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) WithDB(db *gorm.DB) *SubmissionRepository {
	return &SubmissionRepository{db: db}
}

func (r *SubmissionRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *SubmissionRepository) WithinScoringTransaction(ctx context.Context, fn func(repo contestports.ContestSubmissionScoringTxRepository) error) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.WithDB(tx))
	})
}
