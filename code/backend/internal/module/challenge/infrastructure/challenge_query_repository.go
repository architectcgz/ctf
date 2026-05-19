package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type challengeQueryRawRepository interface {
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	List(ctx context.Context, query *challengecontracts.ChallengeQuery) ([]*model.Challenge, int64, error)
	ListPublished(ctx context.Context, query *challengecontracts.ChallengeQuery) ([]*model.Challenge, int64, error)
	ListHintsByChallengeID(ctx context.Context, challengeID int64) ([]*challengeentity.ChallengeHint, error)
	challengeports.ChallengeStatsRepository
	challengeports.ChallengeBatchStatsRepository
}

type ChallengeQueryRepository struct {
	source challengeQueryRawRepository
}

func NewChallengeQueryRepository(source challengeQueryRawRepository) *ChallengeQueryRepository {
	if source == nil {
		return nil
	}
	return &ChallengeQueryRepository{source: source}
}

func (r *ChallengeQueryRepository) FindByID(ctx context.Context, id int64) (*challengeports.ChallengeReadModel, error) {
	challenge, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeQueryChallengeNotFound
	}
	if err != nil || challenge == nil {
		return nil, err
	}
	return challengeReadModelFromModel(challenge), nil
}

func (r *ChallengeQueryRepository) List(ctx context.Context, query *challengecontracts.ChallengeQuery) ([]*challengeports.ChallengeReadModel, int64, error) {
	challenges, total, err := r.source.List(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	result := make([]*challengeports.ChallengeReadModel, 0, len(challenges))
	for _, challenge := range challenges {
		result = append(result, challengeReadModelFromModel(challenge))
	}
	return result, total, nil
}

func (r *ChallengeQueryRepository) ListHintsByChallengeID(ctx context.Context, challengeID int64) ([]*challengeentity.ChallengeHint, error) {
	return r.source.ListHintsByChallengeID(ctx, challengeID)
}

func (r *ChallengeQueryRepository) ListPublished(ctx context.Context, query *challengecontracts.ChallengeQuery) ([]*challengeports.ChallengeReadModel, int64, error) {
	challenges, total, err := r.source.ListPublished(ctx, query)
	if err != nil {
		return nil, 0, err
	}
	result := make([]*challengeports.ChallengeReadModel, 0, len(challenges))
	for _, challenge := range challenges {
		result = append(result, challengeReadModelFromModel(challenge))
	}
	return result, total, nil
}

func (r *ChallengeQueryRepository) GetSolvedStatus(ctx context.Context, userID, challengeID int64) (bool, error) {
	return r.source.GetSolvedStatus(ctx, userID, challengeID)
}

func (r *ChallengeQueryRepository) GetSolvedCount(ctx context.Context, challengeID int64) (int64, error) {
	return r.source.GetSolvedCount(ctx, challengeID)
}

func (r *ChallengeQueryRepository) GetTotalAttempts(ctx context.Context, challengeID int64) (int64, error) {
	return r.source.GetTotalAttempts(ctx, challengeID)
}

func (r *ChallengeQueryRepository) BatchGetSolvedStatus(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error) {
	return r.source.BatchGetSolvedStatus(ctx, userID, challengeIDs)
}

func (r *ChallengeQueryRepository) BatchGetSolvedCount(ctx context.Context, challengeIDs []int64) (map[int64]int64, error) {
	return r.source.BatchGetSolvedCount(ctx, challengeIDs)
}

func (r *ChallengeQueryRepository) BatchGetTotalAttempts(ctx context.Context, challengeIDs []int64) (map[int64]int64, error) {
	return r.source.BatchGetTotalAttempts(ctx, challengeIDs)
}

func challengeReadModelFromModel(source *model.Challenge) *challengeports.ChallengeReadModel {
	if source == nil {
		return nil
	}
	return &challengeports.ChallengeReadModel{
		ID:              source.ID,
		PackageSlug:     source.PackageSlug,
		Title:           source.Title,
		Description:     source.Description,
		Category:        source.Category,
		Difficulty:      source.Difficulty,
		Points:          source.Points,
		ImageID:         source.ImageID,
		AttachmentURL:   source.AttachmentURL,
		Status:          string(source.Status),
		FlagType:        source.FlagType,
		FlagPrefix:      source.FlagPrefix,
		InstanceSharing: string(source.InstanceSharing),
		CreatedBy:       source.CreatedBy,
		CreatedAt:       source.CreatedAt,
		UpdatedAt:       source.UpdatedAt,
	}
}

var _ challengeports.ChallengeReadRepository = (*ChallengeQueryRepository)(nil)
var _ challengeports.ChallengePublishedRepository = (*ChallengeQueryRepository)(nil)
var _ challengeports.ChallengeStatsRepository = (*ChallengeQueryRepository)(nil)
var _ challengeports.ChallengeBatchStatsRepository = (*ChallengeQueryRepository)(nil)
