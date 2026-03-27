package commands

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) resolveCurrentRound(ctx context.Context, contestID int64) (*model.AWDRound, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	return s.resolveCurrentRoundForContest(ctx, contest)
}

func (s *AWDService) resolveCurrentRoundForContest(ctx context.Context, contest *model.Contest) (*model.AWDRound, error) {
	if contest == nil {
		return nil, errcode.ErrContestNotFound
	}

	now := time.Now()
	if activeRoundNumber, ok := s.calculateActiveRoundNumber(contest, now); ok {
		round, err := s.findRoundByNumber(ctx, contest.ID, activeRoundNumber)
		if err == nil {
			return round, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		if err := s.ensureActiveRoundMaterialized(ctx, contest, now); err != nil {
			return nil, err
		}
		round, err = s.findRoundByNumber(ctx, contest.ID, activeRoundNumber)
		if err == nil {
			return round, nil
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrAWDRoundNotActive
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	round, err := s.repo.FindRunningRound(ctx, contest.ID)
	if err == nil {
		return round, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	if s.redis != nil {
		roundNumberStr, err := s.redis.Get(ctx, rediskeys.AWDCurrentRoundKey(contest.ID)).Result()
		if err == nil {
			roundNumber, convErr := strconv.Atoi(strings.TrimSpace(roundNumberStr))
			if convErr == nil && roundNumber > 0 {
				round, findErr := s.repo.FindRoundByNumber(ctx, contest.ID, roundNumber)
				if findErr == nil {
					return round, nil
				}
			}
		} else if !errors.Is(err, redislib.Nil) {
			return nil, errcode.ErrInternal.WithCause(err)
		}
	}

	return nil, errcode.ErrAWDRoundNotActive
}

func (s *AWDService) ensureActiveRoundMaterialized(ctx context.Context, contest *model.Contest, now time.Time) error {
	if contest == nil {
		return errcode.ErrContestNotFound
	}
	if s.roundManager == nil {
		return errcode.ErrInternal.WithCause(errors.New("awd round manager is nil"))
	}
	if err := s.roundManager.EnsureActiveRoundMaterialized(ctx, contest, now); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errcode.ErrAWDRoundNotActive
		}
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *AWDService) findRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	return s.repo.FindRoundByNumber(ctx, contestID, roundNumber)
}
