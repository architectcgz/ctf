package commands

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/infrastructure/redislock"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	startuprecovery "ctf-platform/internal/module/instance/startuprecovery"
	"ctf-platform/internal/shared/lockkeepalive"
)

const (
	defaultStartupRuntimeRecoveryTimeout = 5 * time.Minute
	defaultBootIDPath                    = "/proc/sys/kernel/random/boot_id"
)

type startupRuntimeReconciler interface {
	ReconcileLostActiveRuntimes(ctx context.Context) error
}

type startupRuntimeDesiredReconciler interface {
	ReconcileDesiredAWDInstances(ctx context.Context) error
}

type startupRuntimeContestRepository interface {
	AddPausedDurationToActiveAWDContests(ctx context.Context, activeAt time.Time, recoveryKey string, targetPausedSeconds int64, updatedAt time.Time) ([]*contestcontracts.Contest, error)
}

type startupRuntimeInstanceRepository interface {
	RefreshActiveAWDInstanceExpiryByContest(ctx context.Context, contestID int64, activeAt, expiresAt time.Time) error
}

type startupRuntimeStateStore interface {
	LoadPlatformRuntimeState(ctx context.Context) (string, time.Time, bool, error)
	SavePlatformRuntimeState(ctx context.Context, bootID string, heartbeatAt time.Time) error
	AcquireStartupRecoveryLock(ctx context.Context, ttl time.Duration) (*redislock.Lock, bool, error)
}

type StartupRuntimeRecoveryService struct {
	reconciler   startupRuntimeReconciler
	desired      startupRuntimeDesiredReconciler
	contests     startupRuntimeContestRepository
	instances    startupRuntimeInstanceRepository
	stateStore   startupRuntimeStateStore
	log          *zap.Logger
	now          func() time.Time
	bootIDPath   string
	heartbeatGap time.Duration
	lockTTL      time.Duration
	leaderRetry  time.Duration

	mu      sync.Mutex
	cancel  context.CancelFunc
	started bool
	ready   *startupRuntimeStartReady
	wg      sync.WaitGroup
}

type startupRuntimeStartReady struct {
	done chan struct{}
	err  error
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
	heartbeatInterval = startuprecovery.NormalizeHeartbeatInterval(heartbeatInterval)
	return &StartupRuntimeRecoveryService{
		reconciler:   reconciler,
		contests:     contests,
		instances:    instances,
		stateStore:   stateStore,
		log:          logger,
		now:          func() time.Time { return time.Now().UTC() },
		bootIDPath:   defaultBootIDPath,
		heartbeatGap: heartbeatInterval,
		lockTTL:      startuprecovery.DefaultLockTTL,
		leaderRetry:  startuprecovery.DefaultLeaderRetry,
	}
}

func (s *StartupRuntimeRecoveryService) SetDesiredRuntimeReconciler(reconciler startupRuntimeDesiredReconciler) *StartupRuntimeRecoveryService {
	if s == nil {
		return nil
	}
	s.desired = reconciler
	return s
}

func (s *StartupRuntimeRecoveryService) SetLockTTL(ttl time.Duration) *StartupRuntimeRecoveryService {
	if s == nil {
		return nil
	}
	if ttl > 0 {
		s.lockTTL = ttl
	}
	return s
}

func (s *StartupRuntimeRecoveryService) Start(ctx context.Context) (err error) {
	if ctx == nil {
		return fmt.Errorf("startup runtime recovery requires context")
	}

	s.mu.Lock()
	if s.started {
		ready := s.ready
		s.mu.Unlock()
		if ready != nil {
			return ready.wait(ctx)
		}
		return nil
	}
	runCtx, cancel := context.WithCancel(ctx)
	ready := &startupRuntimeStartReady{done: make(chan struct{})}
	s.cancel = cancel
	s.started = true
	s.ready = ready
	s.mu.Unlock()

	started := false
	defer func() {
		ready.complete(err)
		if started {
			s.mu.Lock()
			if s.ready == ready {
				s.ready = nil
			}
			s.mu.Unlock()
			return
		}
		cancel()
		s.mu.Lock()
		s.cancel = nil
		s.started = false
		if s.ready == ready {
			s.ready = nil
		}
		s.mu.Unlock()
	}()

	currentBootID, err := s.readCurrentBootID()
	if err != nil {
		return err
	}

	initialLock, initialLockAcquired, err := s.tryAcquireStartupRecoveryLock(runCtx)
	if err != nil {
		return err
	}

	initReady := make(chan error, 1)
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runLeaderElectionLoop(runCtx, currentBootID, initialLock, initialLockAcquired, initReady)
	}()
	select {
	case err := <-initReady:
		if err != nil {
			return err
		}
	case <-runCtx.Done():
		return runCtx.Err()
	}
	started = true
	return nil
}

func (r *startupRuntimeStartReady) wait(ctx context.Context) error {
	if r == nil {
		return nil
	}
	select {
	case <-r.done:
		return r.err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (r *startupRuntimeStartReady) complete(err error) {
	if r == nil {
		return
	}
	r.err = err
	close(r.done)
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
		s.mu.Lock()
		s.cancel = nil
		s.started = false
		s.mu.Unlock()
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *StartupRuntimeRecoveryService) recoverFromRuntimeOutage(ctx context.Context, currentBootID, lastBootID string, lastHeartbeatAt, startedAt time.Time) error {
	bootIDChanged := strings.TrimSpace(lastBootID) != "" &&
		strings.TrimSpace(currentBootID) != "" &&
		strings.TrimSpace(lastBootID) != strings.TrimSpace(currentBootID)
	heartbeatStale := s.isRuntimeHeartbeatStale(lastHeartbeatAt, startedAt)
	s.log.Warn(
		"runtime_outage_detected_for_startup_recovery",
		zap.Time("last_heartbeat_at", lastHeartbeatAt),
		zap.Time("started_at", startedAt),
		zap.Duration("outage_duration", startedAt.Sub(lastHeartbeatAt)),
		zap.Bool("boot_id_changed", bootIDChanged),
		zap.Bool("heartbeat_stale", heartbeatStale),
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
	if s.desired != nil {
		if err := s.desired.ReconcileDesiredAWDInstances(recoveryCtx); err != nil {
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

func startupRuntimeContestEffectiveEndTime(contest *contestcontracts.Contest) time.Time {
	if contest == nil {
		return time.Time{}
	}
	return contest.EndTime.UTC().Add(time.Duration(contest.PausedSeconds) * time.Second)
}

func buildStartupRuntimeRecoveryKey(lastBootID string, lastHeartbeatAt time.Time) string {
	return strings.TrimSpace(lastBootID) + "|" + lastHeartbeatAt.UTC().Format(time.RFC3339Nano)
}

func (s *StartupRuntimeRecoveryService) shouldRecoverFromRuntimeOutage(lastBootID, currentBootID string, lastHeartbeatAt time.Time) bool {
	return !lastHeartbeatAt.IsZero() && startupRuntimeBootIDChanged(lastBootID, currentBootID)
}

func (s *StartupRuntimeRecoveryService) isRuntimeHeartbeatStale(lastHeartbeatAt, startedAt time.Time) bool {
	if lastHeartbeatAt.IsZero() || startedAt.IsZero() || !startedAt.After(lastHeartbeatAt) {
		return false
	}
	return startedAt.Sub(lastHeartbeatAt) > s.runtimeHeartbeatStaleThreshold()
}

func (s *StartupRuntimeRecoveryService) runtimeHeartbeatStaleThreshold() time.Duration {
	return startuprecovery.HeartbeatStaleThreshold(s.heartbeatGap)
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

func (s *StartupRuntimeRecoveryService) runLeaderElectionLoop(ctx context.Context, currentBootID string, initialLock *redislock.Lock, initialLockAcquired bool, initReady chan error) {
	lock := initialLock
	lockAcquired := initialLockAcquired
	for {
		if ctx.Err() != nil {
			return
		}
		if !lockAcquired {
			if initReady != nil {
				ready, err := s.observedLeaderReady(ctx, currentBootID)
				if err != nil {
					if ctx.Err() == nil {
						s.log.Warn("startup_runtime_recovery_ready_state_check_failed", zap.Error(err))
					}
				} else if ready {
					initReady <- nil
					initReady = nil
				}
			}
			acquiredLock, acquired, err := s.tryAcquireStartupRecoveryLock(ctx)
			if err != nil {
				if ctx.Err() == nil {
					s.log.Warn("startup_runtime_recovery_lock_acquire_failed", zap.Error(err))
				}
				if !s.waitLeaderRetry(ctx) {
					return
				}
				continue
			}
			if !acquired {
				if !s.waitLeaderRetry(ctx) {
					return
				}
				continue
			}
			lock = acquiredLock
			lockAcquired = true
		}

		if err := s.runAsLeader(ctx, currentBootID, lock, initReady); err != nil {
			if ctx.Err() == nil {
				s.log.Warn("startup_runtime_recovery_leader_init_failed", zap.Error(err))
			}
			if initReady != nil {
				return
			}
			lock = nil
			lockAcquired = false
			if !s.waitLeaderRetry(ctx) {
				return
			}
			continue
		}
		initReady = nil
		lock = nil
		lockAcquired = false
		if !s.waitLeaderRetry(ctx) {
			return
		}
	}
}

func (s *StartupRuntimeRecoveryService) runAsLeader(ctx context.Context, currentBootID string, lock *redislock.Lock, initReady chan error) (err error) {
	leaderCtx, stopKeepalive := lockkeepalive.Start(ctx, s.log, lock, "startup_runtime_recovery", s.startupRecoveryLockTTL())
	defer func() {
		stopKeepalive()
		s.releaseStartupRecoveryLock(ctx, lock)
	}()
	defer func() {
		if initReady != nil && err != nil {
			initReady <- err
		}
	}()

	if err = s.initializeLeader(leaderCtx, currentBootID); err != nil {
		return err
	}
	if initReady != nil {
		initReady <- nil
		initReady = nil
	}

	s.runHeartbeatLoop(leaderCtx, currentBootID)
	return nil
}

func (s *StartupRuntimeRecoveryService) initializeLeader(ctx context.Context, currentBootID string) error {
	startedAt := s.now()
	lastBootID, lastHeartbeatAt, ok, err := s.loadPreviousRuntimeState(ctx)
	if err != nil {
		return err
	}
	if ok && s.shouldRecoverFromRuntimeOutage(lastBootID, currentBootID, lastHeartbeatAt) {
		return s.recoverFromRuntimeOutage(ctx, currentBootID, lastBootID, lastHeartbeatAt, startedAt)
	}
	if ok && !lastHeartbeatAt.IsZero() && s.isRuntimeHeartbeatStale(lastHeartbeatAt, startedAt) {
		s.log.Warn("startup_runtime_recovery_leader_gap_detected",
			zap.Time("last_heartbeat_at", lastHeartbeatAt),
			zap.Time("started_at", startedAt),
			zap.Duration("gap", startedAt.Sub(lastHeartbeatAt)),
			zap.Bool("boot_id_changed", startupRuntimeBootIDChanged(lastBootID, currentBootID)))
	}
	return s.recordHeartbeat(ctx, currentBootID, startedAt)
}

func (s *StartupRuntimeRecoveryService) loadPreviousRuntimeState(ctx context.Context) (string, time.Time, bool, error) {
	if s == nil || s.stateStore == nil {
		return "", time.Time{}, false, nil
	}
	return s.stateStore.LoadPlatformRuntimeState(ctx)
}

func (s *StartupRuntimeRecoveryService) tryAcquireStartupRecoveryLock(ctx context.Context) (*redislock.Lock, bool, error) {
	if s == nil || s.stateStore == nil {
		return nil, true, nil
	}
	return s.stateStore.AcquireStartupRecoveryLock(ctx, s.startupRecoveryLockTTL())
}

func (s *StartupRuntimeRecoveryService) recordHeartbeat(ctx context.Context, bootID string, heartbeatAt time.Time) error {
	if s == nil || s.stateStore == nil {
		return nil
	}
	return s.stateStore.SavePlatformRuntimeState(ctx, bootID, heartbeatAt)
}

func (s *StartupRuntimeRecoveryService) startupRecoveryLockTTL() time.Duration {
	if s == nil || s.lockTTL <= 0 {
		return startuprecovery.DefaultLockTTL
	}
	return s.lockTTL
}

func (s *StartupRuntimeRecoveryService) startupLeaderRetryInterval() time.Duration {
	if s == nil {
		return startuprecovery.DefaultLeaderRetry
	}
	return startuprecovery.NormalizeLeaderRetry(s.leaderRetry)
}

func (s *StartupRuntimeRecoveryService) waitLeaderRetry(ctx context.Context) bool {
	interval := s.startupLeaderRetryInterval()
	select {
	case <-ctx.Done():
		return false
	case <-time.After(interval):
		return true
	}
}

func (s *StartupRuntimeRecoveryService) observedLeaderReady(ctx context.Context, currentBootID string) (bool, error) {
	lastBootID, lastHeartbeatAt, ok, err := s.loadPreviousRuntimeState(ctx)
	if err != nil || !ok || lastHeartbeatAt.IsZero() {
		return false, err
	}
	if !sameBootID(lastBootID, currentBootID) {
		return false, nil
	}
	return !s.isRuntimeHeartbeatStale(lastHeartbeatAt, s.now()), nil
}

func (s *StartupRuntimeRecoveryService) releaseStartupRecoveryLock(ctx context.Context, lock *redislock.Lock) {
	if lock == nil {
		return
	}
	if ctx == nil {
		return
	}

	releaseCtx := context.WithoutCancel(ctx)
	if timeout := s.startupRecoveryLockTTL(); timeout > 0 {
		var cancel context.CancelFunc
		releaseCtx, cancel = context.WithTimeout(releaseCtx, timeout)
		defer cancel()
	}

	released, err := lock.Release(releaseCtx)
	if err != nil {
		s.log.Error("startup_runtime_recovery_lock_release_failed",
			zap.String("lock_key", lock.Key(releaseCtx)),
			zap.Error(err))
		return
	}
	if !released && ctx.Err() == nil {
		s.log.Warn("startup_runtime_recovery_lock_already_lost",
			zap.String("lock_key", lock.Key(releaseCtx)))
	}
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

func startupRuntimeBootIDChanged(lastBootID, currentBootID string) bool {
	return strings.TrimSpace(lastBootID) != "" &&
		strings.TrimSpace(currentBootID) != "" &&
		!sameBootID(lastBootID, currentBootID)
}

func sameBootID(lastBootID, currentBootID string) bool {
	return strings.TrimSpace(lastBootID) == strings.TrimSpace(currentBootID)
}
