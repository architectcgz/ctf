package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/model"
)

const (
	defaultStartupRuntimeHeartbeatInterval = 30 * time.Second
	defaultStartupRuntimeRecoveryTimeout   = 5 * time.Minute
	defaultBootIDPath                      = "/proc/sys/kernel/random/boot_id"
)

type startupRuntimeReconciler interface {
	ReconcileLostActiveRuntimes(ctx context.Context) error
}

type startupRuntimeContestRepository interface {
	AddPausedDurationToActiveAWDContests(ctx context.Context, activeAt time.Time, recoveryKey string, targetPausedSeconds int64, updatedAt time.Time) ([]*model.Contest, error)
}

type startupRuntimeInstanceRepository interface {
	RefreshActiveAWDInstanceExpiryByContest(ctx context.Context, contestID int64, activeAt, expiresAt time.Time) error
}

type startupRuntimeStateStore interface {
	LoadPlatformRuntimeState(ctx context.Context) (string, time.Time, bool, error)
	SavePlatformRuntimeState(ctx context.Context, bootID string, heartbeatAt time.Time) error
}

type StartupRuntimeRecoveryService struct {
	reconciler   startupRuntimeReconciler
	contests     startupRuntimeContestRepository
	instances    startupRuntimeInstanceRepository
	stateStore   startupRuntimeStateStore
	log          *zap.Logger
	now          func() time.Time
	bootIDPath   string
	heartbeatGap time.Duration

	mu      sync.Mutex
	cancel  context.CancelFunc
	started bool
	wg      sync.WaitGroup
}

func NewStartupRuntimeRecoveryService(
	reconciler startupRuntimeReconciler,
	contests startupRuntimeContestRepository,
	instances startupRuntimeInstanceRepository,
	stateStore startupRuntimeStateStore,
	heartbeatInterval time.Duration,
	logger *zap.Logger,
) *StartupRuntimeRecoveryService {
	if logger == nil {
		logger = zap.NewNop()
	}
	if heartbeatInterval <= 0 {
		heartbeatInterval = defaultStartupRuntimeHeartbeatInterval
	}
	return &StartupRuntimeRecoveryService{
		reconciler:   reconciler,
		contests:     contests,
		instances:    instances,
		stateStore:   stateStore,
		log:          logger,
		now:          func() time.Time { return time.Now().UTC() },
		bootIDPath:   defaultBootIDPath,
		heartbeatGap: heartbeatInterval,
	}
}

func (s *StartupRuntimeRecoveryService) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("startup runtime recovery requires context")
	}

	s.mu.Lock()
	if s.started {
		s.mu.Unlock()
		return nil
	}
	runCtx, cancel := context.WithCancel(ctx)
	s.cancel = cancel
	s.started = true
	s.mu.Unlock()

	currentBootID, err := s.readCurrentBootID()
	if err != nil {
		return err
	}

	startedAt := s.now()
	lastBootID, lastHeartbeatAt, ok, err := s.loadPreviousRuntimeState(runCtx)
	if err != nil {
		return err
	}
	if ok && s.shouldRecoverFromHostReboot(lastBootID, currentBootID, lastHeartbeatAt) {
		if err := s.recoverFromHostReboot(runCtx, currentBootID, lastBootID, lastHeartbeatAt, startedAt); err != nil {
			return err
		}
	} else if err := s.recordHeartbeat(runCtx, currentBootID, startedAt); err != nil {
		return err
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runHeartbeatLoop(runCtx, currentBootID)
	}()
	return nil
}

func (s *StartupRuntimeRecoveryService) Stop(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("startup runtime recovery stop requires context")
	}

	s.mu.Lock()
	if !s.started {
		s.mu.Unlock()
		return nil
	}
	cancel := s.cancel
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}

	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *StartupRuntimeRecoveryService) recoverFromHostReboot(ctx context.Context, currentBootID, lastBootID string, lastHeartbeatAt, startedAt time.Time) error {
	s.log.Warn(
		"host_reboot_detected_for_runtime_recovery",
		zap.Time("last_heartbeat_at", lastHeartbeatAt),
		zap.Time("started_at", startedAt),
	)

	recoveryCtx, cancel := context.WithTimeout(ctx, defaultStartupRuntimeRecoveryTimeout)
	defer cancel()

	recoveryKey := buildStartupRuntimeRecoveryKey(lastBootID, lastHeartbeatAt)
	initialPause := startedAt.Sub(lastHeartbeatAt)
	if err := s.extendActiveAWDContests(recoveryCtx, lastHeartbeatAt, recoveryKey, initialPause, startedAt); err != nil {
		return err
	}
	if s.reconciler != nil {
		if err := s.reconciler.ReconcileLostActiveRuntimes(recoveryCtx); err != nil {
			return err
		}
	}

	recoveredAt := s.now()
	totalPause := recoveredAt.Sub(lastHeartbeatAt)
	if err := s.extendActiveAWDContests(recoveryCtx, lastHeartbeatAt, recoveryKey, totalPause, recoveredAt); err != nil {
		return err
	}
	return s.recordHeartbeat(recoveryCtx, currentBootID, recoveredAt)
}

func (s *StartupRuntimeRecoveryService) extendActiveAWDContests(ctx context.Context, activeAt time.Time, recoveryKey string, targetPause time.Duration, updatedAt time.Time) error {
	targetPausedSeconds := int64(targetPause / time.Second)
	if targetPausedSeconds <= 0 || s.contests == nil {
		return nil
	}

	contests, err := s.contests.AddPausedDurationToActiveAWDContests(ctx, activeAt, recoveryKey, targetPausedSeconds, updatedAt)
	if err != nil {
		return err
	}
	if len(contests) == 0 || s.instances == nil {
		return nil
	}
	for _, contest := range contests {
		if contest == nil {
			continue
		}
		if err := s.instances.RefreshActiveAWDInstanceExpiryByContest(
			ctx,
			contest.ID,
			activeAt,
			startupRuntimeContestEffectiveEndTime(contest),
		); err != nil {
			return err
		}
	}
	return nil
}

func startupRuntimeContestEffectiveEndTime(contest *model.Contest) time.Time {
	if contest == nil {
		return time.Time{}
	}
	return contest.EndTime.UTC().Add(time.Duration(contest.PausedSeconds) * time.Second)
}

func buildStartupRuntimeRecoveryKey(lastBootID string, lastHeartbeatAt time.Time) string {
	return strings.TrimSpace(lastBootID) + "|" + lastHeartbeatAt.UTC().Format(time.RFC3339Nano)
}

func (s *StartupRuntimeRecoveryService) shouldRecoverFromHostReboot(lastBootID, currentBootID string, lastHeartbeatAt time.Time) bool {
	if strings.TrimSpace(lastBootID) == "" || strings.TrimSpace(currentBootID) == "" || lastHeartbeatAt.IsZero() {
		return false
	}
	return strings.TrimSpace(lastBootID) != strings.TrimSpace(currentBootID)
}

func (s *StartupRuntimeRecoveryService) runHeartbeatLoop(ctx context.Context, bootID string) {
	ticker := time.NewTicker(s.heartbeatGap)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.recordHeartbeat(ctx, bootID, s.now()); err != nil && ctx.Err() == nil {
				s.log.Warn("save_platform_runtime_heartbeat_failed", zap.Error(err))
			}
		}
	}
}

func (s *StartupRuntimeRecoveryService) loadPreviousRuntimeState(ctx context.Context) (string, time.Time, bool, error) {
	if s == nil || s.stateStore == nil {
		return "", time.Time{}, false, nil
	}
	return s.stateStore.LoadPlatformRuntimeState(ctx)
}

func (s *StartupRuntimeRecoveryService) recordHeartbeat(ctx context.Context, bootID string, heartbeatAt time.Time) error {
	if s == nil || s.stateStore == nil {
		return nil
	}
	return s.stateStore.SavePlatformRuntimeState(ctx, bootID, heartbeatAt)
}

func (s *StartupRuntimeRecoveryService) readCurrentBootID() (string, error) {
	content, err := os.ReadFile(s.bootIDPath)
	if err != nil {
		return "", err
	}
	bootID := strings.TrimSpace(string(content))
	if bootID == "" {
		return "", fmt.Errorf("boot id is empty")
	}
	return bootID, nil
}
