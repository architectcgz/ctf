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

func (s *AWDService) calculateActiveRoundNumber(contest *model.Contest, now time.Time) (int, bool) {
	if contest == nil || s.awdConfig.RoundInterval <= 0 {
		return 0, false
	}
	if !contest.EndTime.After(contest.StartTime) {
		return 0, false
	}
	if now.Before(contest.StartTime) || !now.Before(contest.EndTime) {
		return 0, false
	}

	duration := contest.EndTime.Sub(contest.StartTime)
	totalRounds := int((duration + s.awdConfig.RoundInterval - 1) / s.awdConfig.RoundInterval)
	if totalRounds <= 0 {
		return 0, false
	}

	activeRound := int(now.Sub(contest.StartTime)/s.awdConfig.RoundInterval) + 1
	if activeRound > totalRounds {
		activeRound = totalRounds
	}
	return activeRound, true
}

func (s *AWDService) resolveCurrentRoundID(ctx context.Context, contestID int64) (int64, error) {
	if !s.isLiveContestWindow(ctx, contestID) {
		return 0, nil
	}
	round, err := s.resolveCurrentRound(ctx, contestID)
	if err != nil {
		if err == errcode.ErrAWDRoundNotActive {
			return 0, nil
		}
		return 0, err
	}
	return round.ID, nil
}

func (s *AWDService) isLiveContestWindow(ctx context.Context, contestID int64) bool {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil || contest == nil {
		return false
	}
	now := time.Now()
	if !now.Before(contest.EndTime) {
		return false
	}
	return contest.Status == model.ContestStatusRunning || contest.Status == model.ContestStatusFrozen
}

func (s *AWDService) findRoundByNumber(ctx context.Context, contestID int64, roundNumber int) (*model.AWDRound, error) {
	return s.repo.FindRoundByNumber(ctx, contestID, roundNumber)
}
