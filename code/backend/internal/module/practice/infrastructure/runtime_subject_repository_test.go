package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type runtimeSubjectSourceStub struct {
	findByIDFn                  func(context.Context, int64) (*model.Challenge, error)
	findChallengeTopologyByIDFn func(context.Context, int64) (*challengecontracts.PracticeRuntimeChallengeTopology, error)
}

func (s runtimeSubjectSourceStub) FindPracticeRuntimeChallengeByID(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
	challenge, err := s.findByIDFn(ctx, id)
	if err != nil || challenge == nil {
		return nil, err
	}
	return &challengecontracts.PracticeRuntimeChallenge{
		ID:              challenge.ID,
		PackageSlug:     challenge.PackageSlug,
		Title:           challenge.Title,
		Category:        challenge.Category,
		Difficulty:      challenge.Difficulty,
		Points:          challenge.Points,
		ImageID:         challenge.ImageID,
		Status:          string(challenge.Status),
		FlagType:        challenge.FlagType,
		FlagHash:        challenge.FlagHash,
		FlagSalt:        challenge.FlagSalt,
		FlagRegex:       challenge.FlagRegex,
		FlagPrefix:      challenge.FlagPrefix,
		InstanceSharing: string(challenge.InstanceSharing),
		TargetProtocol:  challenge.TargetProtocol,
		TargetPort:      challenge.TargetPort,
	}, nil
}

func (s runtimeSubjectSourceStub) FindPracticeRuntimeChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*challengecontracts.PracticeRuntimeChallengeTopology, error) {
	return s.findChallengeTopologyByIDFn(ctx, challengeID)
}

func TestRuntimeSubjectRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewRuntimeSubjectRepository(runtimeSubjectSourceStub{
		findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findChallengeTopologyByIDFn: func(context.Context, int64) (*challengecontracts.PracticeRuntimeChallengeTopology, error) {
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
		findChallengeTopologyByIDFn: func(context.Context, int64) (*challengecontracts.PracticeRuntimeChallengeTopology, error) {
			return &challengecontracts.PracticeRuntimeChallengeTopology{ChallengeID: 1}, nil
		},
	})

	_, err := repo.FindByID(context.Background(), 1)
	if !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}
