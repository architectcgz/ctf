package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type topologyServiceRepositoryStub struct {
	findByIDFn     func(context.Context, int64) (*model.Challenge, error)
	findTopologyFn func(context.Context, int64) (*model.ChallengeTopology, error)
	upsertFn       func(context.Context, *model.ChallengeTopology) error
	deleteFn       func(context.Context, int64) error
}

func (s topologyServiceRepositoryStub) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	return s.findByIDFn(ctx, id)
}

func (s topologyServiceRepositoryStub) FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	return s.findTopologyFn(ctx, challengeID)
}

func (s topologyServiceRepositoryStub) UpsertChallengeTopology(ctx context.Context, topology *model.ChallengeTopology) error {
	if s.upsertFn != nil {
		return s.upsertFn(ctx, topology)
	}
	return nil
}

func (s topologyServiceRepositoryStub) DeleteChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, challengeID)
	}
	return nil
}

type topologyTemplateRepositoryStub struct {
	createFn         func(context.Context, *model.EnvironmentTemplate) error
	updateFn         func(context.Context, *model.EnvironmentTemplate) error
	deleteFn         func(context.Context, int64) error
	findByIDFn       func(context.Context, int64) (*model.EnvironmentTemplate, error)
	listFn           func(context.Context, string) ([]*model.EnvironmentTemplate, error)
	incrementUsageFn func(context.Context, int64) error
}

func (s topologyTemplateRepositoryStub) Create(ctx context.Context, template *model.EnvironmentTemplate) error {
	if s.createFn != nil {
		return s.createFn(ctx, template)
	}
	return nil
}

func (s topologyTemplateRepositoryStub) Update(ctx context.Context, template *model.EnvironmentTemplate) error {
	if s.updateFn != nil {
		return s.updateFn(ctx, template)
	}
	return nil
}

func (s topologyTemplateRepositoryStub) Delete(ctx context.Context, id int64) error {
	if s.deleteFn != nil {
		return s.deleteFn(ctx, id)
	}
	return nil
}

func (s topologyTemplateRepositoryStub) FindByID(ctx context.Context, id int64) (*model.EnvironmentTemplate, error) {
	return s.findByIDFn(ctx, id)
}

func (s topologyTemplateRepositoryStub) List(ctx context.Context, keyword string) ([]*model.EnvironmentTemplate, error) {
	if s.listFn != nil {
		return s.listFn(ctx, keyword)
	}
	return nil, nil
}

func (s topologyTemplateRepositoryStub) IncrementUsage(ctx context.Context, id int64) error {
	if s.incrementUsageFn != nil {
		return s.incrementUsageFn(ctx, id)
	}
	return nil
}

type topologyPackageRevisionRepositoryStub struct {
	createFn     func(context.Context, *model.ChallengePackageRevision) error
	findByIDFn   func(context.Context, int64) (*model.ChallengePackageRevision, error)
	findLatestFn func(context.Context, int64) (*model.ChallengePackageRevision, error)
	listFn       func(context.Context, int64) ([]*model.ChallengePackageRevision, error)
}

func (s topologyPackageRevisionRepositoryStub) CreateChallengePackageRevision(ctx context.Context, revision *model.ChallengePackageRevision) error {
	if s.createFn != nil {
		return s.createFn(ctx, revision)
	}
	return nil
}

func (s topologyPackageRevisionRepositoryStub) FindChallengePackageRevisionByID(ctx context.Context, id int64) (*model.ChallengePackageRevision, error) {
	return s.findByIDFn(ctx, id)
}

func (s topologyPackageRevisionRepositoryStub) FindLatestChallengePackageRevisionByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengePackageRevision, error) {
	return s.findLatestFn(ctx, challengeID)
}

func (s topologyPackageRevisionRepositoryStub) ListChallengePackageRevisionsByChallengeID(ctx context.Context, challengeID int64) ([]*model.ChallengePackageRevision, error) {
	if s.listFn != nil {
		return s.listFn(ctx, challengeID)
	}
	return nil, nil
}

func TestTopologyServiceRepositoryMapsRawNotFoundToPortsSentinels(t *testing.T) {
	t.Parallel()

	repo := NewTopologyServiceRepository(topologyServiceRepositoryStub{
		findByIDFn: func(context.Context, int64) (*model.Challenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findTopologyFn: func(context.Context, int64) (*model.ChallengeTopology, error) {
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
			want: challengeports.ErrChallengeTopologyChallengeNotFound,
		},
		{
			name: "topology lookup",
			run: func() error {
				_, err := repo.FindChallengeTopologyByChallengeID(context.Background(), 1)
				return err
			},
			want: challengeports.ErrChallengeTopologyNotFound,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.run(); !errors.Is(err, tc.want) {
				t.Fatalf("expected %v, got %v", tc.want, err)
			}
		})
	}
}

func TestTopologyTemplateRepositoryMapsRawNotFoundToPortsSentinel(t *testing.T) {
	t.Parallel()

	repo := NewTopologyTemplateRepository(topologyTemplateRepositoryStub{
		findByIDFn: func(context.Context, int64) (*model.EnvironmentTemplate, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindByID(context.Background(), 1); !errors.Is(err, challengeports.ErrChallengeTopologyTemplateNotFound) {
		t.Fatalf("expected %v, got %v", challengeports.ErrChallengeTopologyTemplateNotFound, err)
	}
}

func TestTopologyPackageRevisionRepositoryMapsRawNotFoundToPortsSentinel(t *testing.T) {
	t.Parallel()

	repo := NewTopologyPackageRevisionRepository(topologyPackageRevisionRepositoryStub{
		findByIDFn: func(context.Context, int64) (*model.ChallengePackageRevision, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findLatestFn: func(context.Context, int64) (*model.ChallengePackageRevision, error) {
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
				_, err := repo.FindChallengePackageRevisionByID(context.Background(), 1)
				return err
			},
		},
		{
			name: "find latest by challenge id",
			run: func() error {
				_, err := repo.FindLatestChallengePackageRevisionByChallengeID(context.Background(), 1)
				return err
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := tc.run(); !errors.Is(err, challengeports.ErrChallengeTopologyPackageRevisionNotFound) {
				t.Fatalf("expected %v, got %v", challengeports.ErrChallengeTopologyPackageRevisionNotFound, err)
			}
		})
	}
}
