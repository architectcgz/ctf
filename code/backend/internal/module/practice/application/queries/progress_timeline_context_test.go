package queries

import (
	"context"
	"testing"

	practiceports "ctf-platform/internal/module/practice/ports"
)

type stubPracticeProgressTimelineRepository struct {
	getUserProgressFn    func(ctx context.Context, userID int64) (int, int, error)
	getUserRankFn        func(ctx context.Context, userID int64) (int, error)
	getCategoryStatsFn   func(ctx context.Context, userID int64) ([]practiceports.CategoryProgressStat, error)
	getDifficultyStatsFn func(ctx context.Context, userID int64) ([]practiceports.DifficultyProgressStat, error)
	getUserTimelineFn    func(ctx context.Context, userID int64, limit, offset int) ([]practiceports.TimelineEventRecord, error)
}

func (s *stubPracticeProgressTimelineRepository) GetUserProgress(ctx context.Context, userID int64) (int, int, error) {
	if s.getUserProgressFn != nil {
		return s.getUserProgressFn(ctx, userID)
	}
	return 0, 0, nil
}

func (s *stubPracticeProgressTimelineRepository) GetUserRank(ctx context.Context, userID int64) (int, error) {
	if s.getUserRankFn != nil {
		return s.getUserRankFn(ctx, userID)
	}
	return 0, nil
}

func (s *stubPracticeProgressTimelineRepository) GetCategoryStats(ctx context.Context, userID int64) ([]practiceports.CategoryProgressStat, error) {
	if s.getCategoryStatsFn != nil {
		return s.getCategoryStatsFn(ctx, userID)
	}
	return nil, nil
}

func (s *stubPracticeProgressTimelineRepository) GetDifficultyStats(ctx context.Context, userID int64) ([]practiceports.DifficultyProgressStat, error) {
	if s.getDifficultyStatsFn != nil {
		return s.getDifficultyStatsFn(ctx, userID)
	}
	return nil, nil
}

func (s *stubPracticeProgressTimelineRepository) GetUserTimeline(ctx context.Context, userID int64, limit, offset int) ([]practiceports.TimelineEventRecord, error) {
	if s.getUserTimelineFn != nil {
		return s.getUserTimelineFn(ctx, userID, limit, offset)
	}
	return nil, nil
}

func TestGetProgressDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	service := NewProgressTimelineService(&stubPracticeProgressTimelineRepository{
		getUserProgressFn: func(ctx context.Context, userID int64) (int, int, error) {
			if ctx != nil {
				t.Fatalf("expected progress ctx to stay nil, got %v", ctx)
			}
			return 0, 0, nil
		},
		getUserRankFn: func(ctx context.Context, userID int64) (int, error) {
			if ctx != nil {
				t.Fatalf("expected rank ctx to stay nil, got %v", ctx)
			}
			return 0, nil
		},
		getCategoryStatsFn: func(ctx context.Context, userID int64) ([]practiceports.CategoryProgressStat, error) {
			if ctx != nil {
				t.Fatalf("expected category stats ctx to stay nil, got %v", ctx)
			}
			return nil, nil
		},
		getDifficultyStatsFn: func(ctx context.Context, userID int64) ([]practiceports.DifficultyProgressStat, error) {
			if ctx != nil {
				t.Fatalf("expected difficulty stats ctx to stay nil, got %v", ctx)
			}
			return nil, nil
		},
	}, nil, 0, nil)

	if _, err := service.GetProgress(nil, 7); err != nil {
		t.Fatalf("GetProgress() error = %v", err)
	}
}

func TestGetTimelineDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	service := NewProgressTimelineService(&stubPracticeProgressTimelineRepository{
		getUserTimelineFn: func(ctx context.Context, userID int64, limit, offset int) ([]practiceports.TimelineEventRecord, error) {
			if ctx != nil {
				t.Fatalf("expected timeline ctx to stay nil, got %v", ctx)
			}
			return nil, nil
		},
	}, nil, 0, nil)

	if _, err := service.GetTimeline(nil, 7, 20, 0); err != nil {
		t.Fatalf("GetTimeline() error = %v", err)
	}
}
