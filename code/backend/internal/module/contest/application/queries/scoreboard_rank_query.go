package queries

import (
	"context"

	redislib "github.com/redis/go-redis/v9"

	contestdomain "ctf-platform/internal/module/contest/domain"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

func (s *ScoreboardService) GetTeamRank(ctx context.Context, contestID, teamID int64) (*TeamRankResult, error) {
	key := rediskeys.RankContestTeamKey(contestID)
	score, err := s.redis.ZScore(ctx, key, contestdomain.TeamIDToMember(teamID)).Result()
	if err != nil {
		if err == redislib.Nil {
			return &TeamRankResult{TeamID: teamID, Rank: 0, Score: 0}, nil
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	rank, err := s.redis.ZRevRank(ctx, key, contestdomain.TeamIDToMember(teamID)).Result()
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &TeamRankResult{
		TeamID: teamID,
		Rank:   int(rank) + 1,
		Score:  score,
	}, nil
}
