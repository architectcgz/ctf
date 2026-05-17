package ports_test

import (
	"context"
	"time"

	practiceports "ctf-platform/internal/module/practice/ports"
)

type ctxOnlyPracticeProgressQueryRepository struct{}

func (ctxOnlyPracticeProgressQueryRepository) GetUserProgress(context.Context, int64) (int, int, error) {
	return 0, 0, nil
}

func (ctxOnlyPracticeProgressQueryRepository) GetUserRank(context.Context, int64) (int, error) {
	return 0, nil
}

func (ctxOnlyPracticeProgressQueryRepository) GetCategoryStats(context.Context, int64) ([]practiceports.CategoryProgressStat, error) {
	return nil, nil
}

func (ctxOnlyPracticeProgressQueryRepository) GetDifficultyStats(context.Context, int64) ([]practiceports.DifficultyProgressStat, error) {
	return nil, nil
}

func (ctxOnlyPracticeProgressQueryRepository) GetUserTimeline(context.Context, int64, int, int) ([]practiceports.TimelineEventRecord, error) {
	return nil, nil
}

var _ practiceports.PracticeProgressQueryRepository = (*ctxOnlyPracticeProgressQueryRepository)(nil)
var _ practiceports.PracticeTimelineQueryRepository = (*ctxOnlyPracticeProgressQueryRepository)(nil)

type ctxOnlyPracticeProgressCache struct{}

func (ctxOnlyPracticeProgressCache) GetUserProgress(context.Context, int64) (*practiceports.UserProgressSnapshot, bool, error) {
	return nil, false, nil
}

func (ctxOnlyPracticeProgressCache) StoreUserProgress(context.Context, int64, *practiceports.UserProgressSnapshot, time.Duration) error {
	return nil
}

func (ctxOnlyPracticeProgressCache) DeleteUserProgress(context.Context, int64) error {
	return nil
}

var _ practiceports.PracticeUserProgressCache = (*ctxOnlyPracticeProgressCache)(nil)
