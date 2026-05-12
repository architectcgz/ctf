package jobs

import (
	"context"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/module/contest/application/statusmachine"
	contestports "ctf-platform/internal/module/contest/ports"
)

type StatusUpdater struct {
	repo         contestports.ContestStatusRepository
	transitioner contestStatusTransitioner
	recorder     contestStatusTransitionRecorder
	replayer     contestStatusTransitionReplayer
	sideEffects  *statusmachine.SideEffectRunner
	awdRepo      contestports.AWDReadinessQuery
	redis        *redislib.Client
	log          *zap.Logger
	interval     time.Duration
	batchSize    int
	lockTTL      time.Duration
}

func NewStatusUpdater(repo contestports.ContestStatusRepository, redis *redislib.Client, interval time.Duration, batchSize int, lockTTL time.Duration, log *zap.Logger, awdRepos ...contestports.AWDReadinessQuery) *StatusUpdater {
	if log == nil {
		log = zap.NewNop()
	}
	if lockTTL <= 0 {
		lockTTL = 30 * time.Second
	}
	var awdRepo contestports.AWDReadinessQuery
	if len(awdRepos) > 0 {
		awdRepo = awdRepos[0]
	}
	var recorder contestStatusTransitionRecorder
	if repoRecorder, ok := any(repo).(contestStatusTransitionRecorder); ok {
		recorder = repoRecorder
	}
	var replayer contestStatusTransitionReplayer
	if repoReplayer, ok := any(repo).(contestStatusTransitionReplayer); ok {
		replayer = repoReplayer
	}
	return &StatusUpdater{
		repo:         repo,
		transitioner: newContestStatusTransitionService(repo),
		recorder:     recorder,
		replayer:     replayer,
		sideEffects:  statusmachine.NewSideEffectRunner(nil),
		awdRepo:      awdRepo,
		redis:        redis,
		log:          log,
		interval:     interval,
		batchSize:    batchSize,
		lockTTL:      lockTTL,
	}
}

func (u *StatusUpdater) SetStatusSideEffectStore(store contestports.ContestStatusSideEffectStore) *StatusUpdater {
	if u == nil {
		return nil
	}
	u.sideEffects = statusmachine.NewSideEffectRunner(store)
	return u
}

func (u *StatusUpdater) Start(ctx context.Context) {
	u.updateStatuses(ctx)

	ticker := time.NewTicker(u.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.updateStatuses(ctx)
		}
	}
}
