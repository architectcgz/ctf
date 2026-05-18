package jobs

import (
	"context"

	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) loadContestServiceDefinitions(ctx context.Context, contestID int64) ([]contestports.AWDServiceDefinition, error) {
	return u.repo.ListServiceDefinitionsByContest(ctx, contestID)
}

func effectiveAWDDefenseScore(definition contestports.AWDServiceDefinition, round *contestentity.AWDRound) int {
	if definition.DefenseScore > 0 {
		return definition.DefenseScore
	}
	if round == nil {
		return 0
	}
	return round.DefenseScore
}
