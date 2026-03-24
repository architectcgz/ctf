package commands

import (
	"context"

	redislib "github.com/redis/go-redis/v9"

	rediskeys "ctf-platform/internal/pkg/redis"
)

func syncAWDServiceStatusField(
	ctx context.Context,
	redis *redislib.Client,
	contestID, roundID, currentRoundID, teamID, challengeID int64,
	serviceStatus string,
) error {
	if redis == nil || contestID <= 0 || roundID <= 0 || currentRoundID <= 0 || roundID != currentRoundID {
		return nil
	}
	return redis.HSet(
		ctx,
		rediskeys.AWDServiceStatusKey(contestID),
		rediskeys.AWDRoundFlagField(teamID, challengeID),
		serviceStatus,
	).Err()
}
