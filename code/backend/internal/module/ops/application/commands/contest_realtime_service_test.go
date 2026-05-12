package commands

import (
	"context"
	"testing"
	"time"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	platformevents "ctf-platform/internal/platform/events"
	ctfws "ctf-platform/pkg/websocket"
)

type stubContestRealtimeBroadcaster struct {
	channels []string
	users    []int64
	messages []ctfws.Envelope
}

func (b *stubContestRealtimeBroadcaster) SendToChannel(channel string, message ctfws.Envelope) int {
	b.channels = append(b.channels, channel)
	b.messages = append(b.messages, message)
	return 1
}

func (b *stubContestRealtimeBroadcaster) SendToUser(userID int64, message ctfws.Envelope) int {
	b.users = append(b.users, userID)
	b.messages = append(b.messages, message)
	return 1
}

func TestContestRealtimeServiceRegisterContestEventConsumers(t *testing.T) {
	broadcaster := &stubContestRealtimeBroadcaster{}
	service := NewContestRealtimeService(broadcaster)
	bus := &recordingBus{}

	service.RegisterContestEventConsumers(bus)

	expected := []string{
		contestcontracts.EventAnnouncementCreated,
		contestcontracts.EventAnnouncementDeleted,
		contestcontracts.EventScoreboardUpdated,
		contestcontracts.EventAWDPreviewProgress,
	}
	for _, eventName := range expected {
		if got := len(bus.subscribers[eventName]); got != 1 {
			t.Fatalf("%s subscribers = %d, want 1", eventName, got)
		}
	}
	if got := len(bus.subscribers); got != len(expected) {
		t.Fatalf("subscriber count = %d, want %d", got, len(expected))
	}
}

func TestContestRealtimeServiceRelayAnnouncementCreated(t *testing.T) {
	broadcaster := &stubContestRealtimeBroadcaster{}
	service := NewContestRealtimeService(broadcaster)
	bus := &recordingBus{}
	service.RegisterContestEventConsumers(bus)

	occurredAt := time.Date(2026, 5, 12, 3, 4, 5, 0, time.UTC)
	err := bus.Publish(context.Background(), platformevents.Event{
		Name: contestcontracts.EventAnnouncementCreated,
		Payload: contestcontracts.AnnouncementCreatedEvent{
			ContestID:      77,
			AnnouncementID: 501,
			Title:          "比赛开始",
			Content:        "欢迎接入实时公告。",
			CreatedAt:      occurredAt,
			OccurredAt:     occurredAt,
		},
	})
	if err != nil {
		t.Fatalf("Publish(announcement_created) error = %v", err)
	}

	if len(broadcaster.channels) != 1 || broadcaster.channels[0] != "contest:77:announcements" {
		t.Fatalf("unexpected channels = %+v", broadcaster.channels)
	}
	if len(broadcaster.messages) != 1 || broadcaster.messages[0].Type != "contest.announcement.created" {
		t.Fatalf("unexpected messages = %+v", broadcaster.messages)
	}
	messagePayload, ok := broadcaster.messages[0].Payload.(map[string]any)
	if !ok {
		t.Fatalf("unexpected envelope payload = %#v", broadcaster.messages[0].Payload)
	}
	payload, ok := messagePayload["announcement"].(map[string]any)
	if !ok {
		t.Fatalf("unexpected announcement payload = %#v", messagePayload)
	}
	if payload["id"] != int64(501) || payload["title"] != "比赛开始" {
		t.Fatalf("unexpected announcement body = %#v", payload)
	}
	if !broadcaster.messages[0].Timestamp.Equal(occurredAt) {
		t.Fatalf("unexpected timestamp = %v", broadcaster.messages[0].Timestamp)
	}

	err = bus.Publish(context.Background(), platformevents.Event{
		Name: contestcontracts.EventAnnouncementDeleted,
		Payload: contestcontracts.AnnouncementDeletedEvent{
			ContestID:      77,
			AnnouncementID: 501,
			OccurredAt:     occurredAt.Add(time.Second),
		},
	})
	if err != nil {
		t.Fatalf("Publish(announcement_deleted) error = %v", err)
	}

	if len(broadcaster.channels) != 2 || broadcaster.channels[1] != "contest:77:announcements" {
		t.Fatalf("unexpected channels after delete = %+v", broadcaster.channels)
	}
	if len(broadcaster.messages) != 2 || broadcaster.messages[1].Type != "contest.announcement.deleted" {
		t.Fatalf("unexpected delete message = %+v", broadcaster.messages)
	}
}

func TestContestRealtimeServiceRelayScoreboardAndPreview(t *testing.T) {
	broadcaster := &stubContestRealtimeBroadcaster{}
	service := NewContestRealtimeService(broadcaster)
	bus := &recordingBus{}
	service.RegisterContestEventConsumers(bus)

	scoreboardAt := time.Date(2026, 5, 12, 3, 10, 0, 0, time.UTC)
	if err := bus.Publish(context.Background(), platformevents.Event{
		Name: contestcontracts.EventScoreboardUpdated,
		Payload: contestcontracts.ScoreboardUpdatedEvent{
			ContestID:  88,
			OccurredAt: scoreboardAt,
		},
	}); err != nil {
		t.Fatalf("Publish(scoreboard_updated) error = %v", err)
	}

	if len(broadcaster.channels) != 1 || broadcaster.channels[0] != "contest:88:scoreboard" {
		t.Fatalf("unexpected scoreboard channels = %+v", broadcaster.channels)
	}
	if broadcaster.messages[0].Type != "scoreboard.updated" {
		t.Fatalf("unexpected scoreboard message type = %s", broadcaster.messages[0].Type)
	}

	previewAt := time.Date(2026, 5, 12, 3, 11, 0, 0, time.UTC)
	if err := bus.Publish(context.Background(), platformevents.Event{
		Name: contestcontracts.EventAWDPreviewProgress,
		Payload: contestcontracts.AWDPreviewProgressEvent{
			UserID:           9001,
			ContestID:        88,
			PreviewRequestID: "preview-1",
			PhaseKey:         "attempt-1",
			PhaseLabel:       "第 1 轮试跑",
			Detail:           "正在执行第 1 / 3 轮请求校验。",
			Attempt:          1,
			TotalAttempts:    3,
			Status:           "running",
			OccurredAt:       previewAt,
		},
	}); err != nil {
		t.Fatalf("Publish(awd_preview_progress) error = %v", err)
	}

	if len(broadcaster.users) != 1 || broadcaster.users[0] != 9001 {
		t.Fatalf("unexpected preview users = %+v", broadcaster.users)
	}
	if len(broadcaster.messages) != 2 || broadcaster.messages[1].Type != awdPreviewProgressMessageType {
		t.Fatalf("unexpected preview message = %+v", broadcaster.messages)
	}
	previewPayload, ok := broadcaster.messages[1].Payload.(map[string]any)
	if !ok {
		t.Fatalf("unexpected preview envelope payload = %#v", broadcaster.messages[1].Payload)
	}
	if previewPayload["attempt"] != 1 || previewPayload["total_attempts"] != 3 {
		t.Fatalf("unexpected preview payload = %#v", previewPayload)
	}
	if !broadcaster.messages[1].Timestamp.Equal(previewAt) {
		t.Fatalf("unexpected preview timestamp = %v", broadcaster.messages[1].Timestamp)
	}
}
