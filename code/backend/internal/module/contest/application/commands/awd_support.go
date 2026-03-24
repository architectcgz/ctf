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
	contestdomain "ctf-platform/internal/module/contest/domain"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) ensureAWDContest(ctx context.Context, contestID int64) (*model.Contest, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if contest.Mode != model.ContestModeAWD {
		return nil, errcode.ErrForbidden
	}
	return contest, nil
}

func (s *AWDService) ensureAWDRound(ctx context.Context, contestID, roundID int64) (*model.AWDRound, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	round, err := s.repo.FindRoundByContestAndID(ctx, contestID, roundID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return round, nil
}

func (s *AWDService) loadContestTeams(ctx context.Context, contestID int64) (map[int64]*model.Team, error) {
	teams, err := s.repo.FindTeamsByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	result := make(map[int64]*model.Team, len(teams))
	for _, team := range teams {
		result[team.ID] = team
	}
	return result, nil
}

func (s *AWDService) ensureContestChallenge(ctx context.Context, contestID, challengeID int64) error {
	ok, err := s.repo.ContestHasChallenge(ctx, contestID, challengeID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if !ok {
		return errcode.ErrChallengeNotInContest
	}
	return nil
}

func (s *AWDService) resolveUserTeamID(ctx context.Context, userID, contestID int64) (int64, error) {
	registration, err := s.repo.FindRegistration(ctx, contestID, userID)
	if err == nil {
		if err := contestdomain.RegistrationStatusError(registration.Status); err != nil {
			return 0, err
		}
		if registration.TeamID == nil || *registration.TeamID <= 0 {
			return 0, errcode.ErrAWDTeamRequired
		}
		return *registration.TeamID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, errcode.ErrInternal.WithCause(err)
	}

	team, err := s.repo.FindContestTeamByMember(ctx, contestID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errcode.ErrNotRegistered
		}
		return 0, errcode.ErrInternal.WithCause(err)
	}
	return team.ID, nil
}

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

func (s *AWDService) loadChallenge(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	challenge, err := s.repo.FindChallengeByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return challenge, nil
}

func (s *AWDService) resolveAcceptedRoundFlags(
	ctx context.Context,
	contestID int64,
	round *model.AWDRound,
	victimTeamID int64,
	challenge *model.Challenge,
	now time.Time,
) ([]string, error) {
	currentFlag, err := s.resolveRoundFlag(ctx, contestID, round, victimTeamID, challenge)
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
	previousFlag, err := s.resolveRoundFlag(ctx, contestID, previousRound, victimTeamID, challenge)
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

func (s *AWDService) resolveRoundFlag(ctx context.Context, contestID int64, round *model.AWDRound, victimTeamID int64, challenge *model.Challenge) (string, error) {
	if round == nil || challenge == nil {
		return "", errcode.ErrAWDFlagUnavailable
	}
	if s.redis != nil {
		flag, err := s.redis.HGet(ctx, rediskeys.AWDRoundFlagsKey(contestID, round.ID), rediskeys.AWDRoundFlagField(victimTeamID, challenge.ID)).Result()
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
	return contestdomain.BuildAWDRoundFlag(contestID, round.RoundNumber, victimTeamID, challenge.ID, s.flagSecret, challenge.FlagPrefix), nil
}
