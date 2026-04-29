package ports_test

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyAWDChallengeQueryRepository struct{}

func (ctxOnlyAWDChallengeQueryRepository) FindAWDChallengeByID(context.Context, int64) (*model.AWDChallenge, error) {
	return nil, nil
}

func (ctxOnlyAWDChallengeQueryRepository) ListAWDChallenges(context.Context, *dto.AWDChallengeQuery) ([]*model.AWDChallenge, int64, error) {
	return nil, 0, nil
}

var _ challengeports.AWDChallengeQueryRepository = (*ctxOnlyAWDChallengeQueryRepository)(nil)
