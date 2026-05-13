package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ImageQueryRepository struct {
	source challengeports.ImageQueryRepository
}

func NewImageQueryRepository(source challengeports.ImageQueryRepository) *ImageQueryRepository {
	if source == nil {
		return nil
	}
	return &ImageQueryRepository{source: source}
}

func (r *ImageQueryRepository) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	image, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeImageNotFound
	}
	return image, err
}

func (r *ImageQueryRepository) List(ctx context.Context, name, status string, offset, limit int) ([]*model.Image, int64, error) {
	return r.source.List(ctx, name, status, offset, limit)
}

var _ challengeports.ImageQueryRepository = (*ImageQueryRepository)(nil)
