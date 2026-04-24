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

type ctxOnlyContestChallengeContract struct{}

func (ctxOnlyContestChallengeContract) FindByIDWithContext(context.Context, int64) (*model.Challenge, error) {
	return nil, nil
}

func (ctxOnlyContestChallengeContract) BatchGetSolvedStatusWithContext(context.Context, int64, []int64) (map[int64]bool, error) {
	return nil, nil
}

func (ctxOnlyContestChallengeContract) BatchGetSolvedCountWithContext(context.Context, []int64) (map[int64]int64, error) {
	return nil, nil
}

type ctxOnlyChallengeContract struct{}

func (ctxOnlyChallengeContract) FindByIDWithContext(context.Context, int64) (*model.Challenge, error) {
	return nil, nil
}

func (ctxOnlyChallengeContract) BatchGetSolvedStatusWithContext(context.Context, int64, []int64) (map[int64]bool, error) {
	return nil, nil
}

func (ctxOnlyChallengeContract) BatchGetSolvedCountWithContext(context.Context, []int64) (map[int64]int64, error) {
	return nil, nil
}

func (ctxOnlyChallengeContract) FindChallengeTopologyByChallengeIDWithContext(context.Context, int64) (*model.ChallengeTopology, error) {
	return nil, nil
}

func (ctxOnlyChallengeContract) FindPublishedForRecommendationWithContext(context.Context, int, []string, []int64) ([]*model.Challenge, error) {
	return nil, nil
}

var _ challengecontracts.ImageStore = (*ctxOnlyImageStore)(nil)
var _ challengecontracts.ChallengeContract = (*ctxOnlyChallengeContract)(nil)
var _ challengecontracts.ContestChallengeContract = (*ctxOnlyContestChallengeContract)(nil)
var _ challengecontracts.PracticeChallengeContract = (*ctxOnlyPracticeChallengeContract)(nil)
