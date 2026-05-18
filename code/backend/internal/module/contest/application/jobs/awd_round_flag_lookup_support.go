package jobs

import (
	"context"
	"errors"
	"strings"
	"time"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

var errAWDFlagUnavailable = errors.New("awd_flag_unavailable")

func (u *AWDRoundUpdater) resolveAcceptedRoundFlags(
	ctx context.Context,
	contest *contestentity.Contest,
	contestID int64,
	round *contestentity.AWDRound,
	teamID int64,
	definition contestports.AWDServiceDefinition,
	now time.Time,
) ([]string, error) {
	currentFlag, err := u.resolveRoundFlag(ctx, contestID, round, teamID, definition)
	if err != nil {
		return nil, err
	}
	flags := []string{currentFlag}

	if !u.allowPreviousRoundFlag(contest, round, now) {
		return flags, nil
	}

	previousRound, err := u.findRoundByNumber(ctx, contestID, round.RoundNumber-1)
	if err != nil {
		if errors.Is(err, contestports.ErrContestAWDRoundNotFound) {
			return flags, nil
		}
		return nil, err
	}
	previousFlag, err := u.resolveRoundFlag(ctx, contestID, previousRound, teamID, definition)
	if err != nil {
		if errors.Is(err, errAWDFlagUnavailable) {
			return flags, nil
		}
		return nil, err
	}
	return append(flags, previousFlag), nil
}

func (u *AWDRoundUpdater) allowPreviousRoundFlag(contest *contestentity.Contest, round *contestentity.AWDRound, now time.Time) bool {
	if round == nil || round.RoundNumber <= 1 || u.cfg.PreviousRoundGrace <= 0 || round.StartedAt == nil {
		return false
	}
	effectiveNow := contestdomain.ContestEffectiveNow(contest, now)
	return effectiveNow.Before(round.StartedAt.Add(u.cfg.PreviousRoundGrace))
}

func (u *AWDRoundUpdater) resolveRoundFlag(
	ctx context.Context,
	contestID int64,
	round *contestentity.AWDRound,
	teamID int64,
	definition contestports.AWDServiceDefinition,
) (string, error) {
	if round == nil {
		return "", errAWDFlagUnavailable
	}
	if u.stateStore != nil {
		flag, found, err := u.stateStore.LoadAWDRoundFlag(ctx, contestID, round.ID, teamID, definition.ServiceID)
		if err != nil {
			return "", err
		}
		if found && strings.TrimSpace(flag) != "" {
			return flag, nil
		}
	}
	if strings.TrimSpace(u.flagSecret) == "" {
		return "", errAWDFlagUnavailable
	}
	return contestdomain.BuildAWDRoundFlag(
		contestID,
		round.RoundNumber,
		teamID,
		definition.AWDChallengeID,
		u.flagSecret,
		definition.FlagPrefix,
	), nil
}
