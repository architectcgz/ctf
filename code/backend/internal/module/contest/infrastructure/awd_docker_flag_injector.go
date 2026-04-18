package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (i *dockerAWDFlagInjector) InjectRoundFlags(ctx context.Context, contest *model.Contest, round *model.AWDRound, assignments []contestports.AWDFlagAssignment) error {
	if i.db == nil || contest == nil || round == nil || len(assignments) == 0 {
		return nil
	}

	type pair struct {
		teamID      int64
		serviceID   int64
		challengeID int64
	}
	seen := make(map[pair]struct{}, len(assignments))
	for _, item := range assignments {
		key := pair{teamID: item.TeamID, serviceID: item.ServiceID, challengeID: item.ChallengeID}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		containerIDs, err := i.findTargetContainers(ctx, contest.ID, item.TeamID, item.ServiceID, item.ChallengeID)
		if err != nil {
			return err
		}
		for _, containerID := range containerIDs {
			if err := i.writer.WriteFileToContainer(ctx, containerID, i.flagFilePath, []byte(item.Flag)); err != nil {
				return err
			}
		}
	}

	return nil
}
