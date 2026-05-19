package ratelimit

import (
	"context"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"
)

const slidingWindowScript = `
local key = KEYS[1]
local now = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local limit = tonumber(ARGV[3])
local window_start = now - window
redis.call("ZREMRANGEBYSCORE", key, 0, window_start)
local current = redis.call("ZCARD", key)
if current >= limit then
  local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
  local reset_at = now + window
  if oldest[2] then
    reset_at = tonumber(oldest[2]) + window
  end
  return {0, limit, limit - current, reset_at}
end
redis.call("ZADD", key, now, tostring(now) .. "-" .. redis.call("INCR", key .. ":seq"))
redis.call("PEXPIRE", key, window)
redis.call("PEXPIRE", key .. ":seq", window)
local next_count = current + 1
return {1, limit, limit - next_count, now + window}
`

type Checker struct {
	cache     *redislib.Client
	keyPrefix string
}

type Result struct {
	Allowed    bool
	Limit      int
	Remaining  int
	ResetAt    time.Time
	RetryAfter time.Duration
}

func NewChecker(cache *redislib.Client, keyPrefix string) *Checker {
	return &Checker{
		cache:     cache,
		keyPrefix: keyPrefix,
	}
}

func (c *Checker) CheckRate(ctx context.Context, key string, limit int, window time.Duration) (*Result, error) {
	now := time.Now()
	values, err := c.cache.Eval(
		ctx,
		slidingWindowScript,
		[]string{c.redisKey(key)},
		now.UnixMilli(),
		window.Milliseconds(),
		limit,
	).Result()
	if err != nil {
		return nil, err
	}

	array, ok := values.([]any)
	if !ok || len(array) != 4 {
		return nil, fmt.Errorf("unexpected rate limit result")
	}

	allowed := asInt64(array[0]) == 1
	resetAt := time.UnixMilli(asInt64(array[3]))
	result := &Result{
		Allowed:   allowed,
		Limit:     int(asInt64(array[1])),
		Remaining: maxInt(int(asInt64(array[2])), 0),
		ResetAt:   resetAt,
	}
	if !allowed {
		result.RetryAfter = time.Until(resetAt)
		if result.RetryAfter < 0 {
			result.RetryAfter = 0
		}
	}

	return result, nil
}

func (c *Checker) redisKey(key string) string {
	return fmt.Sprintf("%s:%s", c.keyPrefix, key)
}

func asInt64(value any) int64 {
	switch converted := value.(type) {
	case int64:
		return converted
	case int:
		return int64(converted)
	default:
		return 0
	}
}

func maxInt(left, right int) int {
	if left > right {
		return left
	}
	return right
}
