package commands

import (
	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
	"ctf-platform/internal/module/contest/application/statusmachine"
	contestports "ctf-platform/internal/module/contest/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type ScoreboardAdminService struct {
	repo        contestports.ContestScoreboardAdminRepository
	transition  contestCommandStatusTransitionRepository
	sideEffects *statusmachine.SideEffectRunner
	redis       *redislib.Client
	cfg         *config.ContestConfig
	eventBus    platformevents.Bus
}

func NewScoreboardAdminService(repo contestports.ContestScoreboardAdminRepository, redis *redislib.Client, cfg *config.ContestConfig) *ScoreboardAdminService {
	var transitionRepo contestCommandStatusTransitionRepository
	if typedRepo, ok := any(repo).(contestCommandStatusTransitionRepository); ok {
		transitionRepo = typedRepo
	}
	return &ScoreboardAdminService{
		repo:        repo,
		transition:  transitionRepo,
		sideEffects: statusmachine.NewSideEffectRunner(redis),
		redis:       redis,
		cfg:         cfg,
	}
}

func (s *ScoreboardAdminService) SetEventBus(bus platformevents.Bus) *ScoreboardAdminService {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}
