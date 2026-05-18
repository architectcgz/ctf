package infrastructure

import (
	"context"

	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (i *dockerAWDFlagInjector) InjectRoundFlags(ctx context.Context, contest *contestentity.Contest, round *contestentity.AWDRound, assignments []contestports.AWDFlagAssignment) error {
	if i.db == nil || contest == nil || round == nil || len(assignments) == 0 {
		return nil
	}

	type pair struct {
		teamID         int64
		serviceID      int64
		awdChallengeID int64
	}
	seen := make(map[pair]struct{}, len(assignments))
	for _, item := range assignments {
		key := pair{teamID: item.TeamID, serviceID: item.ServiceID, awdChallengeID: item.AWDChallengeID}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = struct{}{}

		containerIDs, err := i.findTargetContainers(ctx, contest.ID, item.TeamID, item.ServiceID, item.AWDChallengeID)
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
