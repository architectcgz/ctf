package commands

import (
	"context"
	"errors"
	"testing"
	"time"

	"go.uber.org/zap"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestports "ctf-platform/internal/module/contest/ports"
)

type stubContestRealtimeOutbox struct {
	pending            []contestports.ContestPendingRealtimeRelay
	listCalls          int
	sentIDs            []int64
	sentStreamMessages []string
	failedIDs          []int64
	failedReasons      []string
	markSentErr        error
	markFailedErr      error
}

func (s *stubContestRealtimeOutbox) EnqueueRealtimeRelay(context.Context, contestcontracts.RealtimeRelayEvent, string) error {
	return nil
}

func (s *stubContestRealtimeOutbox) ListPendingRealtimeRelays(context.Context, time.Time, int) ([]contestports.ContestPendingRealtimeRelay, error) {
	s.listCalls++
	return s.pending, nil
}

func (s *stubContestRealtimeOutbox) MarkRealtimeRelaySent(_ context.Context, outboxID int64, streamMessageID string, _ time.Time) error {
	s.sentIDs = append(s.sentIDs, outboxID)
	s.sentStreamMessages = append(s.sentStreamMessages, streamMessageID)
	return s.markSentErr
}

func (s *stubContestRealtimeOutbox) MarkRealtimeRelayFailed(_ context.Context, outboxID int64, reason error, _ time.Time) error {
	s.failedIDs = append(s.failedIDs, outboxID)
	if reason != nil {
		s.failedReasons = append(s.failedReasons, reason.Error())
	}
	return s.markFailedErr
}

type stubContestRealtimeRelayStreamPublisher struct {
	relays     []contestcontracts.RealtimeRelayEvent
	dedupeKeys []string
	messageID  string
	publishErr error
}

func (s *stubContestRealtimeRelayStreamPublisher) Publish(_ context.Context, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) (string, error) {
	if s.publishErr != nil {
		return "", s.publishErr
	}
	s.relays = append(s.relays, relay)
	s.dedupeKeys = append(s.dedupeKeys, dedupeKey)
	if s.messageID == "" {
		s.messageID = "1710000000000-0"
	}
	return s.messageID, nil
}

func TestContestRealtimeOutboxDispatcherDispatchOnceMarksSent(t *testing.T) {
	outbox := &stubContestRealtimeOutbox{
		pending: []contestports.ContestPendingRealtimeRelay{{
			ID:        42,
			DedupeKey: "contest:77:scoreboard:updated",
			Relay: contestcontracts.RealtimeRelayEvent{
				EventName:   contestcontracts.EventScoreboardUpdated,
				Delivery:    contestcontracts.RealtimeDeliveryChannel,
				Channel:     contestcontracts.ScoreboardChannel(77),
				MessageType: "scoreboard.updated",
				Payload:     contestcontracts.ScoreboardUpdatedRelayPayload{ContestID: 77},
				Timestamp:   time.Now().UTC(),
			},
		}},
	}
	publisher := &stubContestRealtimeRelayStreamPublisher{}
	dispatcher := NewContestRealtimeOutboxDispatcher(outbox, publisher, zap.NewNop())

	if err := dispatcher.DispatchOnce(context.Background()); err != nil {
		t.Fatalf("DispatchOnce() error = %v", err)
	}
	if len(publisher.relays) != 1 || publisher.relays[0].Channel != contestcontracts.ScoreboardChannel(77) {
		t.Fatalf("unexpected published relays: %+v", publisher.relays)
	}
	if len(outbox.sentIDs) != 1 || outbox.sentIDs[0] != 42 {
		t.Fatalf("expected sent outbox id 42, got %+v", outbox.sentIDs)
	}
	if len(outbox.sentStreamMessages) != 1 || outbox.sentStreamMessages[0] != "1710000000000-0" {
		t.Fatalf("expected sent stream message id propagated, got %+v", outbox.sentStreamMessages)
	}
	if len(publisher.dedupeKeys) != 1 || publisher.dedupeKeys[0] != "contest:77:scoreboard:updated" {
		t.Fatalf("expected dedupe key propagated, got %+v", publisher.dedupeKeys)
	}
	if len(outbox.failedIDs) != 0 {
		t.Fatalf("unexpected failed outbox ids: %+v", outbox.failedIDs)
	}
}

func TestContestRealtimeOutboxDispatcherDispatchOnceMarksFailed(t *testing.T) {
	outbox := &stubContestRealtimeOutbox{
		pending: []contestports.ContestPendingRealtimeRelay{{
			ID:        9,
			DedupeKey: "contest:11:announcement:5:deleted",
			Relay: contestcontracts.RealtimeRelayEvent{
				EventName:   contestcontracts.EventAnnouncementDeleted,
				Delivery:    contestcontracts.RealtimeDeliveryChannel,
				Channel:     contestcontracts.AnnouncementChannel(11),
				MessageType: "contest.announcement.deleted",
				Payload:     contestcontracts.AnnouncementDeletedRelayPayload{ContestID: 11, AnnouncementID: 5},
				Timestamp:   time.Now().UTC(),
			},
		}},
	}
	publisher := &stubContestRealtimeRelayStreamPublisher{publishErr: errors.New("redis unavailable")}
	dispatcher := NewContestRealtimeOutboxDispatcher(outbox, publisher, zap.NewNop())

	if err := dispatcher.DispatchOnce(context.Background()); err != nil {
		t.Fatalf("DispatchOnce() error = %v", err)
	}
	if len(outbox.failedIDs) != 1 || outbox.failedIDs[0] != 9 {
		t.Fatalf("expected failed outbox id 9, got %+v", outbox.failedIDs)
	}
	if len(outbox.failedReasons) != 1 || outbox.failedReasons[0] != "redis unavailable" {
		t.Fatalf("unexpected failed reasons: %+v", outbox.failedReasons)
	}
	if len(outbox.sentIDs) != 0 {
		t.Fatalf("unexpected sent outbox ids: %+v", outbox.sentIDs)
	}
}
