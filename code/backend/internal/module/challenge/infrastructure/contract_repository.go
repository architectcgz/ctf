package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeentity "ctf-platform/internal/module/challenge/entity"
)

type ContractRepository struct {
	source *Repository
}

func NewContractRepository(source *Repository) *ContractRepository {
	if source == nil {
		return nil
	}
	return &ContractRepository{source: source}
}

func (r *ContractRepository) FindByID(ctx context.Context, id int64) (*challengecontracts.ContestChallenge, error) {
	challenge, err := r.source.FindByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, challengecontracts.ErrChallengeNotFound
	}
	if err != nil || challenge == nil {
		return nil, err
	}
	return toContestChallengeProjection(challenge), nil
}

func (r *ContractRepository) BatchGetSolvedStatus(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error) {
	return r.source.BatchGetSolvedStatus(ctx, userID, challengeIDs)
}

func (r *ContractRepository) BatchGetSolvedCount(ctx context.Context, challengeIDs []int64) (map[int64]int64, error) {
	return r.source.BatchGetSolvedCount(ctx, challengeIDs)
}

func (r *ContractRepository) FindPracticeRuntimeChallengeByID(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
	return r.source.FindPracticeRuntimeChallengeByID(ctx, id)
}

func (r *ContractRepository) FindPracticeRuntimeChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*challengecontracts.PracticeRuntimeChallengeTopology, error) {
	return r.source.FindPracticeRuntimeChallengeTopologyByChallengeID(ctx, challengeID)
}

func (r *ContractRepository) FindPublishedForRecommendation(ctx context.Context, limit int, dimensions []string, preferredDifficulty string, excludeSolved []int64) ([]*challengecontracts.RecommendationChallenge, error) {
	challenges, err := r.source.FindPublishedForRecommendation(ctx, limit, dimensions, preferredDifficulty, excludeSolved)
	if err != nil || len(challenges) == 0 {
		return []*challengecontracts.RecommendationChallenge{}, err
	}

	items := make([]*challengecontracts.RecommendationChallenge, 0, len(challenges))
	for _, challenge := range challenges {
		if challenge == nil {
			continue
		}
		items = append(items, toRecommendationChallengeProjection(challenge))
	}
	return items, nil
}

func toContestChallengeProjection(challenge *challengeentity.Challenge) *challengecontracts.ContestChallenge {
	return &challengecontracts.ContestChallenge{
		ID:         challenge.ID,
		Title:      challenge.Title,
		Category:   challenge.Category,
		Difficulty: challenge.Difficulty,
		Points:     challenge.Points,
		Status:     string(challenge.Status),
		FlagType:   challenge.FlagType,
		FlagPrefix: challenge.FlagPrefix,
		CreatedBy:  challenge.CreatedBy,
	}
}

func toRecommendationChallengeProjection(challenge *challengeentity.Challenge) *challengecontracts.RecommendationChallenge {
	return &challengecontracts.RecommendationChallenge{
		ID:                      challenge.ID,
		Title:                   challenge.Title,
		Category:                challenge.Category,
		RecommendationDimension: challenge.RecommendationDimension,
		Difficulty:              challenge.Difficulty,
		Points:                  challenge.Points,
	}
}

var _ challengecontracts.ChallengeContract = (*ContractRepository)(nil)
