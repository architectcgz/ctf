package http

type SkillDimension struct {
	Dimension string  `json:"dimension"`
	Score     float64 `json:"score"`
}

type SkillProfileResp struct {
	UserID     int64             `json:"user_id"`
	Dimensions []*SkillDimension `json:"dimensions"`
	UpdatedAt  string            `json:"updated_at"`
}

type RecommendationWeakDimension struct {
	Dimension  string  `json:"dimension"`
	Severity   string  `json:"severity"`
	Confidence float64 `json:"confidence"`
	Evidence   string  `json:"evidence,omitempty"`
}

type ChallengeRecommendation struct {
	ID             int64    `json:"id"`
	Title          string   `json:"title"`
	Category       string   `json:"category"`
	Difficulty     string   `json:"difficulty"`
	Points         int      `json:"points"`
	Dimension      string   `json:"dimension,omitempty"`
	DifficultyBand string   `json:"difficulty_band,omitempty"`
	Severity       string   `json:"severity,omitempty"`
	ReasonCodes    []string `json:"reason_codes,omitempty"`
	Summary        string   `json:"summary"`
	Evidence       string   `json:"evidence,omitempty"`
}

type RecommendationResp struct {
	WeakDimensions []RecommendationWeakDimension `json:"weak_dimensions"`
	Challenges     []*ChallengeRecommendation    `json:"challenges"`
}
