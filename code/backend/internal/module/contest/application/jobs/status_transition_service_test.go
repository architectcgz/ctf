package jobs

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

type statusTransitionRepoStub struct {
	result         contestdomain.ContestStatusTransitionResult
	err            error
	lastTransition contestdomain.ContestStatusTransition
	applyCallCount int
}

func (s *statusTransitionRepoStub) ApplyStatusTransition(_ context.Context, transition contestdomain.ContestStatusTransition) (contestdomain.ContestStatusTransitionResult, error) {
	s.applyCallCount++
	s.lastTransition = transition
	if s.result.Transition.ContestID == 0 {
		s.result.Transition = transition
	}
	return s.result, s.err
}

func TestContestStatusTransitionServiceApplyValidTransition(t *testing.T) {
	repo := &statusTransitionRepoStub{
		result: contestdomain.ContestStatusTransitionResult{
			Applied:       true,
			StatusVersion: 3,
		},
	}
	service := newContestStatusTransitionService(repo)

	result, err := service.Apply(context.Background(), contestdomain.ContestStatusTransition{
		ContestID:         12,
		FromStatus:        model.ContestStatusRunning,
		ToStatus:          model.ContestStatusFrozen,
		FromStatusVersion: 2,
		OccurredAt:        time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !result.Applied || result.StatusVersion != 3 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if repo.applyCallCount != 1 {
		t.Fatalf("expected one repo call, got %d", repo.applyCallCount)
	}
}

func TestContestStatusTransitionServiceApplyAllowsFrozenRollback(t *testing.T) {
	repo := &statusTransitionRepoStub{
		result: contestdomain.ContestStatusTransitionResult{
			Applied:       true,
			StatusVersion: 5,
		},
	}
	service := newContestStatusTransitionService(repo)

	result, err := service.Apply(context.Background(), contestdomain.ContestStatusTransition{
		ContestID:         12,
		FromStatus:        model.ContestStatusFrozen,
		ToStatus:          model.ContestStatusRunning,
		FromStatusVersion: 4,
		OccurredAt:        time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("Apply() error = %v", err)
	}
	if !result.Applied || result.StatusVersion != 5 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if repo.lastTransition.FromStatus != model.ContestStatusFrozen || repo.lastTransition.ToStatus != model.ContestStatusRunning {
		t.Fatalf("unexpected transition forwarded to repo: %+v", repo.lastTransition)
	}
}

func TestContestStatusTransitionServiceRejectsInvalidTransition(t *testing.T) {
	repo := &statusTransitionRepoStub{}
	service := newContestStatusTransitionService(repo)

	_, err := service.Apply(context.Background(), contestdomain.ContestStatusTransition{
		ContestID:         12,
		FromStatus:        model.ContestStatusRegistration,
		ToStatus:          model.ContestStatusEnded,
		FromStatusVersion: 0,
		OccurredAt:        time.Now().UTC(),
	})
	if err != contestdomain.ErrInvalidStatusTransition {
		t.Fatalf("expected ErrInvalidStatusTransition, got %v", err)
	}
	if repo.applyCallCount != 0 {
		t.Fatalf("expected repo not to be called, got %d", repo.applyCallCount)
	}
}
