package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"
	"time"

	"ctf-platform/internal/apperror"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *AWDService) resolveMaterializedActiveRound(ctx context.Context, contest *contestentity.Contest, activeRoundNumber int, now time.Time) (*contestentity.AWDRound, error) {
	round, err := s.findRoundByNumber(ctx, contest.ID, activeRoundNumber)
	if err == nil {
		return round, nil
	}
	if !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if err := s.ensureActiveRoundMaterialized(ctx, contest, now); err != nil {
		return nil, err
	}

	round, err = s.findRoundByNumber(ctx, contest.ID, activeRoundNumber)
	if err == nil {
		return round, nil
	}
	if errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		return nil, contestcontracts.ErrAWDRoundNotActive
	}
	return nil, apperror.ErrInternal.WithCause(err)
}

func (s *AWDService) ensureActiveRoundMaterialized(ctx context.Context, contest *contestentity.Contest, now time.Time) error {
	if contest == nil {
		return contestcontracts.ErrContestNotFound
	}
	if s.roundManager == nil {
		return apperror.ErrInternal.WithCause(errors.New("awd round manager is nil"))
	}
	if err := s.roundManager.EnsureActiveRoundMaterialized(ctx, contest, now); err != nil {
		if errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
			return contestcontracts.ErrAWDRoundNotActive
		}
		return apperror.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *AWDService) findRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*contestentity.AWDRound, error) {
	return s.repo.FindRoundByNumber(ctx, contestID, roundNumber)
}
