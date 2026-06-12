package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
)

func NewClient(ctx context.Context, cfg config.RedisConfig) (*redislib.Client, error) {
	if ctx == nil {
		return nil, fmt.Errorf("redis client requires context")
	}
	mode := strings.ToLower(strings.TrimSpace(cfg.Mode))
	if mode == "" {
		mode = "single"
	}

	var (
		client *redislib.Client
	)
	switch mode {
	case "single":
		client = redislib.NewClient(buildSingleOptions(cfg))
	case "sentinel":
		client = redislib.NewFailoverClient(buildFailoverOptions(cfg))
	default:
		return nil, fmt.Errorf("unsupported redis mode %q", cfg.Mode)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := client.Ping(pingCtx).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}

	return client, nil
}

func buildSingleOptions(cfg config.RedisConfig) *redislib.Options {
	return &redislib.Options{
		Addr:         strings.TrimSpace(cfg.Addr),
		Password:     cfg.Password,
		DB:           cfg.DB,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}
}

func buildFailoverOptions(cfg config.RedisConfig) *redislib.FailoverOptions {
	addrs := make([]string, 0, len(cfg.SentinelAddrs))
	for _, addr := range cfg.SentinelAddrs {
		trimmed := strings.TrimSpace(addr)
		if trimmed == "" {
			continue
		}
		addrs = append(addrs, trimmed)
	}

	return &redislib.FailoverOptions{
		MasterName:       strings.TrimSpace(cfg.MasterName),
		SentinelAddrs:    addrs,
		SentinelUsername: cfg.SentinelUsername,
		SentinelPassword: cfg.SentinelPassword,
		Password:         cfg.Password,
		DB:               cfg.DB,
		DialTimeout:      cfg.DialTimeout,
		ReadTimeout:      cfg.ReadTimeout,
		WriteTimeout:     cfg.WriteTimeout,
	}
}
