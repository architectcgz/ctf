package contracts

import (
	"context"

	challengeentity "ctf-platform/internal/module/challenge/entity"
)

type FlagValidator interface {
	ValidateFlag(ctx context.Context, userID, challengeID int64, input string, nonce string) (bool, error)
}

type Image = challengeentity.Image

type ImageStore interface {
	FindByID(ctx context.Context, id int64) (*Image, error)
}

type ContestChallenge struct {
	ID         int64
	Title      string
	Category   string
	Difficulty string
	Points     int
	Status     string
	FlagType   string
	FlagPrefix string
	CreatedBy  *int64
}

type RecommendationChallenge struct {
	ID                      int64  `gorm:"column:id"`
	Title                   string `gorm:"column:title"`
	Category                string `gorm:"column:category"`
	RecommendationDimension string `gorm:"column:recommendation_dimension"`
	Difficulty              string `gorm:"column:difficulty"`
	Points                  int    `gorm:"column:points"`
}

const ImageStatusAvailable = challengeentity.ImageStatusAvailable

func BuildRuntimeImageRef(image *Image) string {
	return challengeentity.BuildRuntimeImageRef(image)
}

type ContestChallengeContract interface {
	FindByID(ctx context.Context, id int64) (*ContestChallenge, error)
	BatchGetSolvedStatus(ctx context.Context, userID int64, challengeIDs []int64) (map[int64]bool, error)
	BatchGetSolvedCount(ctx context.Context, challengeIDs []int64) (map[int64]int64, error)
}

type PracticeRuntimeChallenge struct {
	ID              int64
	PackageSlug     *string
	Title           string
	Category        string
	Difficulty      string
	Points          int
	ImageID         int64
	Status          string
	FlagType        string
	FlagHash        string
	FlagSalt        string
	FlagRegex       string
	FlagPrefix      string
	InstanceSharing string
	TargetProtocol  string
	TargetPort      int
}

type PracticeRuntimeChallengeTopology struct {
	ChallengeID  int64
	EntryNodeKey string
	Spec         string
}

type PracticeChallengeContract interface {
	FindPracticeRuntimeChallengeByID(ctx context.Context, id int64) (*PracticeRuntimeChallenge, error)
	FindPracticeRuntimeChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*PracticeRuntimeChallengeTopology, error)
}

type ChallengeContract interface {
	ContestChallengeContract
	PracticeChallengeContract
	FindPublishedForRecommendation(ctx context.Context, limit int, dimensions []string, preferredDifficulty string, excludeSolved []int64) ([]*RecommendationChallenge, error)
}
