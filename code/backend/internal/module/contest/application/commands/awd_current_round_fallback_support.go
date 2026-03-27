package commands

import (
	"context"
	"errors"
	"strconv"
	"strings"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) resolveCurrentRoundFromFallbacks(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	round, err := s.repo.FindRunningRound(ctx, contestID)
	if err == nil {
		return round, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	round, err = s.findCurrentRoundFromRedis(ctx, contestID)
	if err == nil {
		return round, nil
	}
	if err != nil {
		return nil, err
	}

	return nil, errcode.ErrAWDRoundNotActive
}

func (s *AWDService) findCurrentRoundFromRedis(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	if s.redis == nil {
		return nil, nil
	}

	roundNumberStr, err := s.redis.Get(ctx, rediskeys.AWDCurrentRoundKey(contestID)).Result()
	if err != nil {
		if errors.Is(err, redislib.Nil) {
			return nil, nil
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	roundNumber, convErr := strconv.Atoi(strings.TrimSpace(roundNumberStr))
	if convErr != nil || roundNumber <= 0 {
		return nil, nil
	}

	round, findErr := s.repo.FindRoundByNumber(ctx, contestID, roundNumber)
	if findErr == nil {
		return round, nil
	}
	if errors.Is(findErr, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return nil, errcode.ErrInternal.WithCause(findErr)
}
