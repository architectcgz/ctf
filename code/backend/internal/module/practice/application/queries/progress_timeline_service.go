package queries

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type progressTimelineQueryRepository interface {
	practiceports.PracticeProgressQueryRepository
	practiceports.PracticeTimelineQueryRepository
}

type ProgressTimelineQueryService interface {
	GetProgress(ctx context.Context, userID int64) (*dto.ProgressResp, error)
	GetTimeline(ctx context.Context, userID int64, limit, offset int) (*dto.TimelineResp, error)
}

type ProgressTimelineService struct {
	repo     progressTimelineQueryRepository
	cache    practiceports.PracticeUserProgressCache
	cacheTTL time.Duration
	logger   *zap.Logger
}

var _ ProgressTimelineQueryService = (*ProgressTimelineService)(nil)

func NewProgressTimelineService(
	repo progressTimelineQueryRepository,
	cache practiceports.PracticeUserProgressCache,
	cacheTTL time.Duration,
	logger *zap.Logger,
) *ProgressTimelineService {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &ProgressTimelineService{
		repo:     repo,
		cache:    cache,
		cacheTTL: cacheTTL,
		logger:   logger,
	}
}

func (s *ProgressTimelineService) RegisterPracticeEventConsumers(bus platformevents.Bus) {
	if s == nil || bus == nil {
		return
	}
	bus.Subscribe(practicecontracts.EventFlagAccepted, s.handleFlagAcceptedEvent)
}

func (s *ProgressTimelineService) handleFlagAcceptedEvent(ctx context.Context, evt platformevents.Event) error {
	if s == nil || s.cache == nil {
		return nil
	}

	payload, ok := evt.Payload.(practicecontracts.FlagAcceptedEvent)
	if !ok {
		return fmt.Errorf("unexpected practice flag event payload: %T", evt.Payload)
	}
	if payload.UserID <= 0 {
		return nil
	}
	return s.cache.DeleteUserProgress(ctx, payload.UserID)
}
