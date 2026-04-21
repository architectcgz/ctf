package commands

import (
	"context"
	"strings"

	contestports "ctf-platform/internal/module/contest/ports"
	ctfws "ctf-platform/pkg/websocket"
)

const awdPreviewProgressEventType = "awd.preview.progress"

type awdPreviewRequesterContextKey struct{}

func WithAWDPreviewRequester(ctx context.Context, userID int64) context.Context {
	if ctx == nil {
		ctx = context.Background()
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
	broadcaster contestports.RealtimeBroadcaster,
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
	if broadcaster == nil {
		return
	}
	userID, ok := awdPreviewRequesterFromContext(ctx)
	if !ok {
		return
	}
	payload := map[string]any{
		"contest_id":         contestID,
		"preview_request_id": strings.TrimSpace(requestID),
		"phase_key":          strings.TrimSpace(phaseKey),
		"phase_label":        strings.TrimSpace(phaseLabel),
		"detail":             strings.TrimSpace(detail),
		"status":             strings.TrimSpace(status),
	}
	if attempt > 0 {
		payload["attempt"] = attempt
	}
	if totalAttempts > 0 {
		payload["total_attempts"] = totalAttempts
	}
	for key, value := range extra {
		payload[key] = value
	}

	broadcaster.SendToUser(userID, ctfws.Envelope{
		Type:    awdPreviewProgressEventType,
		Payload: payload,
	})
}
