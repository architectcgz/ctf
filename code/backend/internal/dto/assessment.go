package dto

type SkillDimension struct {
	Dimension string  `json:"dimension"`
	Score     float64 `json:"score"`
	IsWeak    bool    `json:"is_weak"`
}

type SkillProfileResp struct {
	UserID     int64            `json:"user_id"`
	Dimensions []SkillDimension `json:"dimensions"`
	UpdatedAt  string           `json:"updated_at"`
}

type ChallengeRecommendation struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Category   string `json:"category"`
	Difficulty string `json:"difficulty"`
	Points     int    `json:"points"`
	Reason     string `json:"reason"`
}

type RecommendationResp struct {
	WeakDimensions []string                  `json:"weak_dimensions"`
	Challenges     []ChallengeRecommendation `json:"challenges"`
}
