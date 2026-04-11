package jobs

import (
	"context"
	"errors"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

var errAWDFlagUnavailable = errors.New("awd_flag_unavailable")

func (u *AWDRoundUpdater) resolveAcceptedRoundFlags(
	ctx context.Context,
	contestID int64,
	round *model.AWDRound,
	teamID int64,
	definition contestports.AWDServiceDefinition,
	now time.Time,
) ([]string, error) {
	currentFlag, err := u.resolveRoundFlag(ctx, contestID, round, teamID, definition)
	if err != nil {
		return nil, err
	}
	flags := []string{currentFlag}

	if !u.allowPreviousRoundFlag(round, now) {
		return flags, nil
	}

	previousRound, err := u.findRoundByNumber(ctx, contestID, round.RoundNumber-1)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
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

func (u *AWDRoundUpdater) allowPreviousRoundFlag(round *model.AWDRound, now time.Time) bool {
	if round == nil || round.RoundNumber <= 1 || u.cfg.PreviousRoundGrace <= 0 || round.StartedAt == nil {
		return false
	}
	return now.Before(round.StartedAt.Add(u.cfg.PreviousRoundGrace))
}

func (u *AWDRoundUpdater) resolveRoundFlag(
	ctx context.Context,
	contestID int64,
	round *model.AWDRound,
	teamID int64,
	definition contestports.AWDServiceDefinition,
) (string, error) {
	if round == nil {
		return "", errAWDFlagUnavailable
	}
	if u.redis != nil {
		flag, err := u.redis.HGet(ctx, rediskeys.AWDRoundFlagsKey(contestID, round.ID), rediskeys.AWDRoundFlagField(teamID, definition.ChallengeID)).Result()
		if err == nil && strings.TrimSpace(flag) != "" {
			return flag, nil
		}
		if err != nil && !errors.Is(err, redislib.Nil) {
			return "", err
		}
	}
	if strings.TrimSpace(u.flagSecret) == "" {
		return "", errAWDFlagUnavailable
	}
	return contestdomain.BuildAWDRoundFlag(
		contestID,
		round.RoundNumber,
		teamID,
		definition.ChallengeID,
		u.flagSecret,
		definition.FlagPrefix,
	), nil
}
