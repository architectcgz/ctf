package jobs

import (
	"context"

	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) SetHTTPRuntime(runtime contestports.AWDHTTPRuntime) {
	if u == nil || runtime == nil {
		return
	}
	u.httpRuntime = runtime
}

func (u *AWDRoundUpdater) SetCheckerRunner(runner contestports.CheckerRunner) {
	if u == nil {
		return
	}
	u.checkerRunner = runner
}

func (u *AWDRoundUpdater) SyncRoundServiceChecks(ctx context.Context, contest *contestentity.Contest, activeRound int) error {
	return u.syncRoundServiceChecks(ctx, contest, activeRound)
}
