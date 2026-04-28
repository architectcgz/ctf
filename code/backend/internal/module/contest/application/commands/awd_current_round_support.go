package commands

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) resolveCurrentRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	return s.resolveCurrentRoundForContest(ctx, contest)
}

func (s *AWDService) resolveCurrentRoundForContest(ctx context.Context, contest *model.Contest) (*model.AWDRound, error) {
	if contest == nil {
		return nil, errcode.ErrContestNotFound
	}

	now := time.Now().UTC()
	if activeRoundNumber, ok := s.calculateActiveRoundNumber(contest, now); ok {
		return s.resolveMaterializedActiveRound(ctx, contest, activeRoundNumber, now)
	}

	return s.resolveCurrentRoundFromFallbacks(ctx, contest.ID)
}
