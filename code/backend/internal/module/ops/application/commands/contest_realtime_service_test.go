package commands

import (
	"context"
	"testing"
	"time"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	platformevents "ctf-platform/internal/platform/events"
)

type stubContestRealtimeRelayPublisher struct {
	relays []contestcontracts.RealtimeRelayEvent
}

func (p *stubContestRealtimeRelayPublisher) Publish(_ context.Context, relay contestcontracts.RealtimeRelayEvent, _ string) (string, error) {
	p.relays = append(p.relays, relay)
	return "1-0", nil
}

func TestContestRealtimeServiceRegisterContestEventConsumers(t *testing.T) {
	publisher := &stubContestRealtimeRelayPublisher{}
	service := NewContestRealtimeService(publisher)
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

func TestContestRealtimeServicePublishesAnnouncementRelay(t *testing.T) {
	publisher := &stubContestRealtimeRelayPublisher{}
	service := NewContestRealtimeService(publisher)
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

	if len(publisher.relays) != 1 {
		t.Fatalf("expected 1 relay, got %+v", publisher.relays)
	}
	relay := publisher.relays[0]
	if relay.EventName != contestcontracts.EventAnnouncementCreated {
		t.Fatalf("unexpected relay event name: %+v", relay)
	}
	if relay.Delivery != contestcontracts.RealtimeDeliveryChannel || relay.Channel != contestcontracts.AnnouncementChannel(77) {
		t.Fatalf("unexpected relay delivery target: %+v", relay)
	}
	if relay.MessageType != "contest.announcement.created" {
		t.Fatalf("unexpected relay message type: %+v", relay)
	}
	payload, ok := relay.Payload.(contestcontracts.AnnouncementCreatedRelayPayload)
	if !ok {
		t.Fatalf("unexpected announcement payload: %#v", relay.Payload)
	}
	if payload.Announcement.ID != int64(501) || payload.Announcement.Title != "比赛开始" {
		t.Fatalf("unexpected announcement body: %#v", payload)
	}
}

func TestContestRealtimeServicePublishesScoreboardAndPreviewRelays(t *testing.T) {
	publisher := &stubContestRealtimeRelayPublisher{}
	service := NewContestRealtimeService(publisher)
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

	if len(publisher.relays) != 2 {
		t.Fatalf("expected 2 relays, got %+v", publisher.relays)
	}
	if publisher.relays[0].Channel != contestcontracts.ScoreboardChannel(88) || publisher.relays[0].MessageType != "scoreboard.updated" {
		t.Fatalf("unexpected scoreboard relay: %+v", publisher.relays[0])
	}
	if publisher.relays[1].RecipientUserID == nil || *publisher.relays[1].RecipientUserID != 9001 {
		t.Fatalf("unexpected preview relay recipient: %+v", publisher.relays[1])
	}
	if publisher.relays[1].MessageType != awdPreviewProgressMessageType {
		t.Fatalf("unexpected preview relay message type: %+v", publisher.relays[1])
	}
	if publisher.relays[1].Timestamp != previewAt {
		t.Fatalf("unexpected preview relay timestamp: %+v", publisher.relays[1])
	}
}
