package redislock

import (
	"context"
	"time"

	"github.com/google/uuid"
	redislib "github.com/redis/go-redis/v9"
)

const releaseScript = `
if redis.call("get", KEYS[1]) == ARGV[1] then
	return redis.call("del", KEYS[1])
else
	return 0
end
`

type Lock struct {
	client *redislib.Client
	key    string
	token  string
}

func Acquire(ctx context.Context, client *redislib.Client, key string, ttl time.Duration) (*Lock, bool, error) {
	if client == nil || key == "" || ttl <= 0 {
		return nil, true, nil
	}

	lock := &Lock{
		client: client,
		key:    key,
		token:  uuid.NewString(),
	}

	acquired, err := client.SetNX(ctx, key, lock.token, ttl).Result()
	if err != nil {
		return nil, false, err
	}
	if !acquired {
		return nil, false, nil
	}
	return lock, true, nil
}

func (l *Lock) Release(ctx context.Context) (bool, error) {
	if l == nil || l.client == nil || l.key == "" || l.token == "" {
		return false, nil
	}

	result, err := l.client.Eval(ctx, releaseScript, []string{l.key}, l.token).Result()
	if err != nil {
		return false, err
	}

	released, ok := result.(int64)
	return ok && released > 0, nil
}

func (l *Lock) Key() string {
	if l == nil {
		return ""
	}
	return l.key
}
