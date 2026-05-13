package commands

import (
	"context"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestports "ctf-platform/internal/module/contest/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type AWDService struct {
	repo              awdCommandRepository
	roundManager      contestports.AWDRoundManager
	stateStore        contestports.AWDRoundStateStore
	previewTokenStore contestports.AWDCheckerPreviewTokenStore
	scoreboardCache   contestports.ScoreboardCacheWriter
	contestRepo       contestports.ContestLookupRepository
	flagSecret        string
	awdConfig         config.ContestAWDConfig
	log               *zap.Logger
	eventBus          platformevents.Bus
	imageRepo         challengecontracts.ImageStore
	awdChallengeRepo  challengeports.AWDChallengeQueryRepository
	runtimeProbe      challengeports.ChallengeRuntimeProbe
}

type awdCommandRepository interface {
	contestports.AWDServiceCheckTxRunner
	contestports.AWDAttackLogTxRunner
	contestports.AWDServiceStore
	contestports.AWDRoundStore
	contestports.AWDTeamLookup
	contestports.AWDChallengeLookup
	contestports.AWDReadinessQuery
	contestports.AWDTeamServiceStore
	contestports.AWDAttackLogStore
}

func NewAWDService(
	repo awdCommandRepository,
	contestRepo contestports.ContestLookupRepository,
	stateStore contestports.AWDRoundStateStore,
	previewTokenStore contestports.AWDCheckerPreviewTokenStore,
	flagSecret string,
	awdConfig config.ContestAWDConfig,
	log *zap.Logger,
	roundManager contestports.AWDRoundManager,
	imageRepo challengecontracts.ImageStore,
	awdChallengeRepo challengeports.AWDChallengeQueryRepository,
	runtimeProbe challengeports.ChallengeRuntimeProbe,
	scoreboardCaches ...contestports.ScoreboardCacheWriter,
) *AWDService {
	if log == nil {
		log = zap.NewNop()
	}
	var scoreboardCache contestports.ScoreboardCacheWriter
	if len(scoreboardCaches) > 0 {
		scoreboardCache = scoreboardCaches[0]
	}
	return &AWDService{
		repo:              repo,
		roundManager:      roundManager,
		stateStore:        stateStore,
		previewTokenStore: previewTokenStore,
		scoreboardCache:   scoreboardCache,
		contestRepo:       contestRepo,
		flagSecret:        flagSecret,
		awdConfig:         awdConfig,
		log:               log,
		imageRepo:         imageRepo,
		awdChallengeRepo:  awdChallengeRepo,
		runtimeProbe:      runtimeProbe,
	}
}

func (s *AWDService) SetEventBus(bus platformevents.Bus) *AWDService {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}

func (s *AWDService) publishWeakEvent(ctx context.Context, evt platformevents.Event) {
	if s == nil || s.eventBus == nil {
		return
	}
	if err := s.eventBus.Publish(ctx, evt); err != nil {
		s.log.Warn("publish_contest_event_failed", zap.String("event", evt.Name), zap.Error(err))
	}
}
