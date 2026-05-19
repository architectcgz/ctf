package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func (s *AWDService) resolveCurrentRound(ctx context.Context, contestID int64) (*contestentity.AWDRound, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	return s.resolveCurrentRoundForContest(ctx, contest)
}

func (s *AWDService) resolveCurrentRoundForContest(ctx context.Context, contest *contestentity.Contest) (*contestentity.AWDRound, error) {
	if contest == nil {
		return nil, contestcontracts.ErrContestNotFound
	}

	now := time.Now().UTC()
	if activeRoundNumber, ok := s.calculateActiveRoundNumber(contest, now); ok {
		return s.resolveMaterializedActiveRound(ctx, contest, activeRoundNumber, now)
	}

	return s.resolveCurrentRoundFromFallbacks(ctx, contest.ID)
}
