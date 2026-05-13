package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type awdPreviewRuntimeChallengeSourceStub struct {
	findByIDFn func(context.Context, int64) (*model.AWDChallenge, error)
}

func (s awdPreviewRuntimeChallengeSourceStub) FindAWDChallengeByID(ctx context.Context, id int64) (*model.AWDChallenge, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &model.AWDChallenge{ID: id}, nil
}

func (s awdPreviewRuntimeChallengeSourceStub) ListAWDChallenges(context.Context, *dto.AWDChallengeQuery) ([]*model.AWDChallenge, int64, error) {
	return nil, 0, nil
}

type awdPreviewRuntimeImageSourceStub struct {
	findByIDFn func(context.Context, int64) (*model.Image, error)
}

func (s awdPreviewRuntimeImageSourceStub) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return &model.Image{ID: id}, nil
}

func TestAWDPreviewRuntimeChallengeRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewAWDPreviewRuntimeChallengeRepository(awdPreviewRuntimeChallengeSourceStub{
		findByIDFn: func(context.Context, int64) (*model.AWDChallenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindAWDChallengeByID(context.Background(), 1); !errors.Is(err, contestports.ErrContestAWDPreviewChallengeNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestAWDPreviewChallengeNotFound)
	}
}

func TestAWDPreviewRuntimeImageRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewAWDPreviewRuntimeImageRepository(awdPreviewRuntimeImageSourceStub{
		findByIDFn: func(context.Context, int64) (*model.Image, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, contestports.ErrContestAWDPreviewImageNotFound) {
		t.Fatalf("error = %v, want %v", err, contestports.ErrContestAWDPreviewImageNotFound)
	}
}

func TestAWDPreviewRuntimeLookupRepositoriesPassThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	challengeRepo := NewAWDPreviewRuntimeChallengeRepository(awdPreviewRuntimeChallengeSourceStub{
		findByIDFn: func(context.Context, int64) (*model.AWDChallenge, error) {
			return nil, expectedErr
		},
	})
	imageRepo := NewAWDPreviewRuntimeImageRepository(awdPreviewRuntimeImageSourceStub{
		findByIDFn: func(context.Context, int64) (*model.Image, error) {
			return nil, expectedErr
		},
	})

	if _, err := challengeRepo.FindAWDChallengeByID(context.Background(), 1); !errors.Is(err, expectedErr) {
		t.Fatalf("challenge error = %v, want %v", err, expectedErr)
	}
	if _, err := imageRepo.FindByID(context.Background(), 2); !errors.Is(err, expectedErr) {
		t.Fatalf("image error = %v, want %v", err, expectedErr)
	}
}
