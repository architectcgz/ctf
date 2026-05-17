package ports_test

import (
	"context"
	"time"

	contestports "ctf-platform/internal/module/contest/ports"
)

type ctxOnlyContestSchedulerLockLease struct{}

func (ctxOnlyContestSchedulerLockLease) Key(context.Context) string { return "" }

func (ctxOnlyContestSchedulerLockLease) Refresh(context.Context, time.Duration) (bool, error) {
	return true, nil
}

func (ctxOnlyContestSchedulerLockLease) Release(context.Context) (bool, error) {
	return true, nil
}

type ctxOnlyContestStatusUpdateLockStore struct{}

func (ctxOnlyContestStatusUpdateLockStore) AcquireStatusUpdateLock(context.Context, time.Duration) (contestports.ContestSchedulerLockLease, bool, error) {
	return ctxOnlyContestSchedulerLockLease{}, true, nil
}

var _ contestports.ContestSchedulerLockLease = (*ctxOnlyContestSchedulerLockLease)(nil)
var _ contestports.ContestStatusUpdateLockStore = (*ctxOnlyContestStatusUpdateLockStore)(nil)
