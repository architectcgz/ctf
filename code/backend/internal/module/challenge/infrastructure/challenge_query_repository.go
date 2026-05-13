package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
)

type challengeQueryRawRepository interface {
	challengeports.ChallengeReadRepository
	challengeports.ChallengePublishedRepository
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

func (r *ChallengeQueryRepository) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	challenge, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengeports.ErrChallengeQueryChallengeNotFound
	}
	return challenge, err
}

func (r *ChallengeQueryRepository) List(ctx context.Context, query *dto.ChallengeQuery) ([]*model.Challenge, int64, error) {
	return r.source.List(ctx, query)
}

func (r *ChallengeQueryRepository) ListHintsByChallengeID(ctx context.Context, challengeID int64) ([]*model.ChallengeHint, error) {
	return r.source.ListHintsByChallengeID(ctx, challengeID)
}

func (r *ChallengeQueryRepository) ListPublished(ctx context.Context, query *dto.ChallengeQuery) ([]*model.Challenge, int64, error) {
	return r.source.ListPublished(ctx, query)
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

var _ challengeports.ChallengeReadRepository = (*ChallengeQueryRepository)(nil)
var _ challengeports.ChallengePublishedRepository = (*ChallengeQueryRepository)(nil)
var _ challengeports.ChallengeStatsRepository = (*ChallengeQueryRepository)(nil)
var _ challengeports.ChallengeBatchStatsRepository = (*ChallengeQueryRepository)(nil)
