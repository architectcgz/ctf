package dto

// SkillDimension 能力维度
type SkillDimension struct {
	Dimension string  `json:"dimension"`
	Score     float64 `json:"score"`
}

// SkillProfileResp 能力画像响应
type SkillProfileResp struct {
	UserID     int64             `json:"user_id"`
	Dimensions []*SkillDimension `json:"dimensions"`
	UpdatedAt  string            `json:"updated_at"`
}
