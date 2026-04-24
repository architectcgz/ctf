package contracts_test

import (
	"context"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/model"
)

type ctxOnlyImageStore struct{}

func (ctxOnlyImageStore) FindByIDWithContext(context.Context, int64) (*model.Image, error) {
	return nil, nil
}

type ctxOnlyPracticeChallengeContract struct{}

func (ctxOnlyPracticeChallengeContract) FindByIDWithContext(context.Context, int64) (*model.Challenge, error) {
	return nil, nil
}

func (ctxOnlyPracticeChallengeContract) FindChallengeTopologyByChallengeIDWithContext(context.Context, int64) (*model.ChallengeTopology, error) {
	return nil, nil
}

var _ challengecontracts.ImageStore = (*ctxOnlyImageStore)(nil)
var _ challengecontracts.PracticeChallengeContract = (*ctxOnlyPracticeChallengeContract)(nil)
