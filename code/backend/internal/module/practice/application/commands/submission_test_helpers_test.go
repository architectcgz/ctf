package commands

import (
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	"github.com/redis/go-redis/v9"
)

func newPracticeFlagSubmitRateLimitStoreForTest(redisClient *redis.Client) practiceports.PracticeFlagSubmitRateLimitStore {
	return practiceinfra.NewFlagSubmitRateLimitStore(redisClient, "practice:test")
}
