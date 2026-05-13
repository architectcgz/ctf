package ports_test

import (
	"context"
	"time"

	practiceports "ctf-platform/internal/module/practice/ports"
)

type ctxOnlyPracticeFlagSubmitRateLimitStore struct{}

func (ctxOnlyPracticeFlagSubmitRateLimitStore) AllowFlagSubmit(context.Context, int64, int64, int, time.Duration) (bool, error) {
	return true, nil
}

var _ practiceports.PracticeFlagSubmitRateLimitStore = (*ctxOnlyPracticeFlagSubmitRateLimitStore)(nil)
