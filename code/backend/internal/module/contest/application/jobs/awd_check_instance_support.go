package jobs

import (
	"context"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) loadContestServiceInstances(ctx context.Context, contestID int64, challenges []model.Challenge) ([]contestports.AWDServiceInstance, error) {
	if len(challenges) == 0 {
		return nil, nil
	}

	challengeIDs := make([]int64, 0, len(challenges))
	for _, challenge := range challenges {
		challengeIDs = append(challengeIDs, challenge.ID)
	}

	return u.repo.ListServiceInstancesByContest(ctx, contestID, challengeIDs)
}
