package jobs

import (
	"context"
	"net/http"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

const (
	defaultAWDRoundAttackScore  = contestdomain.AWDDefaultRoundAttackScore
	defaultAWDRoundDefenseScore = contestdomain.AWDDefaultRoundDefenseScore
)

type AWDRoundUpdater struct {
	repo            contestports.AWDRoundUpdateRepository
	redis           *redislib.Client
	scoreboardCache contestports.ScoreboardCacheWriter
	cfg             config.ContestAWDConfig
	flagSecret      string
	injector        contestports.AWDFlagInjector
	checkerRunner   contestports.CheckerRunner
	httpClient      *http.Client
	log             *zap.Logger
}

type awdServiceTargetKey struct {
	teamID    int64
	serviceID int64
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
	repo contestports.AWDRoundUpdateRepository,
	redis *redislib.Client,
	cfg config.ContestAWDConfig,
	flagSecret string,
	injector contestports.AWDFlagInjector,
	log *zap.Logger,
	scoreboardCaches ...contestports.ScoreboardCacheWriter,
) *AWDRoundUpdater {
	if log == nil {
		log = zap.NewNop()
	}
	if injector == nil {
		injector = &noopAWDFlagInjector{log: log.Named("awd_flag_injector")}
	}
	var scoreboardCache contestports.ScoreboardCacheWriter
	if len(scoreboardCaches) > 0 {
		scoreboardCache = scoreboardCaches[0]
	}
	return &AWDRoundUpdater{
		repo:            repo,
		redis:           redis,
		scoreboardCache: scoreboardCache,
		cfg:             cfg,
		flagSecret:      flagSecret,
		injector:        injector,
		httpClient:      &http.Client{Timeout: normalizedAWDCheckerTimeout(cfg.CheckerTimeout)},
		log:             log,
	}
}

func (u *AWDRoundUpdater) Start(ctx context.Context) {
	u.UpdateRoundsAt(ctx, time.Now().UTC())

	ticker := time.NewTicker(u.cfg.SchedulerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			u.UpdateRoundsAt(ctx, time.Now().UTC())
		}
	}
}
