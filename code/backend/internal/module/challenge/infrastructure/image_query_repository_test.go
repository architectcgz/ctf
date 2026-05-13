package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type imageQuerySourceStub struct {
	findByIDFn func(context.Context, int64) (*model.Image, error)
	listFn     func(context.Context, string, string, int, int) ([]*model.Image, int64, error)
}

func (s imageQuerySourceStub) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	return s.findByIDFn(ctx, id)
}

func (s imageQuerySourceStub) List(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error) {
	if s.listFn != nil {
		return s.listFn(ctx, name, status, offset, limit)
	}
	return []*model.Image{}, 0, nil
}

func TestImageQueryRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewImageQueryRepository(imageQuerySourceStub{
		findByIDFn: func(context.Context, int64) (*model.Image, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, challengeports.ErrChallengeImageNotFound) {
		t.Fatalf("error = %v, want %v", err, challengeports.ErrChallengeImageNotFound)
	}
}

func TestImageQueryRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewImageQueryRepository(imageQuerySourceStub{
		findByIDFn: func(context.Context, int64) (*model.Image, error) {
			return nil, expectedErr
		},
	})

	if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}
