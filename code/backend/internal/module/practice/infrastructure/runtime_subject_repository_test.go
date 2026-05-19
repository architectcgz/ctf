package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type runtimeSubjectSourceStub struct {
	findByIDFn                  func(context.Context, int64) (*model.Challenge, error)
	findChallengeTopologyByIDFn func(context.Context, int64) (*challengeentity.ChallengeTopology, error)
}

func (s runtimeSubjectSourceStub) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	return s.findByIDFn(ctx, id)
}

func (s runtimeSubjectSourceStub) FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*challengeentity.ChallengeTopology, error) {
	return s.findChallengeTopologyByIDFn(ctx, challengeID)
}

func TestRuntimeSubjectRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewRuntimeSubjectRepository(runtimeSubjectSourceStub{
		findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findChallengeTopologyByIDFn: func(context.Context, int64) (*challengeentity.ChallengeTopology, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, practiceports.ErrPracticeChallengeNotFound) {
		t.Fatalf("challenge error = %v, want %v", err, practiceports.ErrPracticeChallengeNotFound)
	}
	if _, err := repo.FindChallengeTopologyByChallengeID(context.Background(), 1); !errors.Is(err, practiceports.ErrPracticeChallengeTopologyNotFound) {
		t.Fatalf("topology error = %v, want %v", err, practiceports.ErrPracticeChallengeTopologyNotFound)
	}
}

func TestRuntimeSubjectRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewRuntimeSubjectRepository(runtimeSubjectSourceStub{
		findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, expectedErr
		},
		findChallengeTopologyByIDFn: func(context.Context, int64) (*challengeentity.ChallengeTopology, error) {
			return &challengeentity.ChallengeTopology{ChallengeID: 1}, nil
		},
	})

	_, err := repo.FindByID(context.Background(), 1)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}
