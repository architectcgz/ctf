package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestports "ctf-platform/internal/module/contest/ports"
)

type awdPreviewRuntimeChallengeSourceStub struct {
	findByIDFn func(context.Context, int64) (*challengecontracts.AWDChallenge, error)
}

func (s awdPreviewRuntimeChallengeSourceStub) FindAWDChallengeByID(ctx context.Context, id int64) (*challengecontracts.AWDChallenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &challengecontracts.AWDChallenge{ID: id}, nil
}

func (s awdPreviewRuntimeChallengeSourceStub) ListAWDChallenges(context.Context, *challengecontracts.AWDChallengeQuery) ([]*challengecontracts.AWDChallenge, int64, error) {
	return nil, 0, nil
}

func TestAWDPreviewRuntimeChallengeRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewAWDPreviewRuntimeChallengeRepository(awdPreviewRuntimeChallengeSourceStub{
		findByIDFn: func(context.Context, int64) (*challengecontracts.AWDChallenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindAWDChallengeByID(context.Background(), 1); !errors.Is(err, contestports.ErrContestAWDPreviewChallengeNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestAWDPreviewChallengeNotFound)
	}
}

func TestAWDPreviewRuntimeLookupRepositoriesPassThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	challengeRepo := NewAWDPreviewRuntimeChallengeRepository(awdPreviewRuntimeChallengeSourceStub{
		findByIDFn: func(context.Context, int64) (*challengecontracts.AWDChallenge, error) {
			return nil, expectedErr
		},
	})
	if _, err := challengeRepo.FindAWDChallengeByID(context.Background(), 1); !errors.Is(err, expectedErr) {
		t.Fatalf("challenge error = %v, want %v", err, expectedErr)
	}
}
