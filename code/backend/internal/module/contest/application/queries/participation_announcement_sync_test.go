package queries

import (
	"context"
	"testing"
	"time"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestports "ctf-platform/internal/module/contest/ports"
)

func TestParticipationServiceSyncAnnouncementsWithoutCursorAnchorsLatestEvent(t *testing.T) {
	t.Parallel()

	service := NewParticipationService(
		participationQueryContestLookupStub{},
		participationQueryRepoStub{
			latestAnnouncementSyncCursorFn: func(context.Context, int64) (int64, error) {
				return 42, nil
			},
		},
		participationQueryTeamFinderStub{},
	)

	result, err := service.SyncAnnouncements(context.Background(), 10, nil)
	if err != nil {
		t.Fatalf("SyncAnnouncements() error = %v", err)
	}
	if result.NextCursor != 42 {
		t.Fatalf("expected next cursor 42, got %d", result.NextCursor)
	}
	if result.HasMore {
		t.Fatalf("expected no more events, got %+v", result)
	}
	if len(result.Events) != 0 {
		t.Fatalf("expected no events when anchoring, got %+v", result.Events)
	}
}

func TestParticipationServiceSyncAnnouncementsMapsCreatedAndDeletedEvents(t *testing.T) {
	t.Parallel()

	occurredAt := time.Date(2026, 6, 7, 13, 0, 0, 0, time.UTC)
	afterID := int64(100)
	service := NewParticipationService(
		participationQueryContestLookupStub{},
		participationQueryRepoStub{
			listAnnouncementSyncEventsFn: func(_ context.Context, contestID int64, nextAfterID int64, limit int) ([]*contestports.ContestAnnouncementSyncEventRow, error) {
				if contestID != 10 {
					t.Fatalf("unexpected contest id %d", contestID)
				}
				if nextAfterID != afterID {
					t.Fatalf("unexpected after id %d", nextAfterID)
				}
				if limit != 101 {
					t.Fatalf("unexpected limit %d", limit)
				}
				deleteID := int64(501)
				return []*contestports.ContestAnnouncementSyncEventRow{
					{
						Cursor:      101,
						MessageType: "contest.announcement.created",
						Announcement: &contestcontracts.AnnouncementRealtimePayload{
							ID:        502,
							Title:     "新公告",
							Content:   "增量内容",
							CreatedAt: occurredAt,
						},
						OccurredAt: occurredAt,
					},
					{
						Cursor:         102,
						MessageType:    "contest.announcement.deleted",
						AnnouncementID: &deleteID,
						OccurredAt:     occurredAt.Add(time.Second),
					},
				}, nil
			},
		},
		participationQueryTeamFinderStub{},
	)

	result, err := service.SyncAnnouncements(context.Background(), 10, &afterID)
	if err != nil {
		t.Fatalf("SyncAnnouncements() error = %v", err)
	}
	if result.NextCursor != 102 {
		t.Fatalf("expected next cursor 102, got %d", result.NextCursor)
	}
	if result.HasMore {
		t.Fatalf("expected hasMore=false, got %+v", result)
	}
	if len(result.Events) != 2 {
		t.Fatalf("expected 2 events, got %+v", result.Events)
	}
	if result.Events[0].Announcement == nil || result.Events[0].Announcement.ID != 502 {
		t.Fatalf("unexpected created event: %+v", result.Events[0])
	}
	if result.Events[1].AnnouncementID == nil || *result.Events[1].AnnouncementID != 501 {
		t.Fatalf("unexpected deleted event: %+v", result.Events[1])
	}
}
