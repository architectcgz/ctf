package commands

import (
	"context"
	"errors"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) resolveCurrentRoundFromFallbacks(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	round, err := s.repo.FindRunningRound(ctx, contestID)
	if err == nil {
		return round, nil
	}
	if !errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	round, err = s.findCurrentRoundFromRedis(ctx, contestID)
	if err == nil && round != nil {
		return round, nil
	}
	if err != nil {
		return nil, err
	}

	return nil, errcode.ErrAWDRoundNotActive
}

func (s *AWDService) findCurrentRoundFromRedis(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	roundNumber, ok, err := s.stateStore.LoadAWDCurrentRoundNumber(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if !ok {
		return nil, nil
	}

	round, findErr := s.repo.FindRoundByNumber(ctx, contestID, roundNumber)
	if findErr == nil {
		return round, nil
	}
	if errors.Is(findErr, contestports.ErrContestAWDRoundNotFound) {
		return nil, nil
	}
	return nil, errcode.ErrInternal.WithCause(findErr)
}
