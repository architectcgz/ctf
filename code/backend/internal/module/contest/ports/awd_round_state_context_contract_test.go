package ports_test

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type ctxOnlyAWDRoundStateStore struct{}

func (ctxOnlyAWDRoundStateStore) AcquireAWDSchedulerLock(context.Context, time.Duration) (contestports.ContestSchedulerLockLease, bool, error) {
	return ctxOnlyContestSchedulerLockLease{}, true, nil
}

func (ctxOnlyAWDRoundStateStore) TryAcquireAWDRoundLock(context.Context, int64, int, time.Duration) (bool, error) {
	return true, nil
}

func (ctxOnlyAWDRoundStateStore) IsAWDCurrentRound(context.Context, int64, int) (bool, error) {
	return false, nil
}

func (ctxOnlyAWDRoundStateStore) LoadAWDCurrentRoundNumber(context.Context, int64) (int, bool, error) {
	return 0, false, nil
}

func (ctxOnlyAWDRoundStateStore) LoadAWDRoundFlag(context.Context, int64, int64, int64, int64, int64) (string, bool, error) {
	return "", false, nil
}

func (ctxOnlyAWDRoundStateStore) SyncAWDCurrentRoundState(context.Context, int64, *model.AWDRound, []contestports.AWDFlagAssignment, time.Duration) error {
	return nil
}

func (ctxOnlyAWDRoundStateStore) ClearAWDCurrentRoundState(context.Context, int64) error {
	return nil
}

func (ctxOnlyAWDRoundStateStore) SetAWDServiceStatus(context.Context, int64, int64, int64, string) error {
	return nil
}

func (ctxOnlyAWDRoundStateStore) ReplaceAWDServiceStatus(context.Context, int64, []contestports.AWDServiceStatusEntry) error {
	return nil
}

func (ctxOnlyAWDRoundStateStore) ClearAWDServiceStatus(context.Context, int64) error {
	return nil
}

var _ contestports.AWDRoundStateStore = (*ctxOnlyAWDRoundStateStore)(nil)
