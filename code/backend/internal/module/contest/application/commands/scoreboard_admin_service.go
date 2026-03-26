package commands

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

type ScoreboardAdminService struct {
	repo  contestports.ContestScoreboardAdminRepository
	redis *redislib.Client
	cfg   *config.ContestConfig
}

func NewScoreboardAdminService(repo contestports.ContestScoreboardAdminRepository, redis *redislib.Client, cfg *config.ContestConfig) *ScoreboardAdminService {
	return &ScoreboardAdminService{repo: repo, redis: redis, cfg: cfg}
}

func (s *ScoreboardAdminService) UpdateScore(ctx context.Context, contestID, teamID int64, points float64) error {
	key := rediskeys.RankContestTeamKey(contestID)
	return s.redis.ZIncrBy(ctx, key, points, domain.TeamIDToMember(teamID)).Err()
}

func (s *ScoreboardAdminService) RebuildScoreboard(ctx context.Context, contestID int64) error {
	teams, err := s.repo.FindTeamsByContest(ctx, contestID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	key := rediskeys.RankContestTeamKey(contestID)
	pipe := s.redis.TxPipeline()
	pipe.Del(ctx, key)

	entries := make([]redislib.Z, 0, len(teams))
	for _, team := range teams {
		if team == nil || team.TotalScore <= 0 {
			continue
		}
		entries = append(entries, redislib.Z{
			Score:  float64(team.TotalScore),
			Member: domain.TeamIDToMember(team.ID),
		})
	}
	if len(entries) > 0 {
		pipe.ZAdd(ctx, key, entries...)
	}
	if _, err := pipe.Exec(ctx); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}

func (s *ScoreboardAdminService) CalculateDynamicScoreWithBase(baseScore float64, solveCount int64) int {
	if s.cfg == nil {
		return domain.CalculateDynamicScore(baseScore, 0, 0, solveCount)
	}
	if baseScore <= 0 {
		baseScore = s.cfg.BaseScore
	}
	return domain.CalculateDynamicScore(baseScore, s.cfg.MinScore, s.cfg.Decay, solveCount)
}

func (s *ScoreboardAdminService) FreezeScoreboard(ctx context.Context, contestID int64, minutesBeforeEnd int) error {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == domain.ErrContestNotFound {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	if !now.Before(contest.EndTime) {
		return errcode.ErrContestEnded
	}

	freezeTime := contest.EndTime.Add(-time.Duration(minutesBeforeEnd) * time.Minute)
	contest.FreezeTime = &freezeTime
	if !now.Before(freezeTime) {
		contest.Status = model.ContestStatusFrozen
		if err := s.createSnapshotFromLive(ctx, contestID); err != nil {
			return err
		}
	}

	return s.repo.Update(ctx, contest)
}

func (s *ScoreboardAdminService) UnfreezeScoreboard(ctx context.Context, contestID int64) error {
	contest, err := s.repo.FindByID(ctx, contestID)
	if err != nil {
		if err == domain.ErrContestNotFound {
			return errcode.ErrContestNotFound
		}
		return errcode.ErrInternal.WithCause(err)
	}

	if contest.FreezeTime == nil && contest.Status != model.ContestStatusFrozen {
		return errcode.ErrScoreboardNotFrozen
	}

	contest.FreezeTime = nil
	if contest.Status == model.ContestStatusFrozen && time.Now().Before(contest.EndTime) {
		contest.Status = model.ContestStatusRunning
	}
	if err := s.redis.Del(ctx, rediskeys.RankContestFrozenKey(contestID)).Err(); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	return s.repo.Update(ctx, contest)
}

func (s *ScoreboardAdminService) createSnapshotFromLive(ctx context.Context, contestID int64) error {
	srcKey := rediskeys.RankContestTeamKey(contestID)
	dstKey := rediskeys.RankContestFrozenKey(contestID)
	if err := s.redis.ZUnionStore(ctx, dstKey, &redislib.ZStore{
		Keys:    []string{srcKey},
		Weights: []float64{1},
	}).Err(); err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	return nil
}
