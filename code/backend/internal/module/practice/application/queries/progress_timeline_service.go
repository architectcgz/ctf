package queries

import (
	"context"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/dto"
	practiceports "ctf-platform/internal/module/practice/ports"
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
