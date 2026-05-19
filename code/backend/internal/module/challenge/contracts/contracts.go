package contracts

import (
	"context"

	"ctf-platform/internal/model"
	challengeentity "ctf-platform/internal/module/challenge/entity"
)

type FlagValidator interface {
	ValidateFlag(ctx context.Context, userID, challengeID int64, input string, nonce string) (bool, error)
}

type ImageStore interface {
	FindByID(ctx context.Context, id int64) (*challengeentity.Image, error)
}

const ImageStatusAvailable = challengeentity.ImageStatusAvailable

func BuildRuntimeImageRef(image *challengeentity.Image) string {
	return challengeentity.BuildRuntimeImageRef(image)
}

type ContestChallengeContract interface {
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	BatchGetSolvedStatus(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error)
	BatchGetSolvedCount(ctx context.Context, challengeIDs []int64) (map[int64]int64, error)
}

type PracticeChallengeContract interface {
	FindByID(ctx context.Context, id int64) (*model.Challenge, error)
	FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*challengeentity.ChallengeTopology, error)
}

type ChallengeContract interface {
	ContestChallengeContract
	PracticeChallengeContract
	FindPublishedForRecommendation(ctx context.Context, limit int, dimensions []string, preferredDifficulty string, excludeSolved []int64) ([]*model.Challenge, error)
}
