package ports_test

import (
	"context"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyTagRepository struct{}

func (ctxOnlyTagRepository) CreateWithContext(context.Context, *model.Tag) error {
	return nil
}

func (ctxOnlyTagRepository) ListWithContext(context.Context, string) ([]*model.Tag, error) {
	return nil, nil
}

func (ctxOnlyTagRepository) FindByIDsWithContext(context.Context, []int64) ([]*model.Tag, error) {
	return nil, nil
}

func (ctxOnlyTagRepository) AttachTagsInTxWithContext(context.Context, int64, []int64) error {
	return nil
}

func (ctxOnlyTagRepository) DetachFromChallengeWithContext(context.Context, int64, int64) error {
	return nil
}

func (ctxOnlyTagRepository) FindByChallengeIDWithContext(context.Context, int64) ([]*model.Tag, error) {
	return nil, nil
}

func (ctxOnlyTagRepository) DeleteWithContext(context.Context, int64) error {
	return nil
}

func (ctxOnlyTagRepository) CountChallengesByTagIDWithContext(context.Context, int64) (int64, error) {
	return 0, nil
}

var _ challengeports.TagRepository = (*ctxOnlyTagRepository)(nil)
