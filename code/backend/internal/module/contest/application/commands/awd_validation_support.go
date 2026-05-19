package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"

	"ctf-platform/internal/apperror"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *AWDService) ensureAWDContest(ctx context.Context, contestID int64) (*contestentity.Contest, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, contestcontracts.ErrContestNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if contest.Mode != contestentity.ContestModeAWD {
		return nil, apperror.ErrForbidden
	}
	return contest, nil
}

func (s *AWDService) ensureAWDRound(ctx context.Context, contestID, roundID int64) (*contestentity.AWDRound, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	round, err := s.repo.FindRoundByContestAndID(ctx, contestID, roundID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
			return nil, apperror.ErrNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	return round, nil
}
