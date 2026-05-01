package commands

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	"ctf-platform/internal/module/contest/testsupport"
	ctfws "ctf-platform/pkg/websocket"
)

type contestRealtimeBroadcastCall struct {
	channel string
	message ctfws.Envelope
}

type contestRealtimeUserBroadcastCall struct {
	userID  int64
	message ctfws.Envelope
}

type contestRealtimeBroadcasterStub struct {
	calls     []contestRealtimeBroadcastCall
	userCalls []contestRealtimeUserBroadcastCall
}

func (s *contestRealtimeBroadcasterStub) SendToChannel(channel string, message ctfws.Envelope) int {
	s.calls = append(s.calls, contestRealtimeBroadcastCall{
		channel: channel,
		message: message,
	})
	return 1
}

func (s *contestRealtimeBroadcasterStub) SendToUser(userID int64, message ctfws.Envelope) int {
	s.userCalls = append(s.userCalls, contestRealtimeUserBroadcastCall{
		userID:  userID,
		message: message,
	})
	return 1
}

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
	broadcaster := &contestRealtimeBroadcasterStub{}
	service := &ParticipationService{
		contestRepo: contestRepo,
		repo:        participationRepo,
		teamRepo:    teamRepo,
		broadcaster: broadcaster,
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
	if len(broadcaster.calls) != 1 {
		t.Fatalf("expected 1 realtime broadcast, got %d", len(broadcaster.calls))
	}
	if broadcaster.calls[0].channel != "contest:77:announcements" {
		t.Fatalf("unexpected broadcast channel: %s", broadcaster.calls[0].channel)
	}
	if broadcaster.calls[0].message.Type != "contest.announcement.created" {
		t.Fatalf("unexpected broadcast type: %s", broadcaster.calls[0].message.Type)
	}
}

func TestSubmissionServiceSyncCorrectSubmissionScoreboardBroadcastsRealtimeEvent(t *testing.T) {
	t.Parallel()

	scoreboard := &scoreboardUpdaterStub{}
	broadcaster := &contestRealtimeBroadcasterStub{}
	service := &SubmissionService{
		scoreboardService: scoreboard,
		broadcaster:       broadcaster,
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
	if len(broadcaster.calls) != 1 {
		t.Fatalf("expected 1 realtime broadcast, got %d", len(broadcaster.calls))
	}
	if broadcaster.calls[0].channel != "contest:88:scoreboard" {
		t.Fatalf("unexpected broadcast channel: %s", broadcaster.calls[0].channel)
	}
	if broadcaster.calls[0].message.Type != "scoreboard.updated" {
		t.Fatalf("unexpected broadcast type: %s", broadcaster.calls[0].message.Type)
	}
}
