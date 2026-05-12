package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	"ctf-platform/internal/module/contest/testsupport"
	platformevents "ctf-platform/internal/platform/events"
)

type scoreboardUpdaterStub struct {
	updateCalls [][2]int64
	rebuilds    []int64
}

func (s *scoreboardUpdaterStub) UpdateScore(_ context.Context, contestID, teamID int64, _ float64) error {
	s.updateCalls = append(s.updateCalls, [2]int64{contestID, teamID})
	return nil
}

func (s *scoreboardUpdaterStub) RebuildScoreboard(_ context.Context, contestID int64) error {
	s.rebuilds = append(s.rebuilds, contestID)
	return nil
}

func TestParticipationServiceCreateAnnouncementBroadcastsRealtimeEvent(t *testing.T) {
	t.Parallel()

	db := testsupport.SetupContestTestDB(t)
	contestRepo := contestinfra.NewRepository(db)
	participationRepo := contestinfra.NewParticipationRepository(db)
	teamRepo := contestinfra.NewTeamRepository(db)
	bus := platformevents.NewBus()
	received := make(chan contestcontracts.AnnouncementCreatedEvent, 1)
	bus.Subscribe(contestcontracts.EventAnnouncementCreated, func(_ context.Context, evt platformevents.Event) error {
		payload, ok := evt.Payload.(contestcontracts.AnnouncementCreatedEvent)
		if !ok {
			t.Fatalf("unexpected payload type: %T", evt.Payload)
		}
		received <- payload
		return nil
	})
	service := &ParticipationService{
		contestRepo: contestRepo,
		repo:        participationRepo,
		teamRepo:    teamRepo,
		eventBus:    bus,
	}

	now := time.Now()
	contest := &model.Contest{
		ID:        77,
		Title:     "realtime-announcement",
		Mode:      model.ContestModeJeopardy,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		Status:    model.ContestStatusRunning,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := db.Create(contest).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}

	item, err := service.CreateAnnouncement(context.Background(), contest.ID, 9001, CreateAnnouncementInput{
		Title:   "比赛开始",
		Content: "欢迎接入实时公告。",
	})
	if err != nil {
		t.Fatalf("CreateAnnouncement() error = %v", err)
	}
	if item == nil || item.ID == 0 {
		t.Fatalf("expected created announcement, got %+v", item)
	}

	select {
	case evt := <-received:
		if evt.ContestID != 77 || evt.AnnouncementID != item.ID {
			t.Fatalf("unexpected event payload: %+v", evt)
		}
		if evt.Title != "比赛开始" || evt.Content != "欢迎接入实时公告。" {
			t.Fatalf("unexpected event announcement body: %+v", evt)
		}
		if evt.CreatedAt.IsZero() || evt.OccurredAt.IsZero() {
			t.Fatalf("expected non-zero event timestamps, got %+v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("expected contest.announcement_created event to be published")
	}
}

func TestSubmissionServiceSyncCorrectSubmissionScoreboardBroadcastsRealtimeEvent(t *testing.T) {
	t.Parallel()

	scoreboard := &scoreboardUpdaterStub{}
	bus := platformevents.NewBus()
	received := make(chan contestcontracts.ScoreboardUpdatedEvent, 1)
	bus.Subscribe(contestcontracts.EventScoreboardUpdated, func(_ context.Context, evt platformevents.Event) error {
		payload, ok := evt.Payload.(contestcontracts.ScoreboardUpdatedEvent)
		if !ok {
			t.Fatalf("unexpected payload type: %T", evt.Payload)
		}
		received <- payload
		return nil
	})
	service := &SubmissionService{
		scoreboardService: scoreboard,
		eventBus:          bus,
	}

	contestID := int64(88)
	if err := service.syncCorrectSubmissionScoreboard(context.Background(), &contestID, map[int64]int{
		301: 150,
		302: 0,
	}); err != nil {
		t.Fatalf("syncCorrectSubmissionScoreboard() error = %v", err)
	}

	if len(scoreboard.updateCalls) != 1 {
		t.Fatalf("expected 1 scoreboard update, got %d", len(scoreboard.updateCalls))
	}

	select {
	case evt := <-received:
		if evt.ContestID != contestID {
			t.Fatalf("unexpected event payload: %+v", evt)
		}
		if evt.OccurredAt.IsZero() {
			t.Fatalf("expected non-zero occurred_at, got %+v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("expected contest.scoreboard_updated event to be published")
	}
}
