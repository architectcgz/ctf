package ports_test

import (
	"context"
	"time"

	contestports "ctf-platform/internal/module/contest/ports"
)

type ctxOnlyContestStatusUpdateLockLease struct{}

func (ctxOnlyContestStatusUpdateLockLease) Key() string { return "" }

func (ctxOnlyContestStatusUpdateLockLease) Refresh(context.Context, time.Duration) (bool, error) {
	return true, nil
}

func (ctxOnlyContestStatusUpdateLockLease) Release(context.Context) (bool, error) {
	return true, nil
}

type ctxOnlyContestStatusUpdateLockStore struct{}

func (ctxOnlyContestStatusUpdateLockStore) AcquireStatusUpdateLock(context.Context, time.Duration) (contestports.ContestStatusUpdateLockLease, bool, error) {
	return ctxOnlyContestStatusUpdateLockLease{}, true, nil
}

var _ contestports.ContestStatusUpdateLockLease = (*ctxOnlyContestStatusUpdateLockLease)(nil)
var _ contestports.ContestStatusUpdateLockStore = (*ctxOnlyContestStatusUpdateLockStore)(nil)
