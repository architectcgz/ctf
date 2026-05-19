package ports_test

import (
	"context"

	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyImageRepository struct{}

func (ctxOnlyImageRepository) Create(context.Context, *challengeentity.Image) error {
	return nil
}

func (ctxOnlyImageRepository) FindByID(context.Context, int64) (*challengeentity.Image, error) {
	return nil, nil
}

func (ctxOnlyImageRepository) FindByNameTag(context.Context, string, string) (*challengeentity.Image, error) {
	return nil, nil
}

func (ctxOnlyImageRepository) List(context.Context, string, string, int, int) ([]*challengeentity.Image, int64, error) {
	return nil, 0, nil
}

func (ctxOnlyImageRepository) Update(context.Context, *challengeentity.Image) error {
	return nil
}

func (ctxOnlyImageRepository) Delete(context.Context, int64) error {
	return nil
}

var _ challengeports.ImageCommandRepository = (*ctxOnlyImageRepository)(nil)
var _ challengeports.ImageQueryRepository = (*ctxOnlyImageRepository)(nil)
