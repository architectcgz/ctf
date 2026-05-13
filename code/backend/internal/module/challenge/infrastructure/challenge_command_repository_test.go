package infrastructure

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type challengeCommandRepositorySourceStub struct {
	findByIDFn                       func(context.Context, int64) (*model.Challenge, error)
	findPublishCheckJobByIDFn        func(context.Context, int64) (*model.ChallengePublishCheckJob, error)
	findActivePublishCheckJobByIDFn  func(context.Context, int64) (*model.ChallengePublishCheckJob, error)
	findLatestPublishCheckJobByIDFn  func(context.Context, int64) (*model.ChallengePublishCheckJob, error)
}

func (s challengeCommandRepositorySourceStub) CreateWithHints(context.Context, *model.Challenge, []*model.ChallengeHint) error {
	return nil
}

func (s challengeCommandRepositorySourceStub) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	return s.findByIDFn(ctx, id)
}

func (s challengeCommandRepositorySourceStub) Update(context.Context, *model.Challenge) error {
	return nil
}

func (s challengeCommandRepositorySourceStub) UpdateWithHints(context.Context, *model.Challenge, []*model.ChallengeHint, bool) error {
	return nil
}

func (s challengeCommandRepositorySourceStub) Delete(context.Context, int64) error {
	return nil
}

func (s challengeCommandRepositorySourceStub) HasRunningInstances(context.Context, int64) (bool, error) {
	return false, nil
}

func (s challengeCommandRepositorySourceStub) CreatePublishCheckJob(context.Context, *model.ChallengePublishCheckJob) error {
	return nil
}

func (s challengeCommandRepositorySourceStub) FindPublishCheckJobByID(ctx context.Context, id int64) (*model.ChallengePublishCheckJob, error) {
	return s.findPublishCheckJobByIDFn(ctx, id)
}

func (s challengeCommandRepositorySourceStub) FindActivePublishCheckJobByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error) {
	return s.findActivePublishCheckJobByIDFn(ctx, challengeID)
}

func (s challengeCommandRepositorySourceStub) FindLatestPublishCheckJobByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePublishCheckJob, error) {
	return s.findLatestPublishCheckJobByIDFn(ctx, challengeID)
}

func (s challengeCommandRepositorySourceStub) ListPendingPublishCheckJobs(context.Context, int) ([]*model.ChallengePublishCheckJob, error) {
	return nil, nil
}

func (s challengeCommandRepositorySourceStub) TryStartPublishCheckJob(context.Context, int64, time.Time) (bool, error) {
	return false, nil
}

func (s challengeCommandRepositorySourceStub) UpdatePublishCheckJob(context.Context, *model.ChallengePublishCheckJob) error {
	return nil
}

func TestChallengeCommandRepositoryMapsRawNotFoundToPortsSentinels(t *testing.T) {
	t.Parallel()

	repo := NewChallengeCommandRepository(challengeCommandRepositorySourceStub{
		findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findPublishCheckJobByIDFn: func(context.Context, int64) (*model.ChallengePublishCheckJob, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findActivePublishCheckJobByIDFn: func(context.Context, int64) (*model.ChallengePublishCheckJob, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findLatestPublishCheckJobByIDFn: func(context.Context, int64) (*model.ChallengePublishCheckJob, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	cases := []struct {
		name string
		run  func() error
		want error
	}{
		{
			name: "challenge lookup",
			run: func() error {
				_, err := repo.FindByID(context.Background(), 1)
				return err
			},
			want: challengeports.ErrChallengeCommandChallengeNotFound,
		},
		{
			name: "publish check by id lookup",
			run: func() error {
				_, err := repo.FindPublishCheckJobByID(context.Background(), 1)
				return err
			},
			want: challengeports.ErrChallengePublishCheckJobNotFound,
		},
		{
			name: "active publish check lookup",
			run: func() error {
				_, err := repo.FindActivePublishCheckJobByChallengeID(context.Background(), 1)
				return err
			},
			want: challengeports.ErrChallengePublishCheckJobNotFound,
		},
		{
			name: "latest publish check lookup",
			run: func() error {
				_, err := repo.FindLatestPublishCheckJobByChallengeID(context.Background(), 1)
				return err
			},
			want: challengeports.ErrChallengePublishCheckJobNotFound,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.run(); !errors.Is(err, tc.want) {
				t.Fatalf("error = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestChallengeCommandRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewChallengeCommandRepository(challengeCommandRepositorySourceStub{
		findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, expectedErr
		},
		findPublishCheckJobByIDFn: func(context.Context, int64) (*model.ChallengePublishCheckJob, error) {
			return nil, expectedErr
		},
		findActivePublishCheckJobByIDFn: func(context.Context, int64) (*model.ChallengePublishCheckJob, error) {
			return nil, expectedErr
		},
		findLatestPublishCheckJobByIDFn: func(context.Context, int64) (*model.ChallengePublishCheckJob, error) {
			return nil, expectedErr
		},
	})

	cases := []struct {
		name string
		run  func() error
	}{
		{
			name: "challenge lookup",
			run: func() error {
				_, err := repo.FindByID(context.Background(), 1)
				return err
			},
		},
		{
			name: "publish check by id lookup",
			run: func() error {
				_, err := repo.FindPublishCheckJobByID(context.Background(), 1)
				return err
			},
		},
		{
			name: "active publish check lookup",
			run: func() error {
				_, err := repo.FindActivePublishCheckJobByChallengeID(context.Background(), 1)
				return err
			},
		},
		{
			name: "latest publish check lookup",
			run: func() error {
				_, err := repo.FindLatestPublishCheckJobByChallengeID(context.Background(), 1)
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
