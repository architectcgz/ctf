package contracts_test

import (
	"context"

	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
)

type ctxOnlyImageStore struct{}

func (ctxOnlyImageStore) FindByID(context.Context, int64) (*model.Image, error) {
	return nil, nil
}

type ctxOnlyPracticeChallengeContract struct{}

func (ctxOnlyPracticeChallengeContract) FindByID(context.Context, int64) (*model.Challenge, error) {
	return nil, nil
}

func (ctxOnlyPracticeChallengeContract) FindChallengeTopologyByChallengeID(context.Context, int64) (*model.ChallengeTopology, error) {
	return nil, nil
}

type ctxOnlyContestChallengeContract struct{}

func (ctxOnlyContestChallengeContract) FindByID(context.Context, int64) (*model.Challenge, error) {
	return nil, nil
}

func (ctxOnlyContestChallengeContract) BatchGetSolvedStatus(context.Context, int64, []int64) (map[int64]bool, error) {
	return nil, nil
}

func (ctxOnlyContestChallengeContract) BatchGetSolvedCount(context.Context, []int64) (map[int64]int64, error) {
	return nil, nil
}

type ctxOnlyChallengeContract struct{}

func (ctxOnlyChallengeContract) FindByID(context.Context, int64) (*model.Challenge, error) {
	return nil, nil
}

func (ctxOnlyChallengeContract) BatchGetSolvedStatus(context.Context, int64, []int64) (map[int64]bool, error) {
	return nil, nil
}

func (ctxOnlyChallengeContract) BatchGetSolvedCount(context.Context, []int64) (map[int64]int64, error) {
	return nil, nil
}

func (ctxOnlyChallengeContract) FindChallengeTopologyByChallengeID(context.Context, int64) (*model.ChallengeTopology, error) {
	return nil, nil
}

func (ctxOnlyChallengeContract) FindPublishedForRecommendation(context.Context, int, []string, string, []int64) ([]*model.Challenge, error) {
	return nil, nil
}

var _ challengecontracts.ImageStore = (*ctxOnlyImageStore)(nil)
var _ challengecontracts.ChallengeContract = (*ctxOnlyChallengeContract)(nil)
var _ challengecontracts.ContestChallengeContract = (*ctxOnlyContestChallengeContract)(nil)
var _ challengecontracts.PracticeChallengeContract = (*ctxOnlyPracticeChallengeContract)(nil)
