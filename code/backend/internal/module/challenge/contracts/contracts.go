package contracts

import (
	"context"

	"ctf-platform/internal/model"
)

type FlagValidator interface {
	ValidateFlag(ctx context.Context, userID, challengeID int64, input string, nonce string) (bool, error)
}

type ImageStore interface {
	FindByID(ctx context.Context, id int64) (*model.Image, error)
}

type ContestChallengeContract interface {
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	BatchGetSolvedStatus(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error)
	BatchGetSolvedCount(ctx context.Context, challengeIDs []int64) (map[int64]int64, error)
}

type PracticeChallengeContract interface {
	FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error)
	FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
}

type ChallengeContract interface {
	ContestChallengeContract
	PracticeChallengeContract
	FindPublishedForRecommendationWithContext(ctx context.Context, limit int, dimensions []string, excludeSolved []int64) ([]*model.Challenge, error)
}
