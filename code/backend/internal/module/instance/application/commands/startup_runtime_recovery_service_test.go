package commands

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ctf-platform/internal/model"
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
	contests []*model.Contest
	calls    []startupRuntimeContestCall
}

func (s *startupRuntimeContestRepoStub) AddPausedDurationToActiveAWDContests(_ context.Context, activeAt time.Time, recoveryKey string, targetPausedSeconds int64, updatedAt time.Time) ([]*model.Contest, error) {
	s.calls = append(s.calls, startupRuntimeContestCall{
		activeAt:            activeAt,
		recoveryKey:         recoveryKey,
		targetPausedSeconds: targetPausedSeconds,
		updatedAt:           updatedAt,
	})
	result := make([]*model.Contest, 0, len(s.contests))
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
		result = append(result, &cloned)
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
	loadBootID      string
	loadHeartbeatAt time.Time
	loadOK          bool
	saveCalls       []struct {
		bootID      string
		heartbeatAt time.Time
	}
}

func (s *startupRuntimeStateStoreStub) LoadPlatformRuntimeState(context.Context) (string, time.Time, bool, error) {
	return s.loadBootID, s.loadHeartbeatAt, s.loadOK, nil
}

func (s *startupRuntimeStateStoreStub) SavePlatformRuntimeState(_ context.Context, bootID string, heartbeatAt time.Time) error {
	s.saveCalls = append(s.saveCalls, struct {
		bootID      string
		heartbeatAt time.Time
	}{
		bootID:      bootID,
		heartbeatAt: heartbeatAt,
	})
	return nil
}

func TestStartupRuntimeRecoveryServiceRebootExtendsContestsBeforeReconcile(t *testing.T) {
	t.Parallel()

	lastHeartbeat := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	startedAt := time.Date(2026, 5, 16, 10, 10, 0, 0, time.UTC)
	recoveredAt := startedAt.Add(30 * time.Second)

	contestRepo := &startupRuntimeContestRepoStub{
		contests: []*model.Contest{
			{
				ID:            41,
				Mode:          model.ContestModeAWD,
				Status:        model.ContestStatusRunning,
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
	}

	service := NewStartupRuntimeRecoveryService(reconciler, contestRepo, instanceRepo, stateStore, time.Hour, nil)
	service.SetDesiredRuntimeReconciler(desired)
	service.now = newDeterministicNow(startedAt, recoveredAt)
	service.bootIDPath = writeBootIDFile(t, "boot-new")

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
	}

	service := NewStartupRuntimeRecoveryService(reconciler, contestRepo, instanceRepo, stateStore, 30*time.Second, nil)
	service.now = newDeterministicNow(startedAt)
	service.bootIDPath = writeBootIDFile(t, "boot-same")

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

func TestStartupRuntimeRecoveryServiceSameBootWithStaleHeartbeatTriggersRecovery(t *testing.T) {
	t.Parallel()

	lastHeartbeat := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	startedAt := time.Date(2026, 5, 16, 10, 2, 0, 0, time.UTC)
	recoveredAt := startedAt.Add(20 * time.Second)

	contestRepo := &startupRuntimeContestRepoStub{
		contests: []*model.Contest{
			{
				ID:            61,
				Mode:          model.ContestModeAWD,
				Status:        model.ContestStatusRunning,
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
	}

	service := NewStartupRuntimeRecoveryService(reconciler, contestRepo, instanceRepo, stateStore, 30*time.Second, nil)
	service.now = newDeterministicNow(startedAt, recoveredAt)
	service.bootIDPath = writeBootIDFile(t, "boot-same")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := service.Start(ctx); err != nil {
		t.Fatalf("Start() error = %v", err)
	}
	t.Cleanup(func() {
		_ = service.Stop(context.Background())
	})

	if !reconciler.called {
		t.Fatal("expected runtime reconciler to run when heartbeat is stale")
	}
	if len(contestRepo.calls) != 2 {
		t.Fatalf("expected two contest extension calls, got %d", len(contestRepo.calls))
	}
	if contestRepo.calls[0].targetPausedSeconds != 120 {
		t.Fatalf("expected initial paused seconds target 120, got %d", contestRepo.calls[0].targetPausedSeconds)
	}
	if contestRepo.calls[1].targetPausedSeconds != 140 {
		t.Fatalf("expected final paused seconds target 140, got %d", contestRepo.calls[1].targetPausedSeconds)
	}
	if len(instanceRepo.calls) != 2 {
		t.Fatalf("expected two instance expiry refreshes, got %d", len(instanceRepo.calls))
	}
	expectedRecoveryKey := buildStartupRuntimeRecoveryKey("boot-same", lastHeartbeat)
	if contestRepo.calls[0].recoveryKey != expectedRecoveryKey {
		t.Fatalf("unexpected recovery key %q", contestRepo.calls[0].recoveryKey)
	}
	lastSave := stateStore.saveCalls[len(stateStore.saveCalls)-1]
	if !lastSave.heartbeatAt.Equal(recoveredAt) {
		t.Fatalf("expected recovered heartbeat %s, got %s", recoveredAt, lastSave.heartbeatAt)
	}
}

func TestStartupRuntimeRecoveryServiceRetryDoesNotDoubleCountPreviouslyAppliedPause(t *testing.T) {
	t.Parallel()

	lastHeartbeat := time.Date(2026, 5, 16, 10, 0, 0, 0, time.UTC)
	firstStartedAt := time.Date(2026, 5, 16, 10, 10, 0, 0, time.UTC)
	secondStartedAt := time.Date(2026, 5, 16, 10, 12, 0, 0, time.UTC)
	secondRecoveredAt := secondStartedAt.Add(30 * time.Second)

	contestRepo := &startupRuntimeContestRepoStub{
		contests: []*model.Contest{
			{
				ID:        52,
				Mode:      model.ContestModeAWD,
				Status:    model.ContestStatusRunning,
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
	}

	firstService := NewStartupRuntimeRecoveryService(&startupRuntimeReconcilerStub{err: context.Canceled}, contestRepo, instanceRepo, stateStore, time.Hour, nil)
	firstService.now = newDeterministicNow(firstStartedAt)
	firstService.bootIDPath = writeBootIDFile(t, "boot-new")
	if err := firstService.Start(context.Background()); err == nil {
		t.Fatal("expected first recovery attempt to fail")
	}

	secondService := NewStartupRuntimeRecoveryService(&startupRuntimeReconcilerStub{}, contestRepo, instanceRepo, stateStore, time.Hour, nil)
	secondService.now = newDeterministicNow(secondStartedAt, secondRecoveredAt)
	secondService.bootIDPath = writeBootIDFile(t, "boot-new")
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
	stateStore := &startupRuntimeStateStoreStub{}
	service := NewStartupRuntimeRecoveryService(nil, nil, nil, stateStore, time.Hour, nil)
	service.now = newDeterministicNow(startedAt)
	service.bootIDPath = filepath.Join(t.TempDir(), "missing-boot-id")

	if err := service.Start(context.Background()); err == nil {
		t.Fatal("expected first Start() to fail")
	}
	if len(stateStore.saveCalls) != 0 {
		t.Fatalf("expected no heartbeat save after failed start, got %d", len(stateStore.saveCalls))
	}

	service.bootIDPath = writeBootIDFile(t, "boot-retry-ok")
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
	stateStore := &startupRuntimeStateStoreStub{}
	service := NewStartupRuntimeRecoveryService(nil, nil, nil, stateStore, time.Hour, nil)
	service.now = newDeterministicNow(firstStartedAt, secondStartedAt)
	service.bootIDPath = writeBootIDFile(t, "boot-restart-ok")

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

func writeBootIDFile(t *testing.T, bootID string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "boot_id")
	if err := os.WriteFile(path, []byte(bootID), 0o600); err != nil {
		t.Fatalf("write boot id file: %v", err)
	}
	return path
}
