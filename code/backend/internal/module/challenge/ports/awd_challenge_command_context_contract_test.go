package ports_test

import (
	"context"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyAWDChallengeCommandRepository struct{}

func (ctxOnlyAWDChallengeCommandRepository) CreateAWDChallenge(context.Context, *model.AWDChallenge) error {
	return nil
}

func (ctxOnlyAWDChallengeCommandRepository) FindAWDChallengeByID(context.Context, int64) (*model.AWDChallenge, error) {
	return nil, nil
}

func (ctxOnlyAWDChallengeCommandRepository) UpdateAWDChallenge(context.Context, *model.AWDChallenge) error {
	return nil
}

func (ctxOnlyAWDChallengeCommandRepository) DeleteAWDChallenge(context.Context, int64) error {
	return nil
}

var _ challengeports.AWDChallengeCommandRepository = (*ctxOnlyAWDChallengeCommandRepository)(nil)
