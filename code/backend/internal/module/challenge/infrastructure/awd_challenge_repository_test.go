package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type awdChallengeRepositorySourceStub struct {
	findByIDFn func(context.Context, int64) (*challengeentity.AWDChallenge, error)
	listFn     func(context.Context, *challengecontracts.AWDChallengeQuery) ([]*challengeentity.AWDChallenge, int64, error)
	createFn   func(context.Context, *challengeentity.AWDChallenge) error
	updateFn   func(context.Context, *challengeentity.AWDChallenge) error
	deleteFn   func(context.Context, int64) error
}

func (s awdChallengeRepositorySourceStub) CreateAWDChallenge(ctx context.Context, challenge *challengeentity.AWDChallenge) error {
	if s.createFn != nil {
		return s.createFn(ctx, challenge)
	}
	return nil
}

func (s awdChallengeRepositorySourceStub) FindAWDChallengeByID(ctx context.Context, id int64) (*challengeentity.AWDChallenge, error) {
	return s.findByIDFn(ctx, id)
}

func (s awdChallengeRepositorySourceStub) UpdateAWDChallenge(ctx context.Context, challenge *challengeentity.AWDChallenge) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, challenge)
	}
	return nil
}

func (s awdChallengeRepositorySourceStub) DeleteAWDChallenge(ctx context.Context, id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func (s awdChallengeRepositorySourceStub) ListAWDChallenges(ctx context.Context, query *challengecontracts.AWDChallengeQuery) ([]*challengeentity.AWDChallenge, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, query)
	}
	return []*challengeentity.AWDChallenge{}, 0, nil
}

func TestAWDChallengeRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewAWDChallengeRepository(awdChallengeRepositorySourceStub{
		findByIDFn: func(context.Context, int64) (*challengeentity.AWDChallenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindAWDChallengeByID(context.Background(), 1); !errors.Is(err, challengeports.ErrAWDChallengeNotFound) {
		t.Fatalf("error = %v, want %v", err, challengeports.ErrAWDChallengeNotFound)
	}
}

func TestAWDChallengeRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewAWDChallengeRepository(awdChallengeRepositorySourceStub{
		findByIDFn: func(context.Context, int64) (*challengeentity.AWDChallenge, error) {
			return nil, expectedErr
		},
	})

	if _, err := repo.FindAWDChallengeByID(context.Background(), 1); !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}
