package commands

import (
	"context"
	"errors"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	"ctf-platform/pkg/errcode"
)

func (s *ChallengeService) ensureMutableContest(ctx context.Context, contestID int64) (*contestentity.Contest, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contestdomain.IsContestImmutable(contest) {
		return nil, errcode.ErrContestImmutable
	}
	return contest, nil
}
