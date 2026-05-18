package ports_test

import (
	"context"

	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyAWDChallengeCommandRepository struct{}

func (ctxOnlyAWDChallengeCommandRepository) CreateAWDChallenge(context.Context, *challengeentity.AWDChallenge) error {
	return nil
}

func (ctxOnlyAWDChallengeCommandRepository) FindAWDChallengeByID(context.Context, int64) (*challengeentity.AWDChallenge, error) {
	return nil, nil
}

func (ctxOnlyAWDChallengeCommandRepository) UpdateAWDChallenge(context.Context, *challengeentity.AWDChallenge) error {
	return nil
}

func (ctxOnlyAWDChallengeCommandRepository) DeleteAWDChallenge(context.Context, int64) error {
	return nil
}

var _ challengeports.AWDChallengeCommandRepository = (*ctxOnlyAWDChallengeCommandRepository)(nil)
