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

func TestParticipationRepositoryListsAnnouncementSyncEventsAndCursor(t *testing.T) {
	t.Parallel()

	db := contesttestsupport.SetupContestTestDB(t)
	repo := contestinfra.NewParticipationRepository(db)
	now := time.Date(2026, 6, 7, 13, 10, 0, 0, time.UTC)

	rows := []contestentity.ContestRealtimeOutbox{
		{
			ID:              1,
			EventName:       contestcontracts.EventAnnouncementCreated,
			Delivery:        contestcontracts.RealtimeDeliveryChannel,
			Channel:         contestcontracts.AnnouncementChannel(77),
			MessageType:     "contest.announcement.created",
			Payload:         `{"contest_id":77,"announcement":{"id":501,"title":"公告一","content":"正文","created_at":"2026-06-07T13:09:00Z"}}`,
			DedupeKey:       "contest:77:announcement:501:created",
			Status:          contestdomain.ContestRealtimeOutboxStatusSent,
			EventOccurredAt: now.Add(-time.Minute),
			NextAttemptAt:   now.Add(-time.Minute),
			StreamMessageID: "1710000000001-0",
			SentAt:          ptrTime(now.Add(-time.Minute)),
			CreatedAt:       now.Add(-time.Minute),
			UpdatedAt:       now.Add(-time.Minute),
		},
		{
			ID:              2,
			EventName:       contestcontracts.EventAnnouncementDeleted,
			Delivery:        contestcontracts.RealtimeDeliveryChannel,
			Channel:         contestcontracts.AnnouncementChannel(77),
			MessageType:     "contest.announcement.deleted",
			Payload:         `{"contest_id":77,"announcement_id":501}`,
			DedupeKey:       "contest:77:announcement:501:deleted",
			Status:          contestdomain.ContestRealtimeOutboxStatusSent,
			EventOccurredAt: now,
			NextAttemptAt:   now,
			StreamMessageID: "1710000000002-0",
			SentAt:          ptrTime(now),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
		{
			ID:              3,
			EventName:       contestcontracts.EventAnnouncementCreated,
			Delivery:        contestcontracts.RealtimeDeliveryChannel,
			Channel:         contestcontracts.AnnouncementChannel(88),
			MessageType:     "contest.announcement.created",
			Payload:         `{"contest_id":88,"announcement":{"id":601,"title":"其他竞赛","content":"","created_at":"2026-06-07T13:10:00Z"}}`,
			DedupeKey:       "contest:88:announcement:601:created",
			Status:          contestdomain.ContestRealtimeOutboxStatusSent,
			EventOccurredAt: now,
			NextAttemptAt:   now,
			StreamMessageID: "1710000000003-0",
			SentAt:          ptrTime(now),
			CreatedAt:       now,
			UpdatedAt:       now,
		},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatalf("seed outbox rows: %v", err)
	}

	events, err := repo.ListAnnouncementSyncEvents(context.Background(), 77, 0, 10)
	if err != nil {
		t.Fatalf("ListAnnouncementSyncEvents() error = %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %+v", events)
	}
	if events[0].Announcement == nil || events[0].Announcement.ID != 501 {
		t.Fatalf("unexpected created sync event: %+v", events[0])
	}
	if events[1].AnnouncementID == nil || *events[1].AnnouncementID != 501 {
		t.Fatalf("unexpected deleted sync event: %+v", events[1])
	}

	cursor, err := repo.LatestAnnouncementSyncCursor(context.Background(), 77)
	if err != nil {
		t.Fatalf("LatestAnnouncementSyncCursor() error = %v", err)
	}
	if cursor != 2 {
		t.Fatalf("expected latest cursor 2, got %d", cursor)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
