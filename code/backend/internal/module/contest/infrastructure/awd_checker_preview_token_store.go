package infrastructure

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	redislib "github.com/redis/go-redis/v9"

	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

var _ contestports.AWDCheckerPreviewTokenStore = (*AWDCheckerPreviewTokenStore)(nil)

type AWDCheckerPreviewTokenStore struct {
	cache *redislib.Client
}

func NewAWDCheckerPreviewTokenStore(cache *redislib.Client) *AWDCheckerPreviewTokenStore {
	return &AWDCheckerPreviewTokenStore{cache: cache}
}

func (s *AWDCheckerPreviewTokenStore) StoreAWDCheckerPreviewToken(ctx context.Context, record contestports.AWDCheckerPreviewTokenRecord, ttl time.Duration) (string, error) {
	if s == nil || s.cache == nil {
		return "", contestports.ErrAWDCheckerPreviewTokenStoreUnavailable
	}
	if record.ContestID <= 0 {
		return "", nil
	}

	token := uuid.NewString()
	raw, err := json.Marshal(record)
	if err != nil {
		return "", err
	}
	if err := s.cache.Set(ctx, rediskeys.AWDCheckerPreviewTokenKey(record.ContestID, token), raw, ttl).Err(); err != nil {
		return "", err
	}
	return token, nil
}

func (s *AWDCheckerPreviewTokenStore) LoadAWDCheckerPreviewToken(ctx context.Context, contestID int64, token string) (*contestports.AWDCheckerPreviewTokenRecord, bool, error) {
	if s == nil || s.cache == nil {
		return nil, false, contestports.ErrAWDCheckerPreviewTokenStoreUnavailable
	}
	if contestID <= 0 || strings.TrimSpace(token) == "" {
		return nil, false, nil
	}
	raw, err := s.cache.Get(ctx, rediskeys.AWDCheckerPreviewTokenKey(contestID, token)).Result()
	if err != nil {
		if err == redislib.Nil {
			return nil, false, nil
		}
		return nil, false, err
	}
	var record contestports.AWDCheckerPreviewTokenRecord
	if err := json.Unmarshal([]byte(raw), &record); err != nil {
		return nil, false, err
	}
	return &record, true, nil
}

func (s *AWDCheckerPreviewTokenStore) DeleteAWDCheckerPreviewToken(ctx context.Context, contestID int64, token string) error {
	if s == nil || s.cache == nil {
		return contestports.ErrAWDCheckerPreviewTokenStoreUnavailable
	}
	if contestID <= 0 || strings.TrimSpace(token) == "" {
		return nil
	}
	return s.cache.Del(ctx, rediskeys.AWDCheckerPreviewTokenKey(contestID, token)).Err()
}
