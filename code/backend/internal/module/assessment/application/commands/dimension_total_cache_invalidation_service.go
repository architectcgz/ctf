package commands

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	assessmentports "ctf-platform/internal/module/assessment/ports"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	platformevents "ctf-platform/internal/platform/events"
)

type DimensionTotalCacheInvalidationService struct {
	store  assessmentports.AssessmentDimensionTotalCacheStore
	logger *zap.Logger
}

func NewDimensionTotalCacheInvalidationService(
	store assessmentports.AssessmentDimensionTotalCacheStore,
	logger *zap.Logger,
) *DimensionTotalCacheInvalidationService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &DimensionTotalCacheInvalidationService{
		store:  store,
		logger: logger,
	}
}

func (s *DimensionTotalCacheInvalidationService) RegisterChallengeEventConsumers(bus platformevents.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe(challengecontracts.EventPublishedCatalogChanged, s.handlePublishedCatalogChangedEvent)
}

func (s *DimensionTotalCacheInvalidationService) handlePublishedCatalogChangedEvent(ctx context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(challengecontracts.PublishedCatalogChangedEvent)
	if !ok {
		return fmt.Errorf("unexpected challenge published catalog payload: %T", evt.Payload)
	}
	if s == nil || s.store == nil {
		return nil
	}
	if payload.ChallengeID <= 0 {
		return nil
	}

	if err := s.store.DeletePublishedDimensionTotals(ctx); err != nil {
		return err
	}
	s.logger.Debug(
		"assessment_dimension_total_cache_invalidated",
		zap.Int64("challenge_id", payload.ChallengeID),
		zap.String("change_type", payload.ChangeType),
	)
	return nil
}
