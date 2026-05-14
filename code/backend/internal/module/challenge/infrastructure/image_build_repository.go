package infrastructure

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type imageBuildRepositorySource interface {
	challengeports.ImageCommandRepository
	challengeports.ImageBuildJobRepository
}

type ImageBuildRepository struct {
	commands *ImageCommandRepository
	source   imageBuildRepositorySource
}

func NewImageBuildRepository(source imageBuildRepositorySource) *ImageBuildRepository {
	if source == nil {
		return nil
	}
	return &ImageBuildRepository{
		commands: NewImageCommandRepository(source),
		source:   source,
	}
}

func (r *ImageBuildRepository) Create(ctx context.Context, image *model.Image) error {
	return r.commands.Create(ctx, image)
}

func (r *ImageBuildRepository) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	return r.commands.FindByID(ctx, id)
}

func (r *ImageBuildRepository) FindByNameTag(ctx context.Context, name, tag string) (*model.Image, error) {
	return r.commands.FindByNameTag(ctx, name, tag)
}

func (r *ImageBuildRepository) Update(ctx context.Context, image *model.Image) error {
	return r.commands.Update(ctx, image)
}

func (r *ImageBuildRepository) Delete(ctx context.Context, id int64) error {
	return r.commands.Delete(ctx, id)
}

func (r *ImageBuildRepository) CreateImageBuildJob(ctx context.Context, job *model.ImageBuildJob) error {
	return r.source.CreateImageBuildJob(ctx, job)
}

func (r *ImageBuildRepository) FindImageBuildJobByID(ctx context.Context, id int64) (*model.ImageBuildJob, error) {
	return r.source.FindImageBuildJobByID(ctx, id)
}

func (r *ImageBuildRepository) ListPendingImageBuildJobs(ctx context.Context, limit int) ([]*model.ImageBuildJob, error) {
	return r.source.ListPendingImageBuildJobs(ctx, limit)
}

func (r *ImageBuildRepository) TryStartImageBuildJob(ctx context.Context, id int64, startedAt time.Time) (bool, error) {
	return r.source.TryStartImageBuildJob(ctx, id, startedAt)
}

func (r *ImageBuildRepository) UpdateImageBuildJob(ctx context.Context, job *model.ImageBuildJob) error {
	return r.source.UpdateImageBuildJob(ctx, job)
}

var _ challengeports.ImageCommandRepository = (*ImageBuildRepository)(nil)
var _ challengeports.ImageBuildJobRepository = (*ImageBuildRepository)(nil)
