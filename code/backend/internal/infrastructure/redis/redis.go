package redis

import (
	"context"
	"fmt"
	"time"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
)

func NewClient(cfg config.RedisConfig) (*redislib.Client, error) {
	client := redislib.NewClient(&redislib.Options{
		Addr:         cfg.Addr,
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}
