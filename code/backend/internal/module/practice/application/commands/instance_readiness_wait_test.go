package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"ctf-platform/internal/config"
)

type fastFailThenReadyProbe struct {
	readyAfter time.Duration
	firstCall  time.Time
	callTimes  []time.Time
}

func (p *fastFailThenReadyProbe) ProbeAccessURL(context.Context, string, time.Duration) error {
	now := time.Now()
	if p.firstCall.IsZero() {
		p.firstCall = now
	}
	p.callTimes = append(p.callTimes, now)
	if now.Sub(p.firstCall) < p.readyAfter {
		return errors.New("not ready")
	}
	return nil
}

func TestWaitForInstanceReadinessPreservesProbeBudgetForFastFailures(t *testing.T) {
	t.Parallel()

	probe := &fastFailThenReadyProbe{readyAfter: 25 * time.Millisecond}
	service := &Service{
		config: &config.Config{
			Container: config.ContainerConfig{
				StartProbeTimeout:  40 * time.Millisecond,
				StartProbeInterval: 10 * time.Millisecond,
				StartProbeAttempts: 2,
			},
		},
		readinessProbe: probe,
	}

	if err := service.waitForInstanceReadiness(context.Background(), "http://example.internal"); err != nil {
		t.Fatalf("waitForInstanceReadiness() error = %v", err)
	}
	if len(probe.callTimes) != 2 {
		t.Fatalf("expected 2 probe calls, got %d", len(probe.callTimes))
	}
	if gap := probe.callTimes[1].Sub(probe.callTimes[0]); gap < 30*time.Millisecond {
		t.Fatalf("expected second probe to preserve timeout budget, gap=%s", gap)
	}
}
