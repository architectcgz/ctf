package jobs

import (
	"context"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) loadContestServiceDefinitions(ctx context.Context, contestID int64) ([]contestports.AWDServiceDefinition, error) {
	return u.repo.ListServiceDefinitionsByContest(ctx, contestID)
}

func effectiveAWDDefenseScore(definition contestports.AWDServiceDefinition, round *model.AWDRound) int {
	if definition.DefenseScore > 0 {
		return definition.DefenseScore
	}
	if round == nil {
		return 0
	}
	return round.DefenseScore
}
