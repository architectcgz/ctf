package infrastructure_test

import (
	"context"
	"testing"
	"time"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contesttestsupport "ctf-platform/internal/module/contest/testsupport"
)

func TestRealtimeOutboxRepositoryEnqueueAndListPending(t *testing.T) {
	t.Parallel()

	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewRealtimeOutboxRepository(db)
	now := time.Now().UTC()

	relay := contestcontracts.RealtimeRelayEvent{
		EventName:   contestcontracts.EventAnnouncementCreated,
		Delivery:    contestcontracts.RealtimeDeliveryChannel,
		Channel:     contestcontracts.AnnouncementChannel(77),
		MessageType: "contest.announcement.created",
		Payload: contestcontracts.AnnouncementCreatedRelayPayload{
			ContestID: 77,
			Announcement: contestcontracts.AnnouncementRealtimePayload{
				ID:        501,
				Title:     "公告标题",
				Content:   "公告正文",
				CreatedAt: now,
			},
		},
		Timestamp: now,
	}
	if err := repo.EnqueueRealtimeRelay(context.Background(), relay, "contest:77:announcement:501:created"); err != nil {
		t.Fatalf("EnqueueRealtimeRelay() error = %v", err)
	}

	pending, err := repo.ListPendingRealtimeRelays(context.Background(), now.Add(time.Second), 10)
	if err != nil {
		t.Fatalf("ListPendingRealtimeRelays() error = %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending relay, got %d", len(pending))
	}
	if pending[0].Relay.EventName != contestcontracts.EventAnnouncementCreated {
		t.Fatalf("unexpected relay event name: %+v", pending[0])
	}
	if pending[0].Relay.Channel != contestcontracts.AnnouncementChannel(77) {
		t.Fatalf("unexpected relay channel: %+v", pending[0])
	}
	if pending[0].Relay.Timestamp != now {
		t.Fatalf("expected relay timestamp %s, got %s", now, pending[0].Relay.Timestamp)
	}
}

func TestRealtimeOutboxRepositoryMarksSentAndFailed(t *testing.T) {
	t.Parallel()

	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewRealtimeOutboxRepository(db)
	now := time.Now().UTC()

	if err := repo.EnqueueRealtimeRelay(context.Background(), contestcontracts.RealtimeRelayEvent{
		EventName:   contestcontracts.EventAnnouncementDeleted,
		Delivery:    contestcontracts.RealtimeDeliveryChannel,
		Channel:     contestcontracts.AnnouncementChannel(88),
		MessageType: "contest.announcement.deleted",
		Payload: contestcontracts.AnnouncementDeletedRelayPayload{
			ContestID:      88,
			AnnouncementID: 9,
		},
		Timestamp: now,
	}, "contest:88:announcement:9:deleted"); err != nil {
		t.Fatalf("EnqueueRealtimeRelay() error = %v", err)
	}

	pending, err := repo.ListPendingRealtimeRelays(context.Background(), now.Add(time.Second), 10)
	if err != nil {
		t.Fatalf("ListPendingRealtimeRelays() error = %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending relay, got %d", len(pending))
	}

	sentAt := now.Add(2 * time.Second)
	if err := repo.MarkRealtimeRelaySent(context.Background(), pending[0].ID, "1710000000000-0", sentAt); err != nil {
		t.Fatalf("MarkRealtimeRelaySent() error = %v", err)
	}

	var sent contestentity.ContestRealtimeOutbox
	if err := db.First(&sent, pending[0].ID).Error; err != nil {
		t.Fatalf("load sent outbox row: %v", err)
	}
	if sent.Status != contestdomain.ContestRealtimeOutboxStatusSent {
		t.Fatalf("expected sent status, got %+v", sent)
	}
	if sent.StreamMessageID != "1710000000000-0" {
		t.Fatalf("unexpected stream message id: %+v", sent)
	}

	if err := repo.EnqueueRealtimeRelay(context.Background(), contestcontracts.RealtimeRelayEvent{
		EventName:   contestcontracts.EventAnnouncementDeleted,
		Delivery:    contestcontracts.RealtimeDeliveryChannel,
		Channel:     contestcontracts.AnnouncementChannel(99),
		MessageType: "contest.announcement.deleted",
		Payload: contestcontracts.AnnouncementDeletedRelayPayload{
			ContestID:      99,
			AnnouncementID: 10,
		},
		Timestamp: now,
	}, "contest:99:announcement:10:deleted"); err != nil {
		t.Fatalf("EnqueueRealtimeRelay(second) error = %v", err)
	}

	pending, err = repo.ListPendingRealtimeRelays(context.Background(), now.Add(time.Second), 10)
	if err != nil {
		t.Fatalf("ListPendingRealtimeRelays(second) error = %v", err)
	}
	var failedID int64
	for _, item := range pending {
		if item.Relay.Channel == contestcontracts.AnnouncementChannel(99) {
			failedID = item.ID
			break
		}
	}
	if failedID == 0 {
		t.Fatalf("expected pending relay for contest 99, got %+v", pending)
	}

	nextAttemptAt := now.Add(5 * time.Minute)
	if err := repo.MarkRealtimeRelayFailed(context.Background(), failedID, context.DeadlineExceeded, nextAttemptAt); err != nil {
		t.Fatalf("MarkRealtimeRelayFailed() error = %v", err)
	}

	var failed contestentity.ContestRealtimeOutbox
	if err := db.First(&failed, failedID).Error; err != nil {
		t.Fatalf("load failed outbox row: %v", err)
	}
	if failed.Status != contestdomain.ContestRealtimeOutboxStatusPending {
		t.Fatalf("expected pending retry status, got %+v", failed)
	}
	if failed.AttemptCount != 1 {
		t.Fatalf("expected attempt count 1, got %+v", failed)
	}
	if failed.LastError == "" {
		t.Fatalf("expected failure reason recorded, got %+v", failed)
	}

	pending, err = repo.ListPendingRealtimeRelays(context.Background(), nextAttemptAt.Add(time.Second), 10)
	if err != nil {
		t.Fatalf("ListPendingRealtimeRelays(after failure) error = %v", err)
	}
	for _, item := range pending {
		if item.ID != failedID {
			continue
		}
		if !item.Relay.Timestamp.Equal(now) {
			t.Fatalf("expected original relay timestamp preserved, got %s want %s", item.Relay.Timestamp, now)
		}
		return
	}
	t.Fatalf("expected failed relay %d in pending list, got %+v", failedID, pending)
}

func TestRealtimeOutboxRepositoryDoesNotRevertSentRelayToPendingOnLateFailure(t *testing.T) {
	t.Parallel()

	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewRealtimeOutboxRepository(db)
	now := time.Now().UTC()

	if err := repo.EnqueueRealtimeRelay(context.Background(), contestcontracts.RealtimeRelayEvent{
		EventName:   contestcontracts.EventScoreboardUpdated,
		Delivery:    contestcontracts.RealtimeDeliveryChannel,
		Channel:     contestcontracts.ScoreboardChannel(123),
		MessageType: "scoreboard.updated",
		Payload:     contestcontracts.ScoreboardUpdatedRelayPayload{ContestID: 123},
		Timestamp:   now,
	}, "contest:123:submission:501:scoreboard_updated"); err != nil {
		t.Fatalf("EnqueueRealtimeRelay() error = %v", err)
	}

	pending, err := repo.ListPendingRealtimeRelays(context.Background(), now.Add(time.Second), 10)
	if err != nil {
		t.Fatalf("ListPendingRealtimeRelays() error = %v", err)
	}
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending relay, got %+v", pending)
	}

	if err := repo.MarkRealtimeRelaySent(context.Background(), pending[0].ID, "1710000000000-0", now.Add(time.Second)); err != nil {
		t.Fatalf("MarkRealtimeRelaySent() error = %v", err)
	}
	if err := repo.MarkRealtimeRelayFailed(context.Background(), pending[0].ID, context.DeadlineExceeded, now.Add(2*time.Second)); err != nil {
		t.Fatalf("MarkRealtimeRelayFailed() error = %v", err)
	}

	var row contestentity.ContestRealtimeOutbox
	if err := db.First(&row, pending[0].ID).Error; err != nil {
		t.Fatalf("load row: %v", err)
	}
	if row.Status != contestdomain.ContestRealtimeOutboxStatusSent {
		t.Fatalf("expected sent status to remain stable, got %+v", row)
	}
	if row.AttemptCount != 0 {
		t.Fatalf("expected late failure not to increment attempt count, got %+v", row)
	}
}
