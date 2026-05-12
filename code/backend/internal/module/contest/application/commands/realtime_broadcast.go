package commands

import (
	"context"
	"time"

	platformevents "ctf-platform/internal/platform/events"
)

func contestEventTimestamp(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now().UTC()
	}
	return ts.UTC()
}

func publishContestWeakEvent(ctx context.Context, bus platformevents.Bus, evt platformevents.Event) {
	if bus == nil {
		return
	}
	_ = bus.Publish(ctx, evt)
}
