package jobs

import (
	"context"
	"time"

	contestdomain "ctf-platform/internal/module/contest/domain"
)

const contestStatusUpdaterAppliedBy = "contest_status_updater"

type contestStatusTransitioner interface {
	Apply(ctx context.Context, transition contestdomain.ContestStatusTransition) (contestdomain.ContestStatusTransitionResult, error)
}

type contestStatusTransitionRecorder interface {
	MarkTransitionSideEffectsSucceeded(ctx context.Context, id int64) error
	MarkTransitionSideEffectsFailed(ctx context.Context, id int64, cause error) error
}

type contestStatusTransitionReplayer interface {
	ListTransitionsForSideEffectReplay(ctx context.Context, limit int) ([]contestdomain.ContestStatusTransitionResult, error)
}

type contestStatusTransitionService struct {
	repo contestStatusTransitionRepository
}

type contestStatusTransitionRepository interface {
	ApplyStatusTransition(ctx context.Context, transition contestdomain.ContestStatusTransition) (contestdomain.ContestStatusTransitionResult, error)
}

func newContestStatusTransitionService(repo contestStatusTransitionRepository) *contestStatusTransitionService {
	return &contestStatusTransitionService{repo: repo}
}

func (s *contestStatusTransitionService) Apply(ctx context.Context, transition contestdomain.ContestStatusTransition) (contestdomain.ContestStatusTransitionResult, error) {
	if transition.ContestID <= 0 || transition.FromStatus == "" || transition.ToStatus == "" {
		return contestdomain.ContestStatusTransitionResult{Transition: transition}, contestdomain.ErrInvalidStatusTransition
	}
	if err := contestdomain.ValidateStatusTransition(transition.FromStatus, transition.ToStatus); err != nil {
		return contestdomain.ContestStatusTransitionResult{Transition: transition}, contestdomain.ErrInvalidStatusTransition
	}
	if transition.OccurredAt.IsZero() {
		transition.OccurredAt = time.Now().UTC()
	} else {
		transition.OccurredAt = transition.OccurredAt.UTC()
	}

	return s.repo.ApplyStatusTransition(ctx, transition)
}
