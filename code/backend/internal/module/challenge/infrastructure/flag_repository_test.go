package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type flagRepositorySourceStub struct {
	findByIDFn func(context.Context, int64) (*model.Challenge, error)
	updateFn   func(context.Context, *model.Challenge) error
}

func (s flagRepositorySourceStub) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	return s.findByIDFn(ctx, id)
}

func (s flagRepositorySourceStub) Update(ctx context.Context, challenge *model.Challenge) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, challenge)
	}
	return nil
}

func TestFlagRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewFlagRepository(flagRepositorySourceStub{
		findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, challengeports.ErrChallengeFlagChallengeNotFound) {
		t.Fatalf("error = %v, want %v", err, challengeports.ErrChallengeFlagChallengeNotFound)
	}
}

func TestFlagRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewFlagRepository(flagRepositorySourceStub{
		findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, expectedErr
		},
	})

	if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}
