package ports_test

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyChallengeQueryRepository struct{}

func (ctxOnlyChallengeQueryRepository) FindByID(context.Context, int64) (*model.Challenge, error) {
	return nil, nil
}

func (ctxOnlyChallengeQueryRepository) ListWithContext(context.Context, *dto.ChallengeQuery) ([]*model.Challenge, int64, error) {
	return nil, 0, nil
}

func (ctxOnlyChallengeQueryRepository) ListHintsByChallengeID(context.Context, int64) ([]*model.ChallengeHint, error) {
	return nil, nil
}

func (ctxOnlyChallengeQueryRepository) GetSolvedStatus(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func (ctxOnlyChallengeQueryRepository) GetSolvedCount(context.Context, int64) (int64, error) {
	return 0, nil
}

func (ctxOnlyChallengeQueryRepository) GetTotalAttempts(context.Context, int64) (int64, error) {
	return 0, nil
}

func (ctxOnlyChallengeQueryRepository) BatchGetSolvedStatus(context.Context, int64, []int64) (map[int64]bool, error) {
	return nil, nil
}

func (ctxOnlyChallengeQueryRepository) BatchGetSolvedCount(context.Context, []int64) (map[int64]int64, error) {
	return nil, nil
}

func (ctxOnlyChallengeQueryRepository) BatchGetTotalAttempts(context.Context, []int64) (map[int64]int64, error) {
	return nil, nil
}

func (ctxOnlyChallengeQueryRepository) ListPublished(context.Context, *dto.ChallengeQuery) ([]*model.Challenge, int64, error) {
	return nil, 0, nil
}

var _ challengeports.ChallengeQueryRepository = (*ctxOnlyChallengeQueryRepository)(nil)
