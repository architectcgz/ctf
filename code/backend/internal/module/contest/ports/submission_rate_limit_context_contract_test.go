package ports_test

import (
	"context"
	"time"

	contestports "ctf-platform/internal/module/contest/ports"
)

type ctxOnlyContestSubmissionRateLimitStore struct{}

func (ctxOnlyContestSubmissionRateLimitStore) HasIncorrectSubmissionRateLimit(context.Context, int64, int64, int64) (bool, error) {
	return false, nil
}

func (ctxOnlyContestSubmissionRateLimitStore) SetIncorrectSubmissionRateLimit(context.Context, int64, int64, int64, time.Duration) error {
	return nil
}

var _ contestports.ContestSubmissionRateLimitStore = (*ctxOnlyContestSubmissionRateLimitStore)(nil)
