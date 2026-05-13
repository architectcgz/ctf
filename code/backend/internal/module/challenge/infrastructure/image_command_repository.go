package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ImageCommandRepository struct {
	source challengeports.ImageCommandRepository
}

func NewImageCommandRepository(source challengeports.ImageCommandRepository) *ImageCommandRepository {
	if source == nil {
		return nil
	}
	return &ImageCommandRepository{source: source}
}

func (r *ImageCommandRepository) Create(ctx context.Context, image *model.Image) error {
	return r.source.Create(ctx, image)
}

func (r *ImageCommandRepository) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	image, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeImageNotFound
	}
	return image, err
}

func (r *ImageCommandRepository) FindByNameTag(ctx context.Context, name, tag string) (*model.Image, error) {
	image, err := r.source.FindByNameTag(ctx, name, tag)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeImageNotFound
	}
	return image, err
}

func (r *ImageCommandRepository) Update(ctx context.Context, image *model.Image) error {
	return r.source.Update(ctx, image)
}

func (r *ImageCommandRepository) Delete(ctx context.Context, id int64) error {
	return r.source.Delete(ctx, id)
}

var _ challengeports.ImageCommandRepository = (*ImageCommandRepository)(nil)
