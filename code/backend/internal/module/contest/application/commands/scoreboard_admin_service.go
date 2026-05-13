package commands

import (
	"ctf-platform/internal/config"
	"ctf-platform/internal/module/contest/application/statusmachine"
	contestports "ctf-platform/internal/module/contest/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type ScoreboardAdminService struct {
	repo        contestports.ContestScoreboardAdminRepository
	transition  contestCommandStatusTransitionRepository
	sideEffects *statusmachine.SideEffectRunner
	stateStore  contestports.ContestScoreboardStateStore
	cfg         *config.ContestConfig
	eventBus    platformevents.Bus
}

func NewScoreboardAdminService(repo contestports.ContestScoreboardAdminRepository, stateStore contestports.ContestScoreboardStateStore, cfg *config.ContestConfig) *ScoreboardAdminService {
	var transitionRepo contestCommandStatusTransitionRepository
	if typedRepo, ok := any(repo).(contestCommandStatusTransitionRepository); ok {
		transitionRepo = typedRepo
	}
	return &ScoreboardAdminService{
		repo:        repo,
		transition:  transitionRepo,
		sideEffects: statusmachine.NewSideEffectRunner(nil),
		stateStore:  stateStore,
		cfg:         cfg,
	}
}

func (s *ScoreboardAdminService) SetStatusSideEffectStore(store contestports.ContestStatusSideEffectStore) *ScoreboardAdminService {
	if s == nil {
		return nil
	}
	s.sideEffects = statusmachine.NewSideEffectRunner(store)
	return s
}

func (s *ScoreboardAdminService) SetEventBus(bus platformevents.Bus) *ScoreboardAdminService {
	if s == nil {
		return nil
	}
	s.eventBus = bus
	return s
}
