package ports_test

import (
	"context"

	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type ctxOnlyChallengeTopologyRepository struct{}

func (ctxOnlyChallengeTopologyRepository) FindByID(context.Context, int64) (*model.Challenge, error) {
	return nil, nil
}

func (ctxOnlyChallengeTopologyRepository) FindChallengeTopologyByChallengeID(context.Context, int64) (*model.ChallengeTopology, error) {
	return nil, nil
}

func (ctxOnlyChallengeTopologyRepository) UpsertChallengeTopology(context.Context, *model.ChallengeTopology) error {
	return nil
}

func (ctxOnlyChallengeTopologyRepository) DeleteChallengeTopologyByChallengeID(context.Context, int64) error {
	return nil
}

var _ challengeports.ChallengeTopologyRepository = (*ctxOnlyChallengeTopologyRepository)(nil)
