package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
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

func (r *RuntimeSubjectRepository) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	challenge, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeChallengeNotFound
	}
	return challenge, err
}

func (r *RuntimeSubjectRepository) FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	topology, err := r.source.FindChallengeTopologyByChallengeID(ctx, challengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeChallengeTopologyNotFound
	}
	return topology, err
}

var _ practiceports.PracticeRuntimeSubjectRepository = (*RuntimeSubjectRepository)(nil)
