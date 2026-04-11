package jobs

import (
	"context"

	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) loadContestServiceInstances(ctx context.Context, contestID int64, definitions []contestports.AWDServiceDefinition) ([]contestports.AWDServiceInstance, error) {
	if len(definitions) == 0 {
		return nil, nil
	}

	challengeIDs := make([]int64, 0, len(definitions))
	for _, definition := range definitions {
		challengeIDs = append(challengeIDs, definition.ChallengeID)
	}

	return u.repo.ListServiceInstancesByContest(ctx, contestID, challengeIDs)
}
