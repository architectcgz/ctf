package jobs

import (
	"context"
	"net/http"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) SetHTTPClient(client *http.Client) {
	if u == nil || client == nil {
		return
	}
	u.httpClient = client
}

func (u *AWDRoundUpdater) SetCheckerRunner(runner contestports.CheckerRunner) {
	if u == nil {
		return
	}
	u.checkerRunner = runner
}

func (u *AWDRoundUpdater) SyncRoundServiceChecks(ctx context.Context, contest *model.Contest, activeRound int) error {
	return u.syncRoundServiceChecks(ctx, contest, activeRound)
}
