package ports_test

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyChallengeQueryRepository struct{}

func (ctxOnlyChallengeQueryRepository) FindByIDWithContext(context.Context, int64) (*model.Challenge, error) {
	return nil, nil
}

func (ctxOnlyChallengeQueryRepository) ListWithContext(context.Context, *dto.ChallengeQuery) ([]*model.Challenge, int64, error) {
	return nil, 0, nil
}

func (ctxOnlyChallengeQueryRepository) ListHintsByChallengeIDWithContext(context.Context, int64) ([]*model.ChallengeHint, error) {
	return nil, nil
}

func (ctxOnlyChallengeQueryRepository) GetSolvedStatus(context.Context, int64, int64) (bool, error) {
	return false, nil
}

func (ctxOnlyChallengeQueryRepository) GetSolvedCountWithContext(context.Context, int64) (int64, error) {
	return 0, nil
}

func (ctxOnlyChallengeQueryRepository) GetTotalAttemptsWithContext(context.Context, int64) (int64, error) {
	return 0, nil
}

func (ctxOnlyChallengeQueryRepository) BatchGetSolvedStatusWithContext(context.Context, int64, []int64) (map[int64]bool, error) {
	return nil, nil
}

func (ctxOnlyChallengeQueryRepository) BatchGetSolvedCountWithContext(context.Context, []int64) (map[int64]int64, error) {
	return nil, nil
}

func (ctxOnlyChallengeQueryRepository) BatchGetTotalAttemptsWithContext(context.Context, []int64) (map[int64]int64, error) {
	return nil, nil
}

func (ctxOnlyChallengeQueryRepository) ListPublishedWithContext(context.Context, *dto.ChallengeQuery) ([]*model.Challenge, int64, error) {
	return nil, 0, nil
}

var _ challengeports.ChallengeQueryRepository = (*ctxOnlyChallengeQueryRepository)(nil)
