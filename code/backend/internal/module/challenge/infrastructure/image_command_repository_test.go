package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type imageCommandSourceStub struct {
	findByIDFn      func(context.Context, int64) (*model.Image, error)
	findByNameTagFn func(context.Context, string, string) (*model.Image, error)
}

func (s imageCommandSourceStub) Create(context.Context, *model.Image) error { return nil }

func (s imageCommandSourceStub) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	return s.findByIDFn(ctx, id)
}

func (s imageCommandSourceStub) FindByNameTag(ctx context.Context, name, tag string) (*model.Image, error) {
	return s.findByNameTagFn(ctx, name, tag)
}

func (s imageCommandSourceStub) Update(context.Context, *model.Image) error { return nil }

func (s imageCommandSourceStub) Delete(context.Context, int64) error { return nil }

func TestImageCommandRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewImageCommandRepository(imageCommandSourceStub{
		findByIDFn: func(context.Context, int64) (*model.Image, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findByNameTagFn: func(context.Context, string, string) (*model.Image, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	cases := []struct {
		name string
		run  func() error
	}{
		{
			name: "find by id",
			run: func() error {
				_, err := repo.FindByID(context.Background(), 1)
				return err
			},
		},
		{
			name: "find by name tag",
			run: func() error {
				_, err := repo.FindByNameTag(context.Background(), "web", "latest")
				return err
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.run(); !errors.Is(err, challengeports.ErrChallengeImageNotFound) {
				t.Fatalf("error = %v, want %v", err, challengeports.ErrChallengeImageNotFound)
			}
		})
	}
}

func TestImageCommandRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewImageCommandRepository(imageCommandSourceStub{
		findByIDFn: func(context.Context, int64) (*model.Image, error) {
			return nil, expectedErr
		},
		findByNameTagFn: func(context.Context, string, string) (*model.Image, error) {
			return nil, expectedErr
		},
	})

	cases := []struct {
		name string
		run  func() error
	}{
		{
			name: "find by id",
			run: func() error {
				_, err := repo.FindByID(context.Background(), 1)
				return err
			},
		},
		{
			name: "find by name tag",
			run: func() error {
				_, err := repo.FindByNameTag(context.Background(), "web", "latest")
				return err
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.run(); !errors.Is(err, expectedErr) {
				t.Fatalf("error = %v, want %v", err, expectedErr)
			}
		})
	}
}
