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

func TestContestRealtimeServicePublishesScoreboardRelay(t *testing.T) {
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

	if len(publisher.relays) != 1 {
		t.Fatalf("expected 1 relay, got %+v", publisher.relays)
	}
	if publisher.relays[0].Channel != contestcontracts.ScoreboardChannel(88) || publisher.relays[0].MessageType != "scoreboard.updated" {
		t.Fatalf("unexpected scoreboard relay: %+v", publisher.relays[0])
	}
}
