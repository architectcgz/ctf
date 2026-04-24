package ports_test

import (
	"context"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyImageRepository struct{}

func (ctxOnlyImageRepository) CreateWithContext(context.Context, *model.Image) error {
	return nil
}

func (ctxOnlyImageRepository) FindByIDWithContext(context.Context, int64) (*model.Image, error) {
	return nil, nil
}

func (ctxOnlyImageRepository) FindByNameTagWithContext(context.Context, string, string) (*model.Image, error) {
	return nil, nil
}

func (ctxOnlyImageRepository) ListWithContext(context.Context, string, string, int, int) ([]*model.Image, int64, error) {
	return nil, 0, nil
}

func (ctxOnlyImageRepository) UpdateWithContext(context.Context, *model.Image) error {
	return nil
}

func (ctxOnlyImageRepository) DeleteWithContext(context.Context, int64) error {
	return nil
}

var _ challengeports.ImageRepository = (*ctxOnlyImageRepository)(nil)
