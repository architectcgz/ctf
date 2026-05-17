package ports_test

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type ctxOnlyPracticeScoreLockLease struct{}

func (ctxOnlyPracticeScoreLockLease) Key(context.Context) string { return "practice:score:lock" }

func (ctxOnlyPracticeScoreLockLease) Release(context.Context) (bool, error) { return true, nil }

var _ practiceports.PracticeScoreLockLease = (*ctxOnlyPracticeScoreLockLease)(nil)

type ctxOnlyPracticeScoreStateStore struct{}

func (ctxOnlyPracticeScoreStateStore) AcquireUserScoreUpdateLock(context.Context, int64, time.Duration) (practiceports.PracticeScoreLockLease, bool, error) {
	return ctxOnlyPracticeScoreLockLease{}, true, nil
}

func (ctxOnlyPracticeScoreStateStore) LoadUserScoreCache(context.Context, int64) (*dto.UserScoreInfo, bool, error) {
	return nil, false, nil
}

func (ctxOnlyPracticeScoreStateStore) StoreUserScoreCache(context.Context, *dto.UserScoreInfo, time.Duration) error {
	return nil
}

func (ctxOnlyPracticeScoreStateStore) SyncUserScoreState(context.Context, *dto.UserScoreInfo, time.Duration) error {
	return nil
}

var _ practiceports.PracticeScoreStateStore = (*ctxOnlyPracticeScoreStateStore)(nil)
