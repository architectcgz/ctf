package infrastructure

import (
	"context"
	"strconv"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"

	practiceports "ctf-platform/internal/module/practice/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
)

const desiredAWDReconcileStateRetention = 24 * time.Hour

type DesiredAWDReconcileStateStore struct {
	cache *redislib.Client
}

func NewDesiredAWDReconcileStateStore(cache *redislib.Client) *DesiredAWDReconcileStateStore {
	if cache == nil {
		return nil
	}
	return &DesiredAWDReconcileStateStore{cache: cache}
}

func (s *DesiredAWDReconcileStateStore) LoadDesiredAWDReconcileState(ctx context.Context, contestID, teamID, serviceID int64) (*practiceports.DesiredAWDReconcileState, bool, error) {
	if s == nil || s.cache == nil || contestID <= 0 || teamID <= 0 || serviceID <= 0 {
		return nil, false, nil
	}
	values, err := s.cache.HGetAll(ctx, rediskeys.DesiredAWDReconcileStateKey(contestID, teamID, serviceID)).Result()
	if err != nil {
		return nil, false, err
	}
	if len(values) == 0 {
		return nil, false, nil
	}

	state := &practiceports.DesiredAWDReconcileState{
		LastError: strings.TrimSpace(values["last_error"]),
	}
	if raw := strings.TrimSpace(values["failure_count"]); raw != "" {
		count, err := strconv.Atoi(raw)
		if err != nil {
			return nil, false, err
		}
		state.FailureCount = count
	}

	lastFailureAt, err := parseDesiredAWDReconcileStateTime(values["last_failure_at"])
	if err != nil {
		return nil, false, err
	}
	state.LastFailureAt = lastFailureAt

	nextAttemptAt, err := parseDesiredAWDReconcileStateTime(values["next_attempt_at"])
	if err != nil {
		return nil, false, err
	}
	state.NextAttemptAt = nextAttemptAt

	suppressedUntil, err := parseDesiredAWDReconcileStateTime(values["suppressed_until"])
	if err != nil {
		return nil, false, err
	}
	state.SuppressedUntil = suppressedUntil

	return state, true, nil
}

func (s *DesiredAWDReconcileStateStore) StoreDesiredAWDReconcileState(ctx context.Context, contestID, teamID, serviceID int64, state *practiceports.DesiredAWDReconcileState) error {
	if s == nil || s.cache == nil || contestID <= 0 || teamID <= 0 || serviceID <= 0 || state == nil {
		return nil
	}
	key := rediskeys.DesiredAWDReconcileStateKey(contestID, teamID, serviceID)
	pipe := s.cache.TxPipeline()
	pipe.HSet(ctx, key, map[string]any{
		"failure_count":    state.FailureCount,
		"last_failure_at":  formatDesiredAWDReconcileStateTime(state.LastFailureAt),
		"next_attempt_at":  formatDesiredAWDReconcileStateTime(state.NextAttemptAt),
		"suppressed_until": formatDesiredAWDReconcileStateTime(state.SuppressedUntil),
		"last_error":       strings.TrimSpace(state.LastError),
	})
	pipe.Expire(ctx, key, desiredAWDReconcileStateTTL(state, time.Now().UTC()))
	_, err := pipe.Exec(ctx)
	return err
}

func (s *DesiredAWDReconcileStateStore) DeleteDesiredAWDReconcileState(ctx context.Context, contestID, teamID, serviceID int64) error {
	if s == nil || s.cache == nil || contestID <= 0 || teamID <= 0 || serviceID <= 0 {
		return nil
	}
	return s.cache.Del(ctx, rediskeys.DesiredAWDReconcileStateKey(contestID, teamID, serviceID)).Err()
}

func parseDesiredAWDReconcileStateTime(raw string) (time.Time, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return time.Time{}, nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, err
	}
	return parsed.UTC(), nil
}

func formatDesiredAWDReconcileStateTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func desiredAWDReconcileStateTTL(state *practiceports.DesiredAWDReconcileState, now time.Time) time.Duration {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	expiryAt := now
	if state != nil {
		if state.LastFailureAt.After(expiryAt) {
			expiryAt = state.LastFailureAt
		}
		if state.NextAttemptAt.After(expiryAt) {
			expiryAt = state.NextAttemptAt
		}
		if state.SuppressedUntil.After(expiryAt) {
			expiryAt = state.SuppressedUntil
		}
	}
	ttl := expiryAt.Sub(now) + desiredAWDReconcileStateRetention
	if ttl < desiredAWDReconcileStateRetention {
		return desiredAWDReconcileStateRetention
	}
	return ttl
}
