package commands

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	contestcontracts "ctf-platform/internal/module/contest/contracts"
	instanceports "ctf-platform/internal/module/instance/ports"
)

type startupRuntimeReconcilerStub struct {
	called    bool
	callOrder int
	err       error
	assertFn  func() error
	onCall    func()
}

func (s *startupRuntimeReconcilerStub) ReconcileLostActiveRuntimes(context.Context) error {
	s.called = true
	s.callOrder++
	if s.onCall != nil {
		s.onCall()
	}
	if s.assertFn != nil {
		if err := s.assertFn(); err != nil {
			return err
		}
	}
	return s.err
}

type startupRuntimeDesiredReconcilerStub struct {
	called   bool
	err      error
	assertFn func() error
	onCall   func()
}

func (s *startupRuntimeDesiredReconcilerStub) ReconcileDesiredAWDInstances(context.Context) error {
	s.called = true
	if s.onCall != nil {
		s.onCall()
	}
	if s.assertFn != nil {
		if err := s.assertFn(); err != nil {
			return err
		}
	}
	return s.err
}

type startupRuntimeContestCall struct {
	activeAt            time.Time
	recoveryKey         string
	targetPausedSeconds int64
	updatedAt           time.Time
}

type startupRuntimeContestRepoStub struct {
	contests []*contestcontracts.Contest
	calls    []startupRuntimeContestCall
}

func (s *startupRuntimeContestRepoStub) AddPausedDurationToActiveAWDContests(_ context.Context, activeAt time.Time, recoveryKey string, targetPausedSeconds int64, updatedAt time.Time) ([]*instanceports.ActiveAWDContestPause, error) {
	s.calls = append(s.calls, startupRuntimeContestCall{
		activeAt:            activeAt,
		recoveryKey:         recoveryKey,
		targetPausedSeconds: targetPausedSeconds,
		updatedAt:           updatedAt,
	})
	result := make([]*instanceports.ActiveAWDContestPause, 0, len(s.contests))
	for _, contest := range s.contests {
		if contest == nil {
			continue
		}
		delta := targetPausedSeconds
		if contest.RuntimeRecoveryKey == recoveryKey {
			delta = targetPausedSeconds - contest.RuntimeRecoveryAppliedSeconds
		}
		if delta < 0 {
			delta = 0
		}
		cloned := *contest
		cloned.PausedSeconds += delta
		cloned.RuntimeRecoveryKey = recoveryKey
		cloned.RuntimeRecoveryAppliedSeconds = targetPausedSeconds
		contest.PausedSeconds = cloned.PausedSeconds
		contest.RuntimeRecoveryKey = cloned.RuntimeRecoveryKey
		contest.RuntimeRecoveryAppliedSeconds = cloned.RuntimeRecoveryAppliedSeconds
		result = append(result, &instanceports.ActiveAWDContestPause{
			ID:            cloned.ID,
			EndTime:       cloned.EndTime,
			PausedSeconds: cloned.PausedSeconds,
		})
	}
	return result, nil
}

type startupRuntimeRefreshCall struct {
	contestID int64
	activeAt  time.Time
	expiresAt time.Time
}

type startupRuntimeInstanceRepoStub struct {
	calls []startupRuntimeRefreshCall
}

func (s *startupRuntimeInstanceRepoStub) RefreshActiveAWDInstanceExpiryByContest(_ context.Context, contestID int64, activeAt, expiresAt time.Time) error {
	s.calls = append(s.calls, startupRuntimeRefreshCall{
		contestID: contestID,
		activeAt:  activeAt,
		expiresAt: expiresAt,
	})
	return nil
}

type startupRuntimeStateStoreStub struct {
	loadFn          func(context.Context) (string, time.Time, bool, error)
	saveFn          func(context.Context, string, time.Time) error
	acquireLockFn   func(context.Context, time.Duration) (instanceports.StartupRecoveryLockLease, bool, error)
	loadBootID      string
	loadHeartbeatAt time.Time
	loadOK          bool
	lockAcquired    bool
	lockErr         error
	saveCalls       []struct {
		bootID      string
		heartbeatAt time.Time
	}
}

func (s *startupRuntimeStateStoreStub) LoadPlatformRuntimeState(ctx context.Context) (string, time.Time, bool, error) {
	if s.loadFn != nil {
		return s.loadFn(ctx)
	}
	return s.loadBootID, s.loadHeartbeatAt, s.loadOK, nil
}

func (s *startupRuntimeStateStoreStub) SavePlatformRuntimeState(ctx context.Context, bootID string, heartbeatAt time.Time) error {
	if s.saveFn != nil {
		return s.saveFn(ctx, bootID, heartbeatAt)
	}
	s.saveCalls = append(s.saveCalls, struct {
		bootID      string
		heartbeatAt time.Time
	}{
		bootID:      bootID,
		heartbeatAt: heartbeatAt,
	})
	return nil
}

func (s *startupRuntimeStateStoreStub) AcquireStartupRecoveryLock(ctx context.Context, ttl time.Duration) (instanceports.StartupRecoveryLockLease, bool, error) {
	if s.acquireLockFn != nil {
		return s.acquireLockFn(ctx, ttl)
	}
	if s.lockErr != nil {
		return nil, false, s.lockErr
	}
	if !s.lockAcquired {
		return nil, false, nil
	}
	return nil, true, nil
}

type startupRuntimeBootIDReaderStub struct {
	bootID string
	err    error
}

func (s *startupRuntimeBootIDReaderStub) ReadBootID(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if s.err != nil {
		return "", s.err
	}
	return s.bootID, nil
}

func setStartupRuntimeBootID(service *StartupRuntimeRecoveryService, bootID string) {
	service.bootIDReader = &startupRuntimeBootIDReaderStub{bootID: bootID}
}

func setStartupRuntimeBootIDError(service *StartupRuntimeRecoveryService, err error) {
	service.bootIDReader = &startupRuntimeBootIDReaderStub{err: err}
}

func TestStartupRuntimeRecoveryServiceRebootExtendsContestsBeforeReconcile(t *testing.T) {
	t.Parallel()

	lastHeartbeat := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	startedAt := time.Date(2026, 5, 16, 10, 10, 0, 0, time.UTC)
	recoveredAt := startedAt.Add(30 * time.Second)

	contestRepo := &startupRuntimeContestRepoStub{
		contests: []*contestcontracts.Contest{
			{
				ID:            41,
				Mode:          contestcontracts.ContestModeAWD,
				Status:        contestcontracts.ContestStatusRunning,
				StartTime:     time.Date(2026, 5, 16, 9, 0, 0, 0, time.UTC),
				EndTime:       time.Date(2026, 5, 16, 11, 0, 0, 0, time.UTC),
				UpdatedAt:     lastHeartbeat,
				CreatedAt:     lastHeartbeat,
				PausedSeconds: 0,
			},
		},
	}
	instanceRepo := &startupRuntimeInstanceRepoStub{}
	callSequence := make([]string, 0, 2)
	reconciler := &startupRuntimeReconcilerStub{
		onCall: func() {
			callSequence = append(callSequence, "active")
		},
		assertFn: func() error {
			if len(contestRepo.calls) != 1 {
				return context.Canceled
			}
			if len(instanceRepo.calls) != 1 {
				return context.Canceled
			}
			return nil
		},
	}
	desired := &startupRuntimeDesiredReconcilerStub{
		onCall: func() {
			callSequence = append(callSequence, "desired")
		},
		assertFn: func() error {
			if len(callSequence) != 2 || callSequence[0] != "active" {
				return context.Canceled
			}
			return nil
		},
	}
	stateStore := &startupRuntimeStateStoreStub{
		loadBootID:      "boot-old",
		loadHeartbeatAt: lastHeartbeat,
		loadOK:          true,
		lockAcquired:    true,
	}

	service := NewStartupRuntimeRecoveryService(reconciler, contestRepo, instanceRepo, stateStore, nil, time.Hour, nil)
	service.SetDesiredRuntimeReconciler(desired)
	service.now = newDeterministicNow(startedAt, recoveredAt)
	setStartupRuntimeBootID(service, "boot-new")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() {
		_ = service.Stop(context.Background())
	})

	if !reconciler.called {
		t.Fatal("expected runtime reconciler to be called")
	}
	if !desired.called {
		t.Fatal("expected desired runtime reconciler to be called")
	}
	if len(callSequence) != 2 || callSequence[0] != "active" || callSequence[1] != "desired" {
		t.Fatalf("unexpected recovery order: %+v", callSequence)
	}
	if len(contestRepo.calls) != 2 {
		t.Fatalf("expected two contest extension calls, got %d", len(contestRepo.calls))
	}
	if contestRepo.calls[0].targetPausedSeconds != 600 {
		t.Fatalf("expected initial paused seconds target 600, got %d", contestRepo.calls[0].targetPausedSeconds)
	}
	if contestRepo.calls[1].targetPausedSeconds != 630 {
		t.Fatalf("expected recovery paused seconds target 630, got %d", contestRepo.calls[1].targetPausedSeconds)
	}
	expectedRecoveryKey := buildStartupRuntimeRecoveryKey("boot-old", lastHeartbeat)
	if contestRepo.calls[0].recoveryKey != expectedRecoveryKey || contestRepo.calls[1].recoveryKey != expectedRecoveryKey {
		t.Fatalf("unexpected recovery keys: %+v", contestRepo.calls)
	}
	if len(instanceRepo.calls) != 2 {
		t.Fatalf("expected two instance expiry refreshes, got %d", len(instanceRepo.calls))
	}
	if !instanceRepo.calls[0].expiresAt.Equal(time.Date(2026, 5, 16, 11, 10, 0, 0, time.UTC)) {
		t.Fatalf("unexpected initial refreshed expiry: %s", instanceRepo.calls[0].expiresAt)
	}
	if !instanceRepo.calls[1].expiresAt.Equal(time.Date(2026, 5, 16, 11, 10, 30, 0, time.UTC)) {
		t.Fatalf("unexpected final refreshed expiry: %s", instanceRepo.calls[1].expiresAt)
	}
	if len(stateStore.saveCalls) == 0 {
		t.Fatal("expected heartbeat state to be saved")
	}
	lastSave := stateStore.saveCalls[len(stateStore.saveCalls)-1]
	if lastSave.bootID != "boot-new" {
		t.Fatalf("expected saved boot id boot-new, got %q", lastSave.bootID)
	}
	if !lastSave.heartbeatAt.Equal(recoveredAt) {
		t.Fatalf("expected recovered heartbeat %s, got %s", recoveredAt, lastSave.heartbeatAt)
	}
}

func TestStartupRuntimeRecoveryServiceSameBootOnlyRecordsHeartbeat(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, 5, 16, 10, 10, 0, 0, time.UTC)
	contestRepo := &startupRuntimeContestRepoStub{}
	instanceRepo := &startupRuntimeInstanceRepoStub{}
	reconciler := &startupRuntimeReconcilerStub{}
	stateStore := &startupRuntimeStateStoreStub{
		loadBootID:      "boot-same",
		loadHeartbeatAt: startedAt.Add(-45 * time.Second),
		loadOK:          true,
		lockAcquired:    true,
	}

	service := NewStartupRuntimeRecoveryService(reconciler, contestRepo, instanceRepo, stateStore, nil, 30*time.Second, nil)
	service.now = newDeterministicNow(startedAt)
	setStartupRuntimeBootID(service, "boot-same")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() {
		_ = service.Stop(context.Background())
	})

	if reconciler.called {
		t.Fatal("did not expect runtime reconciler to run on same boot")
	}
	if len(contestRepo.calls) != 0 {
		t.Fatalf("expected no contest extension calls, got %d", len(contestRepo.calls))
	}
	if len(instanceRepo.calls) != 0 {
		t.Fatalf("expected no instance refresh calls, got %d", len(instanceRepo.calls))
	}
	if len(stateStore.saveCalls) != 1 {
		t.Fatalf("expected exactly one heartbeat save, got %d", len(stateStore.saveCalls))
	}
	if !stateStore.saveCalls[0].heartbeatAt.Equal(startedAt) {
		t.Fatalf("expected heartbeat saved at %s, got %s", startedAt, stateStore.saveCalls[0].heartbeatAt)
	}
}

func TestStartupRuntimeRecoveryServiceSameBootWithStaleHeartbeatSkipsRecovery(t *testing.T) {
	t.Parallel()

	lastHeartbeat := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	startedAt := time.Date(2026, 5, 16, 10, 2, 0, 0, time.UTC)

	contestRepo := &startupRuntimeContestRepoStub{
		contests: []*contestcontracts.Contest{
			{
				ID:            61,
				Mode:          contestcontracts.ContestModeAWD,
				Status:        contestcontracts.ContestStatusRunning,
				StartTime:     time.Date(2026, 5, 16, 9, 0, 0, 0, time.UTC),
				EndTime:       time.Date(2026, 5, 16, 11, 0, 0, 0, time.UTC),
				UpdatedAt:     lastHeartbeat,
				CreatedAt:     lastHeartbeat,
				PausedSeconds: 0,
			},
		},
	}
	instanceRepo := &startupRuntimeInstanceRepoStub{}
	reconciler := &startupRuntimeReconcilerStub{}
	stateStore := &startupRuntimeStateStoreStub{
		loadBootID:      "boot-same",
		loadHeartbeatAt: lastHeartbeat,
		loadOK:          true,
		lockAcquired:    true,
	}

	service := NewStartupRuntimeRecoveryService(reconciler, contestRepo, instanceRepo, stateStore, nil, 30*time.Second, nil)
	service.now = newDeterministicNow(startedAt)
	setStartupRuntimeBootID(service, "boot-same")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() {
		_ = service.Stop(context.Background())
	})

	if reconciler.called {
		t.Fatal("expected same-boot stale heartbeat to skip runtime outage recovery")
	}
	if len(contestRepo.calls) != 0 {
		t.Fatalf("expected no contest extension calls, got %d", len(contestRepo.calls))
	}
	if len(instanceRepo.calls) != 0 {
		t.Fatalf("expected no instance expiry refreshes, got %d", len(instanceRepo.calls))
	}
	if len(stateStore.saveCalls) != 1 {
		t.Fatalf("expected exactly one heartbeat save, got %d", len(stateStore.saveCalls))
	}
	lastSave := stateStore.saveCalls[len(stateStore.saveCalls)-1]
	if !lastSave.heartbeatAt.Equal(startedAt) {
		t.Fatalf("expected heartbeat %s, got %s", startedAt, lastSave.heartbeatAt)
	}
}

func TestStartupRuntimeRecoveryServiceRetryDoesNotDoubleCountPreviouslyAppliedPause(t *testing.T) {
	t.Parallel()

	lastHeartbeat := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	firstStartedAt := time.Date(2026, 5, 16, 10, 10, 0, 0, time.UTC)
	secondStartedAt := time.Date(2026, 5, 16, 10, 12, 0, 0, time.UTC)
	secondRecoveredAt := secondStartedAt.Add(30 * time.Second)

	contestRepo := &startupRuntimeContestRepoStub{
		contests: []*contestcontracts.Contest{
			{
				ID:        52,
				Mode:      contestcontracts.ContestModeAWD,
				Status:    contestcontracts.ContestStatusRunning,
				StartTime: time.Date(2026, 5, 16, 9, 0, 0, 0, time.UTC),
				EndTime:   time.Date(2026, 5, 16, 11, 0, 0, 0, time.UTC),
				UpdatedAt: lastHeartbeat,
				CreatedAt: lastHeartbeat,
			},
		},
	}
	instanceRepo := &startupRuntimeInstanceRepoStub{}
	stateStore := &startupRuntimeStateStoreStub{
		loadBootID:      "boot-old",
		loadHeartbeatAt: lastHeartbeat,
		loadOK:          true,
		lockAcquired:    true,
	}

	firstService := NewStartupRuntimeRecoveryService(&startupRuntimeReconcilerStub{err: context.Canceled}, contestRepo, instanceRepo, stateStore, nil, time.Hour, nil)
	firstService.now = newDeterministicNow(firstStartedAt)
	setStartupRuntimeBootID(firstService, "boot-new")
	if err := firstService.Start(context.Background()); err == nil {
		t.Fatal("expected first recovery attempt to fail")
	}

	secondService := NewStartupRuntimeRecoveryService(&startupRuntimeReconcilerStub{}, contestRepo, instanceRepo, stateStore, nil, time.Hour, nil)
	secondService.now = newDeterministicNow(secondStartedAt, secondRecoveredAt)
	setStartupRuntimeBootID(secondService, "boot-new")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := secondService.Start(ctx); err != nil {
		t.Fatalf("second Start() error = %v", err)
	}
	t.Cleanup(func() {
		_ = secondService.Stop(context.Background())
	})

	if got := contestRepo.contests[0].PausedSeconds; got != 750 {
		t.Fatalf("expected total paused seconds 750 after retry, got %d", got)
	}
	if contestRepo.contests[0].RuntimeRecoveryAppliedSeconds != 750 {
		t.Fatalf("expected recorded applied seconds 750, got %d", contestRepo.contests[0].RuntimeRecoveryAppliedSeconds)
	}
}

func TestStartupRuntimeRecoveryServiceStartRetryAfterInitFailure(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, 5, 16, 10, 10, 0, 0, time.UTC)
	stateStore := &startupRuntimeStateStoreStub{lockAcquired: true}
	service := NewStartupRuntimeRecoveryService(nil, nil, nil, stateStore, nil, time.Hour, nil)
	service.now = newDeterministicNow(startedAt)
	setStartupRuntimeBootIDError(service, errors.New("missing boot id"))

	if err := service.Start(context.Background()); err == nil {
		t.Fatal("expected first Start() to fail")
	}
	if len(stateStore.saveCalls) != 0 {
		t.Fatalf("expected no heartbeat save after failed start, got %d", len(stateStore.saveCalls))
	}

	setStartupRuntimeBootID(service, "boot-retry-ok")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("second Start() error = %v", err)
	}
	t.Cleanup(func() {
		_ = service.Stop(context.Background())
	})

	if len(stateStore.saveCalls) != 1 {
		t.Fatalf("expected second Start() to record heartbeat once, got %d", len(stateStore.saveCalls))
	}
	if stateStore.saveCalls[0].bootID != "boot-retry-ok" {
		t.Fatalf("expected saved boot id boot-retry-ok, got %q", stateStore.saveCalls[0].bootID)
	}
	if !stateStore.saveCalls[0].heartbeatAt.Equal(startedAt) {
		t.Fatalf("expected saved heartbeat at %s, got %s", startedAt, stateStore.saveCalls[0].heartbeatAt)
	}
}

func TestStartupRuntimeRecoveryServiceCanRestartAfterStop(t *testing.T) {
	t.Parallel()

	firstStartedAt := time.Date(2026, 5, 16, 10, 10, 0, 0, time.UTC)
	secondStartedAt := time.Date(2026, 5, 16, 10, 12, 0, 0, time.UTC)
	stateStore := &startupRuntimeStateStoreStub{lockAcquired: true}
	service := NewStartupRuntimeRecoveryService(nil, nil, nil, stateStore, nil, time.Hour, nil)
	service.now = newDeterministicNow(firstStartedAt, secondStartedAt)
	setStartupRuntimeBootID(service, "boot-restart-ok")

	firstCtx, firstCancel := context.WithCancel(context.Background())
	defer firstCancel()
	if err := service.Start(firstCtx); err != nil {
		t.Fatalf("first Start() error = %v", err)
	}
	if len(stateStore.saveCalls) != 1 {
		t.Fatalf("expected first Start() to record heartbeat once, got %d", len(stateStore.saveCalls))
	}

	if err := service.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	secondCtx, secondCancel := context.WithCancel(context.Background())
	defer secondCancel()
	if err := service.Start(secondCtx); err != nil {
		t.Fatalf("second Start() error = %v", err)
	}
	t.Cleanup(func() {
		_ = service.Stop(context.Background())
	})

	if len(stateStore.saveCalls) != 2 {
		t.Fatalf("expected second Start() to record a second heartbeat, got %d", len(stateStore.saveCalls))
	}
	if !stateStore.saveCalls[1].heartbeatAt.Equal(secondStartedAt) {
		t.Fatalf("expected second heartbeat at %s, got %s", secondStartedAt, stateStore.saveCalls[1].heartbeatAt)
	}
}

func TestStartupRuntimeRecoveryServiceStandbyReplicaWaitsForLeaderReady(t *testing.T) {
	t.Parallel()

	reconciler := &startupRuntimeReconcilerStub{}
	leaderStateChecked := make(chan struct{}, 1)
	stateStore := &startupRuntimeStateStoreStub{
		loadFn: func(context.Context) (string, time.Time, bool, error) {
			select {
			case leaderStateChecked <- struct{}{}:
			default:
			}
			return "boot-prev", time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC), true, nil
		},
		lockAcquired: false,
	}
	service := NewStartupRuntimeRecoveryService(reconciler, nil, nil, stateStore, nil, time.Hour, nil)
	service.now = newDeterministicNow(time.Date(2026, 5, 16, 10, 1, 0, 0, time.UTC))
	setStartupRuntimeBootID(service, "boot-standby")
	service.leaderRetry = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	startDone := make(chan error, 1)
	go func() {
		startDone <- service.Start(ctx)
	}()

	waitForStartupRecoveryLoad(t, leaderStateChecked)
	assertStartupRecoveryStillWaiting(t, startDone, "standby Start() should wait for leader readiness")

	if reconciler.called {
		t.Fatal("expected standby replica to skip recovery")
	}
	if len(stateStore.saveCalls) != 0 {
		t.Fatalf("expected standby replica to skip heartbeat writes, got %d", len(stateStore.saveCalls))
	}

	cancel()
	select {
	case err := <-startDone:
		if err == nil {
			t.Fatal("expected canceled standby Start() to return context error")
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected standby Start() to stop after context cancellation")
	}
}

func TestStartupRuntimeRecoveryServiceConcurrentStartWaitsForLeaderReady(t *testing.T) {
	t.Parallel()

	leaderStateChecked := make(chan struct{}, 1)
	stateStore := &startupRuntimeStateStoreStub{
		loadFn: func(context.Context) (string, time.Time, bool, error) {
			select {
			case leaderStateChecked <- struct{}{}:
			default:
			}
			return "boot-prev", time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC), true, nil
		},
		lockAcquired: false,
	}
	service := NewStartupRuntimeRecoveryService(nil, nil, nil, stateStore, nil, time.Hour, nil)
	service.now = newDeterministicNow(time.Date(2026, 5, 16, 10, 1, 0, 0, time.UTC))
	setStartupRuntimeBootID(service, "boot-standby")
	service.leaderRetry = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	firstDone := make(chan error, 1)
	go func() {
		firstDone <- service.Start(ctx)
	}()
	waitForStartupRecoveryLoad(t, leaderStateChecked)
	assertStartupRecoveryStillWaiting(t, firstDone, "first Start() should wait for leader readiness")

	secondDone := make(chan error, 1)
	go func() {
		secondDone <- service.Start(ctx)
	}()
	assertStartupRecoveryStillWaiting(t, secondDone, "concurrent Start() should share readiness wait")

	cancel()
	for name, done := range map[string]<-chan error{
		"first Start()":  firstDone,
		"second Start()": secondDone,
	} {
		select {
		case err := <-done:
			if err == nil {
				t.Fatalf("expected canceled %s to return context error", name)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatalf("expected canceled %s to stop", name)
		}
	}
}

func waitForStartupRecoveryLoad(t *testing.T, observed <-chan struct{}) {
	t.Helper()
	select {
	case <-observed:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected startup recovery to inspect leader readiness state")
	}
}

func assertStartupRecoveryStillWaiting(t *testing.T, done <-chan error, message string) {
	t.Helper()
	select {
	case err := <-done:
		t.Fatalf("%s, got err=%v", message, err)
	case <-time.After(25 * time.Millisecond):
	}
}

func TestStartupRuntimeRecoveryServiceStandbyReplicaCanTakeOverBeforeStartReturns(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, 5, 16, 10, 1, 0, 0, time.UTC)
	var mu sync.Mutex
	allowAcquire := false
	leaderStateChecked := make(chan struct{}, 1)
	heartbeatSaved := make(chan struct{}, 1)
	stateStore := &startupRuntimeStateStoreStub{
		loadFn: func(context.Context) (string, time.Time, bool, error) {
			select {
			case leaderStateChecked <- struct{}{}:
			default:
			}
			return "boot-ready", startedAt.Add(-2 * time.Minute), true, nil
		},
		acquireLockFn: func(context.Context, time.Duration) (instanceports.StartupRecoveryLockLease, bool, error) {
			mu.Lock()
			defer mu.Unlock()
			return nil, allowAcquire, nil
		},
		saveFn: func(context.Context, string, time.Time) error {
			select {
			case heartbeatSaved <- struct{}{}:
			default:
			}
			return nil
		},
	}
	service := NewStartupRuntimeRecoveryService(&startupRuntimeReconcilerStub{}, nil, nil, stateStore, nil, 0, nil)
	service.now = newDeterministicNow(startedAt)
	setStartupRuntimeBootID(service, "boot-ready")
	service.leaderRetry = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startDone := make(chan error, 1)
	go func() {
		startDone <- service.Start(ctx)
	}()

	waitForStartupRecoveryLoad(t, leaderStateChecked)
	assertStartupRecoveryStillWaiting(t, startDone, "standby Start() should wait before takeover")

	mu.Lock()
	allowAcquire = true
	mu.Unlock()

	select {
	case err := <-startDone:
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected standby Start() to return after takeover records heartbeat")
	}
	t.Cleanup(func() {
		_ = service.Stop(context.Background())
	})

	select {
	case <-heartbeatSaved:
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected async standby loop to take over and record heartbeat")
	}
}

func TestStartupRuntimeRecoveryServiceStandbyReplicaReturnsAfterForeignLeaderReady(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, 5, 16, 10, 1, 0, 0, time.UTC)
	var mu sync.Mutex
	foreignLeaderReady := false
	leaderStateChecked := make(chan struct{}, 1)
	reconciler := &startupRuntimeReconcilerStub{}
	stateStore := &startupRuntimeStateStoreStub{
		loadFn: func(context.Context) (string, time.Time, bool, error) {
			select {
			case leaderStateChecked <- struct{}{}:
			default:
			}
			mu.Lock()
			defer mu.Unlock()
			if foreignLeaderReady {
				return "boot-ready", startedAt, true, nil
			}
			return "boot-ready", startedAt.Add(-2 * time.Minute), true, nil
		},
		acquireLockFn: func(context.Context, time.Duration) (instanceports.StartupRecoveryLockLease, bool, error) {
			return nil, false, nil
		},
	}
	service := NewStartupRuntimeRecoveryService(reconciler, nil, nil, stateStore, nil, 0, nil)
	service.now = func() time.Time { return startedAt }
	setStartupRuntimeBootID(service, "boot-ready")
	service.leaderRetry = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	startDone := make(chan error, 1)
	go func() {
		startDone <- service.Start(ctx)
	}()

	waitForStartupRecoveryLoad(t, leaderStateChecked)
	assertStartupRecoveryStillWaiting(t, startDone, "standby Start() should wait before foreign leader heartbeat is ready")

	mu.Lock()
	foreignLeaderReady = true
	mu.Unlock()

	select {
	case err := <-startDone:
		if err != nil {
			t.Fatalf("Start() error = %v", err)
		}
	case <-time.After(250 * time.Millisecond):
		t.Fatal("expected standby Start() to return after foreign leader heartbeat is ready")
	}
	t.Cleanup(func() {
		_ = service.Stop(context.Background())
	})

	if reconciler.called {
		t.Fatal("expected standby replica to skip recovery when foreign leader is ready")
	}
	if len(stateStore.saveCalls) != 0 {
		t.Fatalf("expected standby replica to skip heartbeat writes, got %d", len(stateStore.saveCalls))
	}
}

func TestStartupRuntimeRecoveryServiceLeaderFailoverWithinSafeWindowSkipsRecovery(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, 5, 16, 10, 1, 0, 0, time.UTC)
	reconciler := &startupRuntimeReconcilerStub{}
	stateStore := &startupRuntimeStateStoreStub{
		loadBootID:      "boot-shared",
		loadHeartbeatAt: startedAt.Add(-55 * time.Second),
		loadOK:          true,
		lockAcquired:    true,
	}
	service := NewStartupRuntimeRecoveryService(reconciler, nil, nil, stateStore, nil, 0, nil)
	service.now = newDeterministicNow(startedAt)
	setStartupRuntimeBootID(service, "boot-shared")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() {
		_ = service.Stop(context.Background())
	})

	if reconciler.called {
		t.Fatal("expected leader failover within safe window to skip recovery")
	}
	if len(stateStore.saveCalls) != 1 {
		t.Fatalf("expected leader takeover to record a fresh heartbeat once, got %d", len(stateStore.saveCalls))
	}
	if !stateStore.saveCalls[0].heartbeatAt.Equal(startedAt) {
		t.Fatalf("expected failover heartbeat at %s, got %s", startedAt, stateStore.saveCalls[0].heartbeatAt)
	}
}

func TestStartupRuntimeRecoveryServiceSameBootLeaderGapSkipsRecovery(t *testing.T) {
	t.Parallel()

	startedAt := time.Date(2026, 5, 16, 10, 10, 0, 0, time.UTC)
	reconciler := &startupRuntimeReconcilerStub{}
	stateStore := &startupRuntimeStateStoreStub{
		loadBootID:      "boot-shared",
		loadHeartbeatAt: startedAt.Add(-10 * time.Minute),
		loadOK:          true,
		lockAcquired:    true,
	}
	service := NewStartupRuntimeRecoveryService(reconciler, nil, nil, stateStore, nil, 0, nil)
	service.now = newDeterministicNow(startedAt)
	setStartupRuntimeBootID(service, "boot-shared")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() {
		_ = service.Stop(context.Background())
	})

	if reconciler.called {
		t.Fatal("expected same-boot leader gap to skip runtime outage recovery")
	}
	if len(stateStore.saveCalls) != 1 {
		t.Fatalf("expected leader gap takeover to record a fresh heartbeat once, got %d", len(stateStore.saveCalls))
	}
	if !stateStore.saveCalls[0].heartbeatAt.Equal(startedAt) {
		t.Fatalf("expected leader gap heartbeat at %s, got %s", startedAt, stateStore.saveCalls[0].heartbeatAt)
	}
}

func newDeterministicNow(values ...time.Time) func() time.Time {
	index := 0
	return func() time.Time {
		if len(values) == 0 {
			return time.Time{}
		}
		if index >= len(values) {
			return values[len(values)-1]
		}
		value := values[index]
		index++
		return value
	}
}
