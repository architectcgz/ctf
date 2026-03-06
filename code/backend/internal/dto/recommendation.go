package dto

type ChallengeRecommendation struct {
	ID         int64  `json:"id"`
	Title      string `json:"title"`
	Category   string `json:"category"`
	Difficulty string `json:"difficulty"`
	Points     int    `json:"points"`
	Reason     string `json:"reason"`
}

type RecommendationResp struct {
	WeakDimensions []string                   `json:"weak_dimensions"`
	Challenges     []*ChallengeRecommendation `json:"challenges"`
}
