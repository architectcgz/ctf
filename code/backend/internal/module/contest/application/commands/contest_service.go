package commands

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"ctf-platform/internal/model"
	"ctf-platform/internal/module/contest/application/statusmachine"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

type ContestService struct {
	repo           contestCommandRepository
	transitionRepo contestCommandStatusTransitionRepository
	sideEffects    *statusmachine.SideEffectRunner
	awdRepo        contestports.AWDReadinessQuery
	log            *zap.Logger
}

type contestCommandRepository interface {
	contestports.ContestLookupRepository
	contestports.ContestWriteRepository
}

type contestCommandStatusTransitionRepository interface {
	UpdateContestWithStatusTransition(ctx context.Context, contest *model.Contest, transition contestdomain.ContestStatusTransition) (contestdomain.ContestStatusTransitionResult, error)
	MarkTransitionSideEffectsSucceeded(ctx context.Context, id int64) error
	MarkTransitionSideEffectsFailed(ctx context.Context, id int64, cause error) error
}

func NewContestService(repo contestCommandRepository, awdRepo contestports.AWDReadinessQuery, redis *redislib.Client, log *zap.Logger) *ContestService {
	if log == nil {
		log = zap.NewNop()
	}
	var transitionRepo contestCommandStatusTransitionRepository
	if typedRepo, ok := any(repo).(contestCommandStatusTransitionRepository); ok {
		transitionRepo = typedRepo
	}
	return &ContestService{
		repo:           repo,
		transitionRepo: transitionRepo,
		sideEffects:    statusmachine.NewSideEffectRunner(redis),
		awdRepo:        awdRepo,
		log:            log,
	}
}
