package commands

import (
	"context"
	"errors"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) ensureAWDContest(ctx context.Context, contestID int64) (*model.Contest, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest.Mode != model.ContestModeAWD {
		return nil, errcode.ErrForbidden
	}
	return contest, nil
}

func (s *AWDService) ensureAWDRound(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	round, err := s.repo.FindRoundByContestAndID(ctx, contestID, roundID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return round, nil
}
