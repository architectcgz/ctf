package jobs

import (
	"context"
	"net/http"

	"ctf-platform/internal/model"
)

func (u *AWDRoundUpdater) SetHTTPClient(client *http.Client) {
	if u == nil || client == nil {
		return
	}
	u.httpClient = client
}

func (u *AWDRoundUpdater) SyncRoundServiceChecks(ctx context.Context, contest *model.Contest, activeRound int) error {
	return u.syncRoundServiceChecks(ctx, contest, activeRound)
}
