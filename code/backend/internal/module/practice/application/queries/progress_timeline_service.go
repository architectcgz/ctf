package queries

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type progressTimelineQueryRepository interface {
	practiceports.PracticeProgressQueryRepository
	practiceports.PracticeTimelineQueryRepository
}

type ProgressTimelineQueryService interface {
	GetProgress(ctx context.Context, userID int64) (*practiceports.UserProgressSnapshot, error)
	GetTimeline(ctx context.Context, userID int64, limit, offset int) (*practiceports.TimelineSnapshot, error)
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

func (s *ProgressTimelineService) HandleFlagAcceptedOutboxEvent(ctx context.Context, event platformevents.OutboxEvent) error {
	if s == nil || s.cache == nil {
		return nil
	}

	codec := platformevents.NewOutboxCodec()
	codec.Register(practicecontracts.EventFlagAccepted, practicecontracts.EventFlagAcceptedPayloadVersion, func() any {
		return &practicecontracts.FlagAcceptedEvent{}
	})
	decoded, err := codec.Decode(event)
	if err != nil {
		return err
	}
	payload, ok := decoded.Payload.(*practicecontracts.FlagAcceptedEvent)
	if !ok {
		return fmt.Errorf("unexpected practice flag event payload: %T", decoded.Payload)
	}
	if payload.UserID <= 0 {
		return nil
	}
	return s.cache.DeleteUserProgress(ctx, payload.UserID)
}
