package jobs

import (
	"context"
	"net/http"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	rediskeys "ctf-platform/internal/pkg/redis"
	"ctf-platform/internal/pkg/redislock"
)

const (
	defaultAWDRoundAttackScore  = 50
	defaultAWDRoundDefenseScore = 50
)

type AWDRoundUpdater struct {
	repo       contestports.AWDRepository
	redis      *redislib.Client
	cfg        config.ContestAWDConfig
	flagSecret string
	injector   contestports.AWDFlagInjector
	httpClient *http.Client
	log        *zap.Logger
}

type awdServiceTargetKey struct {
	teamID      int64
	challengeID int64
}

type noopAWDFlagInjector struct {
	log *zap.Logger
}

func (i *noopAWDFlagInjector) InjectRoundFlags(_ context.Context, contest *model.Contest, round *model.AWDRound, assignments []contestports.AWDFlagAssignment) error {
	if i == nil || i.log == nil || contest == nil || round == nil {
		return nil
	}
	i.log.Debug("skip_awd_flag_injection",
		zap.Int64("contest_id", contest.ID),
		zap.Int64("round_id", round.ID),
		zap.Int("assignment_count", len(assignments)),
	)
	return nil
}

func NewAWDRoundUpdater(
	repo contestports.AWDRepository,
	redis *redislib.Client,
	cfg config.ContestAWDConfig,
	flagSecret string,
	injector contestports.AWDFlagInjector,
	log *zap.Logger,
) *AWDRoundUpdater {
	if log == nil {
		log = zap.NewNop()
	}
	if injector == nil {
		injector = &noopAWDFlagInjector{log: log.Named("awd_flag_injector")}
	}
	return &AWDRoundUpdater{
		repo:       repo,
		redis:      redis,
		cfg:        cfg,
		flagSecret: flagSecret,
		injector:   injector,
		httpClient: &http.Client{Timeout: normalizedAWDCheckerTimeout(cfg.CheckerTimeout)},
		log:        log,
	}
}

func (u *AWDRoundUpdater) Start(ctx context.Context) {
	u.UpdateRoundsAt(ctx, time.Now())

	ticker := time.NewTicker(u.cfg.SchedulerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.UpdateRoundsAt(ctx, time.Now())
		}
	}
}

func (u *AWDRoundUpdater) UpdateRoundsAt(ctx context.Context, now time.Time) {
	if u.repo == nil {
		return
	}
	lock, acquired, err := redislock.Acquire(ctx, u.redis, rediskeys.AWDSchedulerLockKey(), u.cfg.SchedulerLockTTL)
	if err != nil {
		u.log.Error("acquire_awd_scheduler_lock_failed", zap.Error(err))
		return
	}
	if !acquired {
		u.log.Debug("awd_scheduler_lock_held_elsewhere")
		return
	}
	if lock != nil {
		defer func() {
			released, releaseErr := lock.Release(ctx)
			if releaseErr != nil {
				u.log.Error("release_awd_scheduler_lock_failed", zap.String("lock_key", lock.Key()), zap.Error(releaseErr))
				return
			}
			if !released {
				u.log.Warn("awd_scheduler_lock_expired_before_release", zap.String("lock_key", lock.Key()))
			}
		}()
	}

	recentCutoff := now.Add(-u.cfg.RoundInterval)
	contests, err := u.repo.ListSchedulableAWDContests(ctx, now, recentCutoff, u.cfg.SchedulerBatchSize)
	if err != nil {
		u.log.Error("list_awd_contests_failed", zap.Error(err))
		return
	}

	for i := range contests {
		contestCopy := contests[i]
		u.syncContestRounds(ctx, &contestCopy, now)
	}
}

func (u *AWDRoundUpdater) syncContestRounds(ctx context.Context, contest *model.Contest, now time.Time) {
	activeRound, totalRounds, ok := u.calculateRoundPlan(contest, now)
	if !ok {
		return
	}

	lockRound := activeRound
	if lockRound == 0 {
		lockRound = totalRounds
	}
	if lockRound <= 0 {
		return
	}

	acquired, err := u.acquireRoundLock(ctx, contest.ID, lockRound)
	if err != nil {
		u.log.Error("acquire_awd_round_lock_failed", zap.Int64("contest_id", contest.ID), zap.Int("round_number", lockRound), zap.Error(err))
		return
	}
	if !acquired {
		return
	}

	if err := u.reconcileRounds(ctx, contest, activeRound, totalRounds); err != nil {
		u.log.Error("sync_awd_rounds_failed", zap.Int64("contest_id", contest.ID), zap.Int("active_round", activeRound), zap.Int("total_rounds", totalRounds), zap.Error(err))
		return
	}

	if err := u.syncRoundFlags(ctx, contest, activeRound, now); err != nil {
		u.log.Error("sync_awd_round_flags_failed", zap.Int64("contest_id", contest.ID), zap.Int("active_round", activeRound), zap.Error(err))
	}
	if err := u.syncRoundServiceChecks(ctx, contest, activeRound); err != nil {
		u.log.Error("sync_awd_service_checks_failed", zap.Int64("contest_id", contest.ID), zap.Int("active_round", activeRound), zap.Error(err))
	}
}

func (u *AWDRoundUpdater) EnsureActiveRoundMaterialized(ctx context.Context, contest *model.Contest, now time.Time) error {
	activeRound, totalRounds, ok := u.calculateRoundPlan(contest, now)
	if !ok || activeRound <= 0 {
		return gorm.ErrRecordNotFound
	}
	if err := u.reconcileRounds(ctx, contest, activeRound, totalRounds); err != nil {
		return err
	}
	return u.syncRoundFlags(ctx, contest, activeRound, now)
}

func (u *AWDRoundUpdater) SetHTTPClient(client *http.Client) {
	if u == nil || client == nil {
		return
	}
	u.httpClient = client
}

func (u *AWDRoundUpdater) SyncRoundServiceChecks(ctx context.Context, contest *model.Contest, activeRound int) error {
	return u.syncRoundServiceChecks(ctx, contest, activeRound)
}
