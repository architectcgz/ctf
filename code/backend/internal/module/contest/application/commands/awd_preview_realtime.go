package commands

import (
	"context"
	"strings"
	"time"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	platformevents "ctf-platform/internal/platform/events"
)

type awdPreviewRequesterContextKey struct{}

func WithAWDPreviewRequester(ctx context.Context, userID int64) context.Context {
	if ctx == nil {
		return nil
	}
	if userID <= 0 {
		return ctx
	}
	return context.WithValue(ctx, awdPreviewRequesterContextKey{}, userID)
}

func awdPreviewRequesterFromContext(ctx context.Context) (int64, bool) {
	if ctx == nil {
		return 0, false
	}
	userID, ok := ctx.Value(awdPreviewRequesterContextKey{}).(int64)
	if !ok || userID <= 0 {
		return 0, false
	}
	return userID, true
}

func broadcastAWDPreviewProgress(
	ctx context.Context,
	bus platformevents.Bus,
	contestID int64,
	requestID string,
	phaseKey string,
	phaseLabel string,
	detail string,
	attempt int,
	totalAttempts int,
	status string,
	extra map[string]any,
) {
	if bus == nil {
		return
	}
	userID, ok := awdPreviewRequesterFromContext(ctx)
	if !ok {
		return
	}
	event := contestcontracts.AWDPreviewProgressEvent{
		UserID:           userID,
		ContestID:        contestID,
		PreviewRequestID: strings.TrimSpace(requestID),
		PhaseKey:         strings.TrimSpace(phaseKey),
		PhaseLabel:       strings.TrimSpace(phaseLabel),
		Detail:           strings.TrimSpace(detail),
		Attempt:          attempt,
		TotalAttempts:    totalAttempts,
		Status:           strings.TrimSpace(status),
		OccurredAt:       contestEventTimestamp(time.Now().UTC()),
	}
	if extraError, ok := extra["error"].(string); ok {
		event.Error = strings.TrimSpace(extraError)
	}

	publishContestWeakEvent(ctx, bus, platformevents.Event{
		Name:    contestcontracts.EventAWDPreviewProgress,
		Payload: event,
	})
}
