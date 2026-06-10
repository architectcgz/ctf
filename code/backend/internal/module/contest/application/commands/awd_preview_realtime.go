package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
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

func enqueueAWDPreviewProgressRealtime(
	ctx context.Context,
	outbox contestcontractsRealtimeOutbox,
	contestID int64,
	requestID string,
	phaseKey string,
	phaseLabel string,
	detail string,
	attempt int,
	totalAttempts int,
	status string,
	extra map[string]any,
) error {
	if outbox == nil {
		return nil
	}
	userID, ok := awdPreviewRequesterFromContext(ctx)
	if !ok {
		return nil
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
	return outbox.EnqueueRealtimeRelay(
		ctx,
		contestcontracts.RelayAWDPreviewProgress(event),
		awdPreviewProgressDedupeKey(contestID, userID, event.PreviewRequestID, event.PhaseKey, event.Attempt, event.Status),
	)
}

type contestcontractsRealtimeOutbox interface {
	EnqueueRealtimeRelay(ctx context.Context, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) error
}

func awdPreviewProgressDedupeKey(contestID, userID int64, requestID, phaseKey string, attempt int, status string) string {
	return fmt.Sprintf(
		"contest:%d:awd_preview:%d:%s:%s:%d:%s",
		contestID,
		userID,
		strings.TrimSpace(requestID),
		strings.TrimSpace(phaseKey),
		attempt,
		strings.TrimSpace(status),
	)
}
