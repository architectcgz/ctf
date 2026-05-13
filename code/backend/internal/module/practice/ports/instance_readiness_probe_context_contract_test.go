package ports_test

import (
	"context"
	"time"

	practiceports "ctf-platform/internal/module/practice/ports"
)

type ctxOnlyPracticeInstanceReadinessProbe struct{}

func (ctxOnlyPracticeInstanceReadinessProbe) ProbeAccessURL(context.Context, string, time.Duration) error {
	return nil
}

var _ practiceports.PracticeInstanceReadinessProbe = (*ctxOnlyPracticeInstanceReadinessProbe)(nil)
