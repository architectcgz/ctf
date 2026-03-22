package contracts

import (
	"context"

	"ctf-platform/internal/model"
)

type FlagValidator interface {
	ValidateFlag(userID, challengeID int64, input string, nonce string) (bool, error)
}

type ContestChallengeContract interface {
	FindByID(id int64) (*model.Challenge, error)
	BatchGetSolvedStatus(userID int64, challengeIDs []int64) (map[int64]bool, error)
	BatchGetSolvedCount(challengeIDs []int64) (map[int64]int64, error)
}

type PracticeChallengeContract interface {
	FindByID(id int64) (*model.Challenge, error)
	FindHintByLevel(challengeID int64, level int) (*model.ChallengeHint, error)
	CreateHintUnlock(unlock *model.ChallengeHintUnlock) error
	FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error)
}

type ChallengeContract interface {
	ContestChallengeContract
	PracticeChallengeContract
	FindPublishedForRecommendation(limit int, dimensions []string, excludeSolved []int64) ([]*model.Challenge, error)
	FindPublishedForRecommendationWithContext(ctx context.Context, limit int, dimensions []string, excludeSolved []int64) ([]*model.Challenge, error)
}
