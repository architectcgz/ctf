package commands

import (
	"context"
	"fmt"
	"time"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
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

func scoreboardUpdatedRelay(contestID int64, occurredAt time.Time) contestcontracts.RealtimeRelayEvent {
	return contestcontracts.RelayScoreboardUpdated(contestcontracts.ScoreboardUpdatedEvent{
		ContestID:  contestID,
		OccurredAt: contestEventTimestamp(occurredAt),
	})
}

func scoreboardSubmissionDedupeKey(contestID, submissionID int64) string {
	return fmt.Sprintf("contest:%d:submission:%d:scoreboard_updated", contestID, submissionID)
}

func scoreboardOperationDedupeKey(contestID int64, operation string, occurredAt time.Time) string {
	return fmt.Sprintf("contest:%d:scoreboard:%s:%s", contestID, operation, contestEventTimestamp(occurredAt).Format(time.RFC3339Nano))
}
