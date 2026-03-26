package infrastructure

import (
	"context"

	"gorm.io/gorm"
)

type ParticipationRepository struct {
	db *gorm.DB
}

func NewParticipationRepository(db *gorm.DB) *ParticipationRepository {
	return &ParticipationRepository{db: db}
}

func (r *ParticipationRepository) dbWithContext(ctx context.Context) *gorm.DB {
	if ctx == nil {
		ctx = context.Background()
	}
	return r.db.WithContext(ctx)
}
