package infrastructure

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type ProgressCache struct {
	client *redis.Client
}

var _ practiceports.PracticeUserProgressCache = (*ProgressCache)(nil)

func NewProgressCache(client *redis.Client) *ProgressCache {
	if client == nil {
		return nil
	}
	return &ProgressCache{client: client}
}

func (c *ProgressCache) GetUserProgress(ctx context.Context, userID int64) (*dto.ProgressResp, bool, error) {
	if c == nil || c.client == nil {
		return nil, false, nil
	}

	cached, err := c.client.Get(ctx, constants.UserProgressKey(userID)).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get user progress cache: %w", err)
	}

	var resp dto.ProgressResp
	if err := json.Unmarshal([]byte(cached), &resp); err != nil {
		return nil, false, fmt.Errorf("decode user progress cache: %w", err)
	}
	return &resp, true, nil
}

func (c *ProgressCache) StoreUserProgress(ctx context.Context, userID int64, resp *dto.ProgressResp, ttl time.Duration) error {
	if c == nil || c.client == nil || resp == nil {
		return nil
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("encode user progress cache: %w", err)
	}
	if err := c.client.Set(ctx, constants.UserProgressKey(userID), data, ttl).Err(); err != nil {
		return fmt.Errorf("store user progress cache: %w", err)
	}
	return nil
}
