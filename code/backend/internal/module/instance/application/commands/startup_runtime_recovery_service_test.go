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
}

func (s *startupRuntimeReconcilerStub) ReconcileLostActiveRuntimes(context.Context) error {
	s.called = true
	s.callOrder++
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
	reconciler := &startupRuntimeReconcilerStub{
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
	stateStore := &startupRuntimeStateStoreStub{
		loadBootID:      "boot-old",
		loadHeartbeatAt: lastHeartbeat,
		loadOK:          true,
	}

	service := NewStartupRuntimeRecoveryService(reconciler, contestRepo, instanceRepo, stateStore, time.Hour, nil)
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
		loadHeartbeatAt: startedAt.Add(-time.Minute),
		loadOK:          true,
	}

	service := NewStartupRuntimeRecoveryService(reconciler, contestRepo, instanceRepo, stateStore, time.Hour, nil)
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
