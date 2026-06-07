package commands

import (
	"context"
	"fmt"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	platformevents "ctf-platform/internal/platform/events"
)

const awdPreviewProgressMessageType = "awd.preview.progress"

type contestRealtimeRelayPublisher interface {
	Publish(ctx context.Context, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) (string, error)
}

type ContestRealtimeService struct {
	publisher contestRealtimeRelayPublisher
}

func NewContestRealtimeService(publisher contestRealtimeRelayPublisher) *ContestRealtimeService {
	return &ContestRealtimeService{publisher: publisher}
}

func (s *ContestRealtimeService) RegisterContestEventConsumers(bus platformevents.Bus) {
	if s == nil || s.publisher == nil || bus == nil {
		return
	}
	bus.Subscribe(contestcontracts.EventAnnouncementCreated, s.handleAnnouncementCreated)
	bus.Subscribe(contestcontracts.EventAnnouncementDeleted, s.handleAnnouncementDeleted)
	bus.Subscribe(contestcontracts.EventScoreboardUpdated, s.handleScoreboardUpdated)
	bus.Subscribe(contestcontracts.EventAWDPreviewProgress, s.handleAWDPreviewProgress)
}

func (s *ContestRealtimeService) handleAnnouncementCreated(ctx context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(contestcontracts.AnnouncementCreatedEvent)
	if !ok {
		return fmt.Errorf("unexpected contest announcement created payload: %T", evt.Payload)
	}
	_, err := s.publisher.Publish(ctx, contestcontracts.RelayAnnouncementCreated(payload), "")
	return err
}

func (s *ContestRealtimeService) handleAnnouncementDeleted(ctx context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(contestcontracts.AnnouncementDeletedEvent)
	if !ok {
		return fmt.Errorf("unexpected contest announcement deleted payload: %T", evt.Payload)
	}
	_, err := s.publisher.Publish(ctx, contestcontracts.RelayAnnouncementDeleted(payload), "")
	return err
}

func (s *ContestRealtimeService) handleScoreboardUpdated(ctx context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(contestcontracts.ScoreboardUpdatedEvent)
	if !ok {
		return fmt.Errorf("unexpected contest scoreboard updated payload: %T", evt.Payload)
	}
	_, err := s.publisher.Publish(ctx, contestcontracts.RelayScoreboardUpdated(payload), "")
	return err
}

func (s *ContestRealtimeService) handleAWDPreviewProgress(ctx context.Context, evt platformevents.Event) error {
	payload, ok := evt.Payload.(contestcontracts.AWDPreviewProgressEvent)
	if !ok {
		return fmt.Errorf("unexpected contest awd preview progress payload: %T", evt.Payload)
	}
	_, err := s.publisher.Publish(ctx, contestcontracts.RelayAWDPreviewProgress(payload), "")
	return err
}
