package ports_test

import (
	"context"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyTagRepository struct{}

func (ctxOnlyTagRepository) Create(context.Context, *model.Tag) error {
	return nil
}

func (ctxOnlyTagRepository) List(context.Context, string) ([]*model.Tag, error) {
	return nil, nil
}

func (ctxOnlyTagRepository) FindByIDs(context.Context, []int64) ([]*model.Tag, error) {
	return nil, nil
}

func (ctxOnlyTagRepository) AttachTagsInTx(context.Context, int64, []int64) error {
	return nil
}

func (ctxOnlyTagRepository) DetachFromChallenge(context.Context, int64, int64) error {
	return nil
}

func (ctxOnlyTagRepository) FindByChallengeID(context.Context, int64) ([]*model.Tag, error) {
	return nil, nil
}

func (ctxOnlyTagRepository) DeleteWithContext(context.Context, int64) error {
	return nil
}

func (ctxOnlyTagRepository) CountChallengesByTagID(context.Context, int64) (int64, error) {
	return 0, nil
}

var _ challengeports.TagRepository = (*ctxOnlyTagRepository)(nil)
