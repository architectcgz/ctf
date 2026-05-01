package infrastructure

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	rediskeys "ctf-platform/internal/pkg/redis"
)

type redisScoreboardCache interface {
	TxPipeline() redislib.Pipeliner
}

type ScoreboardCache struct {
	db    *gorm.DB
	redis redisScoreboardCache
}

func NewScoreboardCache(db *gorm.DB, redis *redislib.Client) *ScoreboardCache {
	if db == nil || redis == nil {
		return nil
	}
	return &ScoreboardCache{db: db, redis: redis}
}

func (c *ScoreboardCache) RebuildContestScoreboard(ctx context.Context, contestID int64) error {
	if c == nil {
		return nil
	}
	return RebuildContestScoreboardCache(ctx, c.db, c.redis, contestID)
}

func RebuildContestScoreboardCache(ctx context.Context, db *gorm.DB, redis redisScoreboardCache, contestID int64) error {
	if db == nil || redis == nil || contestID <= 0 {
		return nil
	}

	var teams []model.Team
	if err := db.WithContext(ctx).
		Where("contest_id = ?", contestID).
		Order("id ASC").
		Find(&teams).Error; err != nil {
		return err
	}

	key := rediskeys.RankContestTeamKey(contestID)
	pipe := redis.TxPipeline()
	pipe.Del(ctx, key)

	entries := make([]redislib.Z, 0, len(teams))
	for _, team := range teams {
		if team.TotalScore <= 0 {
			continue
		}
		entries = append(entries, redislib.Z{
			Score:  float64(team.TotalScore),
			Member: contestdomain.TeamIDToMember(team.ID),
		})
	}
	if len(entries) > 0 {
		pipe.ZAdd(ctx, key, entries...)
	}
	_, err := pipe.Exec(ctx)
	return err
}
