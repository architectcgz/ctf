package commands

import (
	"context"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"
	"strings"
	"time"

	"ctf-platform/internal/apperror"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *AWDService) resolveAcceptedRoundFlags(
	ctx context.Context,
	contest *contestentity.Contest,
	contestID int64,
	round *contestentity.AWDRound,
	victimTeamID int64,
	awdChallengeID int64,
	flagPrefix string,
	serviceID int64,
	now time.Time,
) ([]string, error) {
	currentFlag, err := s.resolveRoundFlag(ctx, contestID, round, victimTeamID, awdChallengeID, flagPrefix, serviceID)
	if err != nil {
		return nil, err
	}
	flags := []string{currentFlag}

	if !s.allowPreviousRoundFlag(contest, round, now) {
		return flags, nil
	}

	previousRound, err := s.findRoundByNumber(ctx, contestID, round.RoundNumber-1)
	if err != nil {
		if errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
			return flags, nil
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	previousFlag, err := s.resolveRoundFlag(ctx, contestID, previousRound, victimTeamID, awdChallengeID, flagPrefix, serviceID)
	if err != nil {
		if err == contestcontracts.ErrAWDFlagUnavailable {
			return flags, nil
		}
		return nil, err
	}
	return append(flags, previousFlag), nil
}

func (s *AWDService) allowPreviousRoundFlag(contest *contestentity.Contest, round *contestentity.AWDRound, now time.Time) bool {
	if round == nil || round.RoundNumber <= 1 || s.awdConfig.PreviousRoundGrace <= 0 || round.StartedAt == nil {
		return false
	}
	effectiveNow := contestdomain.ContestEffectiveNow(contest, now)
	return effectiveNow.Before(round.StartedAt.Add(s.awdConfig.PreviousRoundGrace))
}

func (s *AWDService) resolveRoundFlag(ctx context.Context, contestID int64, round *contestentity.AWDRound, victimTeamID int64, awdChallengeID int64, flagPrefix string, serviceID int64) (string, error) {
	if round == nil || awdChallengeID <= 0 {
		return "", contestcontracts.ErrAWDFlagUnavailable
	}
	flag, ok, err := s.stateStore.LoadAWDRoundFlag(ctx, contestID, round.ID, victimTeamID, serviceID)
	if err != nil {
		return "", apperror.ErrInternal.WithCause(err)
	}
	if ok {
		return flag, nil
	}
	if strings.TrimSpace(s.flagSecret) == "" {
		return "", contestcontracts.ErrAWDFlagUnavailable
	}
	return contestdomain.BuildAWDRoundFlag(contestID, round.RoundNumber, victimTeamID, awdChallengeID, s.flagSecret, flagPrefix), nil
}
