package queries

import (
	"context"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/pkg/errcode"
)

func (s *ScoreboardService) createSnapshotFromLive(ctx context.Context, contestID int64) error {
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

func teamName(team *model.Team) string {
	if team == nil {
		return ""
	}
	return team.Name
}
