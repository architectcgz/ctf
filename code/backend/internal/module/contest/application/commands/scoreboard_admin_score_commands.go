package commands

import (
	"context"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/module/contest/domain"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

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
