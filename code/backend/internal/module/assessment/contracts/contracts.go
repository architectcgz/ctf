package contracts

import (
	"context"
)

type ProfileService interface {
	UpdateSkillProfileForDimension(ctx context.Context, userID int64, dimension string) error
}

type SkillDimension struct {
	Dimension string  `json:"dimension"`
	Score     float64 `json:"score"`
}

type SkillProfile struct {
	UserID     int64             `json:"user_id"`
	Dimensions []*SkillDimension `json:"dimensions"`
	UpdatedAt  string            `json:"updated_at"`
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
