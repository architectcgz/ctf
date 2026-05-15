package jobs

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	"ctf-platform/internal/module/contest/testsupport"
	rediskeys "ctf-platform/internal/pkg/redis"
)

type statusUpdaterRepoStub struct {
	contests             []*model.Contest
	updatedStatus        map[int64]string
	updatedStatusVersion map[int64]int64
	receivedStatus       []string
	listCalls            int
	listCalled           chan struct{}
	listBlock            chan struct{}
	transitionApplied    bool
	transitionConfigured bool
}

type endedContestRuntimeCleanerStub struct {
	cleanedContestIDs []int64
	err               error
}

func (s *endedContestRuntimeCleanerStub) CleanupEndedContestAWDInstances(_ context.Context, contestID int64) error {
	s.cleanedContestIDs = append(s.cleanedContestIDs, contestID)
	return s.err
}

func (s *statusUpdaterRepoStub) ListByStatusesAndTimeRange(_ context.Context, statuses []string, _ time.Time, _, _ int) ([]*model.Contest, int64, error) {
	s.listCalls++
	s.receivedStatus = append([]string(nil), statuses...)
	if s.listCalled != nil && s.listCalls == 1 {
		close(s.listCalled)
	}
	if s.listBlock != nil {
		<-s.listBlock
	}
	return s.contests, int64(len(s.contests)), nil
}

func (s *statusUpdaterRepoStub) ApplyStatusTransition(_ context.Context, transition contestdomain.ContestStatusTransition) (contestdomain.ContestStatusTransitionResult, error) {
	if s.updatedStatus == nil {
		s.updatedStatus = make(map[int64]string)
	}
	if s.updatedStatusVersion == nil {
		s.updatedStatusVersion = make(map[int64]int64)
	}
	s.updatedStatus[transition.ContestID] = transition.ToStatus
	s.updatedStatusVersion[transition.ContestID] = transition.FromStatusVersion + 1
	applied := s.transitionApplied
	if !s.transitionConfigured {
		applied = true
	}
	return contestdomain.ContestStatusTransitionResult{
		Transition:    transition,
		Applied:       applied,
		StatusVersion: transition.FromStatusVersion + 1,
	}, nil
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
	updater := NewStatusUpdater(repo, time.Minute, 100, 30*time.Second, nil)
	repo.transitionApplied = true

	updater.updateStatuses(context.Background())

	if got := repo.updatedStatus[7]; got != model.ContestStatusEnded {
		t.Fatalf("expected frozen contest to end, got %q", got)
	}
}

func TestStatusUpdaterUpdateStatuses_RequestsFrozenStatus(t *testing.T) {
	repo := &statusUpdaterRepoStub{}
	updater := NewStatusUpdater(repo, time.Minute, 100, 30*time.Second, nil)

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

	runtimeCleaner := &endedContestRuntimeCleanerStub{}
	updater := NewStatusUpdater(repo, time.Minute, 100, 30*time.Second, nil)
	updater.SetStatusSideEffectStore(contestinfra.NewContestStatusSideEffectStore(redisClient, runtimeCleaner))
	repo.transitionApplied = true

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
	if len(runtimeCleaner.cleanedContestIDs) != 1 || runtimeCleaner.cleanedContestIDs[0] != 11 {
		t.Fatalf("expected ended runtime cleaner to run for contest 11, got %+v", runtimeCleaner.cleanedContestIDs)
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

func TestStatusUpdaterRecordsAppliedTransitionAndSideEffectStatus(t *testing.T) {
	db := testsupport.SetupContestTestDB(t)
	now := time.Now().UTC()
	freezeTime := now.Add(-time.Minute)
	if err := db.Create(&model.Contest{
		ID:            41,
		Title:         "recorded-freeze",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusRunning,
		StatusVersion: 0,
		StartTime:     now.Add(-time.Hour),
		EndTime:       now.Add(time.Hour),
		FreezeTime:    &freezeTime,
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
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
	if err := redisClient.ZAdd(context.Background(), rediskeys.RankContestTeamKey(41), redis.Z{Score: 20, Member: "team-9"}).Err(); err != nil {
		t.Fatalf("seed team rank: %v", err)
	}

	updater := NewStatusUpdater(contestinfra.NewRepository(db), time.Minute, 100, time.Minute, nil)
	updater.SetStatusSideEffectStore(contestinfra.NewContestStatusSideEffectStore(redisClient))
	updater.updateStatuses(context.Background())

	var contest model.Contest
	if err := db.First(&contest, 41).Error; err != nil {
		t.Fatalf("load contest: %v", err)
	}
	if contest.Status != model.ContestStatusFrozen || contest.StatusVersion != 1 {
		t.Fatalf("unexpected contest state: %+v", contest)
	}

	var transition model.ContestStatusTransition
	if err := db.Where("contest_id = ? AND status_version = ?", 41, 1).First(&transition).Error; err != nil {
		t.Fatalf("load transition record: %v", err)
	}
	if transition.FromStatus != model.ContestStatusRunning || transition.ToStatus != model.ContestStatusFrozen {
		t.Fatalf("unexpected transition record: %+v", transition)
	}
	if transition.SideEffectStatus != contestdomain.ContestStatusTransitionSideEffectSucceeded {
		t.Fatalf("expected succeeded side effect status, got %+v", transition)
	}
	if !mini.Exists(rediskeys.RankContestFrozenKey(41)) {
		t.Fatal("expected frozen snapshot key to be created")
	}
}

func TestStatusUpdaterReplaysFailedTransitionSideEffects(t *testing.T) {
	db := testsupport.SetupContestTestDB(t)
	now := time.Now().UTC()
	if err := db.Create(&model.Contest{
		ID:            42,
		Title:         "replay-freeze",
		Mode:          model.ContestModeJeopardy,
		Status:        model.ContestStatusFrozen,
		StatusVersion: 1,
		StartTime:     now.Add(-time.Hour),
		EndTime:       now.Add(time.Hour),
		CreatedAt:     now,
		UpdatedAt:     now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.ContestStatusTransition{
		ID:               4201,
		ContestID:        42,
		StatusVersion:    1,
		FromStatus:       model.ContestStatusRunning,
		ToStatus:         model.ContestStatusFrozen,
		Reason:           contestdomain.ContestStatusTransitionReasonTimeWindow,
		AppliedBy:        contestStatusUpdaterAppliedBy,
		SideEffectStatus: contestdomain.ContestStatusTransitionSideEffectFailed,
		SideEffectError:  "redis timeout",
		OccurredAt:       now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}).Error; err != nil {
		t.Fatalf("create failed transition record: %v", err)
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
	if err := redisClient.ZAdd(context.Background(), rediskeys.RankContestTeamKey(42), redis.Z{Score: 8, Member: "team-2"}).Err(); err != nil {
		t.Fatalf("seed team rank: %v", err)
	}

	updater := NewStatusUpdater(contestinfra.NewRepository(db), time.Minute, 100, time.Minute, nil)
	updater.SetStatusSideEffectStore(contestinfra.NewContestStatusSideEffectStore(redisClient))
	updater.updateStatuses(context.Background())

	if !mini.Exists(rediskeys.RankContestFrozenKey(42)) {
		t.Fatal("expected replay to rebuild frozen snapshot")
	}

	var transition model.ContestStatusTransition
	if err := db.First(&transition, 4201).Error; err != nil {
		t.Fatalf("load transition record: %v", err)
	}
	if transition.SideEffectStatus != contestdomain.ContestStatusTransitionSideEffectSucceeded {
		t.Fatalf("expected replayed record to succeed, got %+v", transition)
	}
}

func TestStatusUpdaterRefreshesSchedulerLockWhileRunning(t *testing.T) {
	repo := &statusUpdaterRepoStub{
		listCalled: make(chan struct{}),
		listBlock:  make(chan struct{}),
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

	repo.contests = []*model.Contest{
		{
			ID:            31,
			Status:        model.ContestStatusRegistration,
			StatusVersion: 0,
			StartTime:     time.Now().Add(-time.Minute),
			EndTime:       time.Now().Add(time.Hour),
		},
	}
	updater := NewStatusUpdater(repo, time.Minute, 100, 60*time.Millisecond, nil)
	updater.SetStatusSideEffectStore(contestinfra.NewContestStatusSideEffectStore(redisClient))
	updater.SetStatusUpdateLockStore(contestinfra.NewContestStatusUpdateLockStore(redisClient))
	repo.transitionApplied = true

	lockKey := rediskeys.ContestStatusUpdateLockKey()
	done := make(chan struct{})
	go func() {
		defer close(done)
		updater.updateStatuses(context.Background())
	}()

	select {
	case <-repo.listCalled:
	case <-time.After(time.Second):
		t.Fatal("expected updater to start list call")
	}

	time.Sleep(140 * time.Millisecond)
	if !mini.Exists(lockKey) {
		t.Fatalf("expected lock %q to be refreshed during run", lockKey)
	}
	if ttl := mini.TTL(lockKey); ttl <= 0 {
		t.Fatalf("expected positive ttl during run, got %s", ttl)
	}

	close(repo.listBlock)
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("expected updater to finish")
	}

	if mini.Exists(lockKey) {
		t.Fatalf("expected lock %q to be released after run", lockKey)
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

	updater := NewStatusUpdater(repo, time.Minute, 100, time.Minute, nil)
	updater.SetStatusSideEffectStore(contestinfra.NewContestStatusSideEffectStore(redisClient))
	updater.SetStatusUpdateLockStore(contestinfra.NewContestStatusUpdateLockStore(redisClient))
	updater.updateStatuses(context.Background())

	if len(repo.receivedStatus) != 0 {
		t.Fatalf("expected scheduler to skip when lock held, got statuses %+v", repo.receivedStatus)
	}
}

func TestStatusUpdaterSkipsSideEffectsWhenTransitionIsStale(t *testing.T) {
	repo := &statusUpdaterRepoStub{
		contests: []*model.Contest{
			{
				ID:            51,
				Status:        model.ContestStatusRunning,
				StatusVersion: 4,
				StartTime:     time.Now().Add(-time.Hour),
				EndTime:       time.Now().Add(time.Hour),
				FreezeTime:    timePtr(time.Now().Add(-time.Minute)),
			},
		},
		transitionApplied:    false,
		transitionConfigured: true,
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
	if err := redisClient.ZAdd(context.Background(), rediskeys.RankContestTeamKey(51), redis.Z{Score: 10, Member: "team-1"}).Err(); err != nil {
		t.Fatalf("seed source rank: %v", err)
	}

	updater := NewStatusUpdater(repo, time.Minute, 100, time.Minute, nil)
	updater.SetStatusSideEffectStore(contestinfra.NewContestStatusSideEffectStore(redisClient))
	updater.updateStatuses(context.Background())

	if mini.Exists(rediskeys.RankContestFrozenKey(51)) {
		t.Fatal("expected stale transition to skip frozen snapshot")
	}
}

func timePtr(v time.Time) *time.Time {
	return &v
}

func TestStatusUpdaterStartRunsImmediately(t *testing.T) {
	repo := &statusUpdaterRepoStub{
		listCalled: make(chan struct{}),
	}
	updater := NewStatusUpdater(repo, time.Hour, 100, 30*time.Second, nil)

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
