package commands

import (
	"context"

	contestports "ctf-platform/internal/module/contest/ports"
)

func syncAWDServiceStatusField(
	ctx context.Context,
	store contestports.AWDRoundStateStore,
	contestID, roundID, currentRoundID, teamID, serviceID int64,
	serviceStatus string,
) error {
	if store == nil || contestID <= 0 || roundID <= 0 || currentRoundID <= 0 || roundID != currentRoundID {
		return nil
	}
	return store.SetAWDServiceStatus(ctx, contestID, teamID, serviceID, serviceStatus)
}
