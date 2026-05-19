package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
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
	return mapPracticeRuntimeChallenge(challenge), nil
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

func mapPracticeRuntimeChallenge(source *model.Challenge) *practiceentity.Challenge {
	if source == nil {
		return nil
	}
	return &practiceentity.Challenge{
		ID:              source.ID,
		PackageSlug:     source.PackageSlug,
		Title:           source.Title,
		Category:        source.Category,
		Difficulty:      source.Difficulty,
		Points:          source.Points,
		ImageID:         source.ImageID,
		Status:          string(source.Status),
		FlagType:        source.FlagType,
		FlagHash:        source.FlagHash,
		FlagSalt:        source.FlagSalt,
		FlagRegex:       source.FlagRegex,
		FlagPrefix:      source.FlagPrefix,
		InstanceSharing: string(source.InstanceSharing),
		TargetProtocol:  source.TargetProtocol,
		TargetPort:      source.TargetPort,
	}
}
