package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	practiceentity "ctf-platform/internal/module/practice/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type RuntimeSubjectRepository struct {
	source challengecontracts.PracticeChallengeContract
}

func NewRuntimeSubjectRepository(source challengecontracts.PracticeChallengeContract) *RuntimeSubjectRepository {
	if source == nil {
		return nil
	}
	return &RuntimeSubjectRepository{source: source}
}

func (r *RuntimeSubjectRepository) FindByID(ctx context.Context, id int64) (*practiceentity.Challenge, error) {
	challenge, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeChallengeNotFound
	}
	if err != nil || challenge == nil {
		return nil, err
	}
	return &practiceentity.Challenge{
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

func (r *RuntimeSubjectRepository) FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*practiceports.RuntimeChallengeTopology, error) {
	topology, err := r.source.FindChallengeTopologyByChallengeID(ctx, challengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeChallengeTopologyNotFound
	}
	if err != nil || topology == nil {
		return nil, err
	}
	return &practiceports.RuntimeChallengeTopology{
		ChallengeID:  topology.ChallengeID,
		EntryNodeKey: topology.EntryNodeKey,
		Spec:         topology.Spec,
	}, nil
}

var _ practiceports.PracticeRuntimeSubjectRepository = (*RuntimeSubjectRepository)(nil)
