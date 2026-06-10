package infrastructure

import (
	"context"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestports "ctf-platform/internal/module/contest/ports"
	"gorm.io/gorm"
)

type ParticipationRepository struct {
	db *gorm.DB
}

func NewParticipationRepository(db *gorm.DB) *ParticipationRepository {
	return &ParticipationRepository{db: db}
}

func (r *ParticipationRepository) WithDB(db *gorm.DB) *ParticipationRepository {
	return &ParticipationRepository{db: db}
}

func (r *ParticipationRepository) dbWithContext(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx)
}

func (r *ParticipationRepository) WithinAnnouncementTransaction(ctx context.Context, fn func(repo contestports.ContestParticipationAnnouncementTxRepository) error) error {
	return r.dbWithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txRepo := &participationAnnouncementTxRepository{
			ParticipationRepository: r.WithDB(tx),
			outbox:                  NewRealtimeOutboxRepository(tx),
		}
		return fn(txRepo)
	})
}

type participationAnnouncementTxRepository struct {
	*ParticipationRepository
	outbox *RealtimeOutboxRepository
}

func (r *participationAnnouncementTxRepository) EnqueueRealtimeRelay(ctx context.Context, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) error {
	return r.outbox.EnqueueRealtimeRelay(ctx, relay, dedupeKey)
}
