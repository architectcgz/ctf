package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) resolveAcceptedRoundFlags(
	ctx context.Context,
	contestID int64,
	round *model.AWDRound,
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

	if !s.allowPreviousRoundFlag(round, now) {
		return flags, nil
	}

	previousRound, err := s.findRoundByNumber(ctx, contestID, round.RoundNumber-1)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return flags, nil
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	previousFlag, err := s.resolveRoundFlag(ctx, contestID, previousRound, victimTeamID, awdChallengeID, flagPrefix, serviceID)
	if err != nil {
		if err == errcode.ErrAWDFlagUnavailable {
			return flags, nil
		}
		return nil, err
	}
	return append(flags, previousFlag), nil
}

func (s *AWDService) allowPreviousRoundFlag(round *model.AWDRound, now time.Time) bool {
	if round == nil || round.RoundNumber <= 1 || s.awdConfig.PreviousRoundGrace <= 0 || round.StartedAt == nil {
		return false
	}
	return now.Before(round.StartedAt.Add(s.awdConfig.PreviousRoundGrace))
}

func (s *AWDService) resolveRoundFlag(ctx context.Context, contestID int64, round *model.AWDRound, victimTeamID int64, awdChallengeID int64, flagPrefix string, serviceID int64) (string, error) {
	if round == nil || awdChallengeID <= 0 {
		return "", errcode.ErrAWDFlagUnavailable
	}
	if s.redis != nil {
		flag, err := s.redis.HGet(ctx, rediskeys.AWDRoundFlagsKey(contestID, round.ID), rediskeys.AWDRoundFlagServiceField(victimTeamID, serviceID)).Result()
		if err == nil && strings.TrimSpace(flag) != "" {
			return flag, nil
		}
		if err != nil && !errors.Is(err, redislib.Nil) {
			return "", errcode.ErrInternal.WithCause(err)
		}
	}
	if strings.TrimSpace(s.flagSecret) == "" {
		return "", errcode.ErrAWDFlagUnavailable
	}
	return contestdomain.BuildAWDRoundFlag(contestID, round.RoundNumber, victimTeamID, awdChallengeID, s.flagSecret, flagPrefix), nil
}
