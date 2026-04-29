package commands

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestports "ctf-platform/internal/module/contest/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type AWDService struct {
	repo             contestports.AWDRepository
	roundManager     contestports.AWDRoundManager
	redis            *redislib.Client
	contestRepo      contestports.ContestLookupRepository
	flagSecret       string
	awdConfig        config.ContestAWDConfig
	log              *zap.Logger
	eventBus         platformevents.Bus
	broadcaster      contestports.RealtimeBroadcaster
	imageRepo        challengecontracts.ImageStore
	awdChallengeRepo challengeports.AWDChallengeQueryRepository
	runtimeProbe     challengeports.ChallengeRuntimeProbe
}

func NewAWDService(
	repo contestports.AWDRepository,
	contestRepo contestports.ContestLookupRepository,
	redis *redislib.Client,
	flagSecret string,
	awdConfig config.ContestAWDConfig,
	log *zap.Logger,
	roundManager contestports.AWDRoundManager,
	imageRepo challengecontracts.ImageStore,
	awdChallengeRepo challengeports.AWDChallengeQueryRepository,
	runtimeProbe challengeports.ChallengeRuntimeProbe,
) *AWDService {
	if log == nil {
		log = zap.NewNop()
	}
	return &AWDService{
		repo:             repo,
		roundManager:     roundManager,
		redis:            redis,
		contestRepo:      contestRepo,
		flagSecret:       flagSecret,
		awdConfig:        awdConfig,
		log:              log,
		imageRepo:        imageRepo,
		awdChallengeRepo: awdChallengeRepo,
		runtimeProbe:     runtimeProbe,
	}
}

func (s *AWDService) SetEventBus(bus platformevents.Bus) *AWDService {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}

func (s *AWDService) SetRealtimeBroadcaster(broadcaster contestports.RealtimeBroadcaster) *AWDService {
	if s == nil {
		return nil
	}
	s.broadcaster = broadcaster
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
