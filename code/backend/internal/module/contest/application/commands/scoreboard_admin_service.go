package commands

import (
	"context"

	"ctf-platform/internal/config"
	"ctf-platform/internal/module/contest/application/statusmachine"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type ScoreboardAdminService struct {
	repo        contestports.ContestScoreboardAdminRepository
	transition  contestCommandStatusTransitionRepository
	realtimeTx  contestScoreboardRealtimeTransactionRepository
	sideEffects *statusmachine.SideEffectRunner
	stateStore  contestports.ContestScoreboardStateStore
	cfg         *config.ContestConfig
	eventBus    platformevents.Bus
	outbox      contestports.ContestRealtimeOutboxRepository
}

func NewScoreboardAdminService(repo contestports.ContestScoreboardAdminRepository, stateStore contestports.ContestScoreboardStateStore, cfg *config.ContestConfig) *ScoreboardAdminService {
	var transitionRepo contestCommandStatusTransitionRepository
	if typedRepo, ok := any(repo).(contestCommandStatusTransitionRepository); ok {
		transitionRepo = typedRepo
	}
	var realtimeTxRepo contestScoreboardRealtimeTransactionRepository
	if typedRepo, ok := any(repo).(contestScoreboardRealtimeTransactionRepository); ok {
		realtimeTxRepo = typedRepo
	}
	return &ScoreboardAdminService{
		repo:        repo,
		transition:  transitionRepo,
		realtimeTx:  realtimeTxRepo,
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

func (s *ScoreboardAdminService) SetRealtimeOutbox(repo contestports.ContestRealtimeOutboxRepository) *ScoreboardAdminService {
	if s == nil {
		return nil
	}
	s.outbox = repo
	return s
}

type contestScoreboardRealtimeTransactionRepository interface {
	UpdateContestWithRealtimeRelay(ctx context.Context, contest *contestentity.Contest, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) error
	UpdateContestWithStatusTransitionAndRealtimeRelay(ctx context.Context, contest *contestentity.Contest, transition contestdomain.ContestStatusTransition, relay contestcontracts.RealtimeRelayEvent, dedupeKey string) (contestdomain.ContestStatusTransitionResult, error)
}
