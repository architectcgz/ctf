package ports_test

import (
	"context"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyImageRepository struct{}

func (ctxOnlyImageRepository) Create(context.Context, *model.Image) error {
	return nil
}

func (ctxOnlyImageRepository) FindByID(context.Context, int64) (*model.Image, error) {
	return nil, nil
}

func (ctxOnlyImageRepository) FindByNameTag(context.Context, string, string) (*model.Image, error) {
	return nil, nil
}

func (ctxOnlyImageRepository) List(context.Context, string, string, int, int) ([]*model.Image, int64, error) {
	return nil, 0, nil
}

func (ctxOnlyImageRepository) Update(context.Context, *model.Image) error {
	return nil
}

func (ctxOnlyImageRepository) DeleteWithContext(context.Context, int64) error {
	return nil
}

var _ challengeports.ImageRepository = (*ctxOnlyImageRepository)(nil)
