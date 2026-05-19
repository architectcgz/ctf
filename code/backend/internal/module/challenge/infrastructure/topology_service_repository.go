package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type topologyServiceRawRepository interface {
	challengeports.ChallengeTopologyChallengeLookupRepository
	challengeports.ChallengeTopologyReadRepository
	challengeports.ChallengeTopologyWriteRepository
}

type TopologyServiceRepository struct {
	source topologyServiceRawRepository
}

func NewTopologyServiceRepository(source topologyServiceRawRepository) *TopologyServiceRepository {
	if source == nil {
		return nil
	}
	return &TopologyServiceRepository{source: source}
}

func (r *TopologyServiceRepository) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	item, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeTopologyChallengeNotFound
	}
	return item, err
}

func (r *TopologyServiceRepository) FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*challengeentity.ChallengeTopology, error) {
	item, err := r.source.FindChallengeTopologyByChallengeID(ctx, challengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeTopologyNotFound
	}
	return item, err
}

func (r *TopologyServiceRepository) UpsertChallengeTopology(ctx context.Context, topology *challengeentity.ChallengeTopology) error {
	return r.source.UpsertChallengeTopology(ctx, topology)
}

func (r *TopologyServiceRepository) DeleteChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) error {
	return r.source.DeleteChallengeTopologyByChallengeID(ctx, challengeID)
}

var _ challengeports.ChallengeTopologyChallengeLookupRepository = (*TopologyServiceRepository)(nil)
var _ challengeports.ChallengeTopologyReadRepository = (*TopologyServiceRepository)(nil)
var _ challengeports.ChallengeTopologyWriteRepository = (*TopologyServiceRepository)(nil)

type topologyTemplateRawRepository interface {
	challengeports.EnvironmentTemplateCommandRepository
	challengeports.EnvironmentTemplateQueryRepository
	challengeports.EnvironmentTemplateUsageRepository
}

type TopologyTemplateRepository struct {
	source topologyTemplateRawRepository
}

func NewTopologyTemplateRepository(source topologyTemplateRawRepository) *TopologyTemplateRepository {
	if source == nil {
		return nil
	}
	return &TopologyTemplateRepository{source: source}
}

func (r *TopologyTemplateRepository) Create(ctx context.Context, template *challengeentity.EnvironmentTemplate) error {
	return r.source.Create(ctx, template)
}

func (r *TopologyTemplateRepository) Update(ctx context.Context, template *challengeentity.EnvironmentTemplate) error {
	return r.source.Update(ctx, template)
}

func (r *TopologyTemplateRepository) Delete(ctx context.Context, id int64) error {
	return r.source.Delete(ctx, id)
}

func (r *TopologyTemplateRepository) FindByID(ctx context.Context, id int64) (*challengeentity.EnvironmentTemplate, error) {
	item, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeTopologyTemplateNotFound
	}
	return item, err
}

func (r *TopologyTemplateRepository) List(ctx context.Context, keyword string) ([]*challengeentity.EnvironmentTemplate, error) {
	return r.source.List(ctx, keyword)
}

func (r *TopologyTemplateRepository) IncrementUsage(ctx context.Context, id int64) error {
	return r.source.IncrementUsage(ctx, id)
}

var _ challengeports.EnvironmentTemplateCommandRepository = (*TopologyTemplateRepository)(nil)
var _ challengeports.EnvironmentTemplateQueryRepository = (*TopologyTemplateRepository)(nil)
var _ challengeports.EnvironmentTemplateUsageRepository = (*TopologyTemplateRepository)(nil)

type TopologyPackageRevisionRepository struct {
	source challengeports.ChallengePackageRevisionRepository
}

func NewTopologyPackageRevisionRepository(source challengeports.ChallengePackageRevisionRepository) *TopologyPackageRevisionRepository {
	if source == nil {
		return nil
	}
	return &TopologyPackageRevisionRepository{source: source}
}

func (r *TopologyPackageRevisionRepository) CreateChallengePackageRevision(ctx context.Context, revision *challengeentity.ChallengePackageRevision) error {
	return r.source.CreateChallengePackageRevision(ctx, revision)
}

func (r *TopologyPackageRevisionRepository) FindChallengePackageRevisionByID(ctx context.Context, id int64) (*challengeentity.ChallengePackageRevision, error) {
	item, err := r.source.FindChallengePackageRevisionByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeTopologyPackageRevisionNotFound
	}
	return item, err
}

func (r *TopologyPackageRevisionRepository) FindLatestChallengePackageRevisionByChallengeID(ctx context.Context, challengeID int64) (*challengeentity.ChallengePackageRevision, error) {
	item, err := r.source.FindLatestChallengePackageRevisionByChallengeID(ctx, challengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeTopologyPackageRevisionNotFound
	}
	return item, err
}

func (r *TopologyPackageRevisionRepository) ListChallengePackageRevisionsByChallengeID(ctx context.Context, challengeID int64) ([]*challengeentity.ChallengePackageRevision, error) {
	return r.source.ListChallengePackageRevisionsByChallengeID(ctx, challengeID)
}

var _ challengeports.ChallengePackageRevisionRepository = (*TopologyPackageRevisionRepository)(nil)
