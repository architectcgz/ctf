package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	"ctf-platform/internal/module/contest/testsupport"
	rediskeys "ctf-platform/internal/pkg/redis"
)

type statusUpdaterRepoStub struct {
	contests       []*model.Contest
	updatedStatus  map[int64]string
	receivedStatus []string
	listCalls      int
	listCalled     chan struct{}
}

func (s *statusUpdaterRepoStub) ListByStatusesAndTimeRange(_ context.Context, statuses []string, _ time.Time, _, _ int) ([]*model.Contest, int64, error) {
	s.listCalls++
	s.receivedStatus = append([]string(nil), statuses...)
	if s.listCalled != nil && s.listCalls == 1 {
		close(s.listCalled)
	}
	return s.contests, int64(len(s.contests)), nil
}

func (s *statusUpdaterRepoStub) UpdateStatus(_ context.Context, id int64, status string) error {
	if s.updatedStatus == nil {
		s.updatedStatus = make(map[int64]string)
	}
	s.updatedStatus[id] = status
	return nil
}

func TestStatusUpdaterUpdateStatuses_EndsFrozenContest(t *testing.T) {
	now := time.Now()
	repo := &statusUpdaterRepoStub{
		contests: []*model.Contest{
			{
				ID:        7,
				Status:    model.ContestStatusFrozen,
				StartTime: now.Add(-2 * time.Hour),
				EndTime:   now.Add(-time.Minute),
			},
		},
	}
	updater := NewStatusUpdater(repo, nil, time.Minute, 100, 30*time.Second, nil)

	updater.updateStatuses(context.Background())

	if got := repo.updatedStatus[7]; got != model.ContestStatusEnded {
		t.Fatalf("expected frozen contest to end, got %q", got)
	}
}

func TestStatusUpdaterUpdateStatuses_RequestsFrozenStatus(t *testing.T) {
	repo := &statusUpdaterRepoStub{}
	updater := NewStatusUpdater(repo, nil, time.Minute, 100, 30*time.Second, nil)

	updater.updateStatuses(context.Background())

	expected := []string{
		model.ContestStatusRegistration,
		model.ContestStatusRunning,
		model.ContestStatusFrozen,
	}
	if len(repo.receivedStatus) != len(expected) {
		t.Fatalf("expected %d statuses, got %d", len(expected), len(repo.receivedStatus))
	}
	for i, status := range expected {
		if repo.receivedStatus[i] != status {
			t.Fatalf("expected status %q at index %d, got %q", status, i, repo.receivedStatus[i])
		}
	}
}

func TestStatusUpdaterUpdateStatuses_ClearsAWDRuntimeStateWhenContestEnds(t *testing.T) {
	now := time.Now()
	repo := &statusUpdaterRepoStub{
		contests: []*model.Contest{
			{
				ID:        11,
				Status:    model.ContestStatusRunning,
				StartTime: now.Add(-2 * time.Hour),
				EndTime:   now.Add(-time.Minute),
			},
		},
	}

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	if err := redisClient.Set(context.Background(), rediskeys.AWDCurrentRoundKey(11), "4", 0).Err(); err != nil {
		t.Fatalf("seed current round: %v", err)
	}
	if err := redisClient.HSet(context.Background(), rediskeys.AWDServiceStatusKey(11), "11:22", model.AWDServiceStatusUp).Err(); err != nil {
		t.Fatalf("seed service status cache: %v", err)
	}

	updater := NewStatusUpdater(repo, redisClient, time.Minute, 100, 30*time.Second, nil)

	updater.updateStatuses(context.Background())

	if got := repo.updatedStatus[11]; got != model.ContestStatusEnded {
		t.Fatalf("expected running contest to end, got %q", got)
	}
	if mini.Exists(rediskeys.AWDCurrentRoundKey(11)) {
		t.Fatalf("expected current round key to be cleared")
	}
	if mini.Exists(rediskeys.AWDServiceStatusKey(11)) {
		t.Fatalf("expected service status key to be cleared")
	}
}

func TestStatusUpdaterUpdateStatuses_BlocksAWDRegistrationStartWhenReadinessNotReady(t *testing.T) {
	db := testsupport.SetupAWDTestDB(t)
	now := time.Now().UTC()
	contestID := int64(12)
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "blocked-awd-start",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRegistration,
		StartTime: now.Add(-time.Minute),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}

	updater := NewStatusUpdater(
		contestinfra.NewRepository(db),
		nil,
		time.Minute,
		100,
		30*time.Second,
		nil,
		contestinfra.NewAWDRepository(db),
	)
	updater.updateStatuses(context.Background())

	var contest model.Contest
	if err := db.First(&contest, contestID).Error; err != nil {
		t.Fatalf("load contest: %v", err)
	}
	if contest.Status != model.ContestStatusRegistration {
		t.Fatalf("expected readiness to keep contest in registration, got %q", contest.Status)
	}
}

func TestStatusUpdaterUpdateStatuses_SkipsWhenSchedulerLockHeld(t *testing.T) {
	repo := &statusUpdaterRepoStub{
		contests: []*model.Contest{
			{ID: 1, Status: model.ContestStatusRunning, StartTime: time.Now().Add(-time.Hour), EndTime: time.Now().Add(time.Hour)},
		},
	}

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	redisClient := redis.NewClient(&redis.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = redisClient.Close()
	})

	if err := mini.Set(rediskeys.ContestStatusUpdateLockKey(), "busy"); err != nil {
		t.Fatalf("seed scheduler lock: %v", err)
	}

	updater := NewStatusUpdater(repo, redisClient, time.Minute, 100, time.Minute, nil)
	updater.updateStatuses(context.Background())

	if len(repo.receivedStatus) != 0 {
		t.Fatalf("expected scheduler to skip when lock held, got statuses %+v", repo.receivedStatus)
	}
}

func TestStatusUpdaterStartRunsImmediately(t *testing.T) {
	repo := &statusUpdaterRepoStub{
		listCalled: make(chan struct{}),
	}
	updater := NewStatusUpdater(repo, nil, time.Hour, 100, 30*time.Second, nil)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		defer close(done)
		updater.Start(ctx)
	}()

	select {
	case <-repo.listCalled:
		cancel()
	case <-time.After(time.Second):
		cancel()
		t.Fatal("expected updater to run immediately on start")
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected updater goroutine to exit after cancel")
	}

	if repo.listCalls != 1 {
		t.Fatalf("expected exactly one immediate update, got %d", repo.listCalls)
	}
}
