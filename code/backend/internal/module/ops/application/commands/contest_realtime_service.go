package commands

import (
	"context"
	"fmt"
	"strings"
	"time"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	platformevents "ctf-platform/internal/platform/events"
	ctfws "ctf-platform/pkg/websocket"
)

const awdPreviewProgressMessageType = "awd.preview.progress"

type contestRealtimeBroadcaster interface {
	SendToChannel(channel string, message ctfws.Envelope) int
	SendToUser(userID int64, message ctfws.Envelope) int
}

type ContestRealtimeService struct {
	broadcaster contestRealtimeBroadcaster
}

func NewContestRealtimeService(broadcaster contestRealtimeBroadcaster) *ContestRealtimeService {
	return &ContestRealtimeService{broadcaster: broadcaster}
}

func (s *ContestRealtimeService) RegisterContestEventConsumers(bus platformevents.Bus) {
	if s == nil || s.broadcaster == nil || bus == nil {
		return
	}
	bus.Subscribe(contestcontracts.EventAnnouncementCreated, s.handleAnnouncementCreated)
	bus.Subscribe(contestcontracts.EventAnnouncementDeleted, s.handleAnnouncementDeleted)
	bus.Subscribe(contestcontracts.EventScoreboardUpdated, s.handleScoreboardUpdated)
	bus.Subscribe(contestcontracts.EventAWDPreviewProgress, s.handleAWDPreviewProgress)
}

func (s *ContestRealtimeService) handleAnnouncementCreated(_ context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(contestcontracts.AnnouncementCreatedEvent)
	if !ok {
		return fmt.Errorf("unexpected contest announcement created payload: %T", evt.Payload)
	}
	s.broadcaster.SendToChannel(contestcontracts.AnnouncementChannel(payload.ContestID), ctfws.Envelope{
		Type: "contest.announcement.created",
		Payload: map[string]any{
			"contest_id": payload.ContestID,
			"announcement": map[string]any{
				"id":         payload.AnnouncementID,
				"title":      payload.Title,
				"content":    payload.Content,
				"created_at": payload.CreatedAt,
			},
		},
		Timestamp: contestRealtimeTimestamp(payload.OccurredAt),
	})
	return nil
}

func (s *ContestRealtimeService) handleAnnouncementDeleted(_ context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(contestcontracts.AnnouncementDeletedEvent)
	if !ok {
		return fmt.Errorf("unexpected contest announcement deleted payload: %T", evt.Payload)
	}
	s.broadcaster.SendToChannel(contestcontracts.AnnouncementChannel(payload.ContestID), ctfws.Envelope{
		Type: "contest.announcement.deleted",
		Payload: map[string]any{
			"contest_id":      payload.ContestID,
			"announcement_id": payload.AnnouncementID,
		},
		Timestamp: contestRealtimeTimestamp(payload.OccurredAt),
	})
	return nil
}

func (s *ContestRealtimeService) handleScoreboardUpdated(_ context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(contestcontracts.ScoreboardUpdatedEvent)
	if !ok {
		return fmt.Errorf("unexpected contest scoreboard updated payload: %T", evt.Payload)
	}
	s.broadcaster.SendToChannel(contestcontracts.ScoreboardChannel(payload.ContestID), ctfws.Envelope{
		Type: "scoreboard.updated",
		Payload: map[string]any{
			"contest_id": payload.ContestID,
		},
		Timestamp: contestRealtimeTimestamp(payload.OccurredAt),
	})
	return nil
}

func (s *ContestRealtimeService) handleAWDPreviewProgress(_ context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(contestcontracts.AWDPreviewProgressEvent)
	if !ok {
		return fmt.Errorf("unexpected contest awd preview progress payload: %T", evt.Payload)
	}
	messagePayload := map[string]any{
		"contest_id":         payload.ContestID,
		"preview_request_id": strings.TrimSpace(payload.PreviewRequestID),
		"phase_key":          strings.TrimSpace(payload.PhaseKey),
		"phase_label":        strings.TrimSpace(payload.PhaseLabel),
		"detail":             strings.TrimSpace(payload.Detail),
		"status":             strings.TrimSpace(payload.Status),
	}
	if payload.Attempt > 0 {
		messagePayload["attempt"] = payload.Attempt
	}
	if payload.TotalAttempts > 0 {
		messagePayload["total_attempts"] = payload.TotalAttempts
	}
	if payload.Error != "" {
		messagePayload["error"] = strings.TrimSpace(payload.Error)
	}
	s.broadcaster.SendToUser(payload.UserID, ctfws.Envelope{
		Type:      awdPreviewProgressMessageType,
		Payload:   messagePayload,
		Timestamp: contestRealtimeTimestamp(payload.OccurredAt),
	})
	return nil
}

func contestRealtimeTimestamp(ts time.Time) time.Time {
	if ts.IsZero() {
		return time.Now().UTC()
	}
	return ts.UTC()
}
