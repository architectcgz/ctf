package contracts

import (
	"context"
)

type ProfileService interface {
	UpdateSkillProfileForDimension(ctx context.Context, userID int64, dimension string) error
}

type RecommendationWeakDimension struct {
	Dimension  string
	Severity   string
	Confidence float64
	Evidence   string
}

type ChallengeRecommendation struct {
	ID             int64
	Title          string
	Category       string
	Difficulty     string
	Points         int
	Dimension      string
	DifficultyBand string
	Severity       string
	ReasonCodes    []string
	Summary        string
	Evidence       string
}

type Recommendation struct {
	WeakDimensions []RecommendationWeakDimension
	Challenges     []*ChallengeRecommendation
}

type RecommendationProvider interface {
	Recommend(ctx context.Context, userID int64, limit int) (*Recommendation, error)
}
