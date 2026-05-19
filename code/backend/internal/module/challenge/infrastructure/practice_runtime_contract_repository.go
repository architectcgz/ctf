package infrastructure

import (
	"context"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
)

func (r *Repository) FindPracticeRuntimeChallengeByID(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
	challenge, err := r.FindByID(ctx, id)
	if err != nil || challenge == nil {
		return nil, err
	}
	return &challengecontracts.PracticeRuntimeChallenge{
		ID:              challenge.ID,
		PackageSlug:     challenge.PackageSlug,
		Title:           challenge.Title,
		Category:        challenge.Category,
		Difficulty:      challenge.Difficulty,
		Points:          challenge.Points,
		ImageID:         challenge.ImageID,
		Status:          string(challenge.Status),
		FlagType:        challenge.FlagType,
		FlagHash:        challenge.FlagHash,
		FlagSalt:        challenge.FlagSalt,
		FlagRegex:       challenge.FlagRegex,
		FlagPrefix:      challenge.FlagPrefix,
		InstanceSharing: string(challenge.InstanceSharing),
		TargetProtocol:  challenge.TargetProtocol,
		TargetPort:      challenge.TargetPort,
	}, nil
}

func (r *Repository) FindPracticeRuntimeChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*challengecontracts.PracticeRuntimeChallengeTopology, error) {
	topology, err := r.FindChallengeTopologyByChallengeID(ctx, challengeID)
	if err != nil || topology == nil {
		return nil, err
	}
	return &challengecontracts.PracticeRuntimeChallengeTopology{
		ChallengeID:  topology.ChallengeID,
		EntryNodeKey: topology.EntryNodeKey,
		Spec:         topology.Spec,
	}, nil
}
