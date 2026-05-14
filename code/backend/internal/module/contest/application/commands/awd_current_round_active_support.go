package commands

import (
	"context"
	"errors"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) resolveMaterializedActiveRound(ctx context.Context, contest *model.Contest, activeRoundNumber int, now time.Time) (*model.AWDRound, error) {
	round, err := s.findRoundByNumber(ctx, contest.ID, activeRoundNumber)
	if err == nil {
		return round, nil
	}
	if !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if err := s.ensureActiveRoundMaterialized(ctx, contest, now); err != nil {
		return nil, err
	}

	round, err = s.findRoundByNumber(ctx, contest.ID, activeRoundNumber)
	if err == nil {
		return round, nil
	}
	if errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		return nil, errcode.ErrAWDRoundNotActive
	}
	return nil, errcode.ErrInternal.WithCause(err)
}

func (s *AWDService) ensureActiveRoundMaterialized(ctx context.Context, contest *model.Contest, now time.Time) error {
	if contest == nil {
		return errcode.ErrContestNotFound
	}
	if s.roundManager == nil {
		return errcode.ErrInternal.WithCause(errors.New("awd round manager is nil"))
	}
	if err := s.roundManager.EnsureActiveRoundMaterialized(ctx, contest, now); err != nil {
		if errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
			return errcode.ErrAWDRoundNotActive
		}
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *AWDService) findRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	return s.repo.FindRoundByNumber(ctx, contestID, roundNumber)
}
