package ports_test

import (
	"context"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyAWDChallengeQueryRepository struct{}

func (ctxOnlyAWDChallengeQueryRepository) FindAWDChallengeByID(context.Context, int64) (*challengeentity.AWDChallenge, error) {
	return nil, nil
}

func (ctxOnlyAWDChallengeQueryRepository) ListAWDChallenges(context.Context, *challengecontracts.AWDChallengeQuery) ([]*challengeentity.AWDChallenge, int64, error) {
	return nil, 0, nil
}

var _ challengeports.AWDChallengeQueryRepository = (*ctxOnlyAWDChallengeQueryRepository)(nil)
