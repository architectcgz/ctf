package dto

// SkillDimension 能力维度
type SkillDimension struct {
	Dimension string  `json:"dimension"` // 维度名称（web/pwn/reverse/crypto/misc/forensics）
	Score     float64 `json:"score"`     // 得分率（0.0-1.0，表示该维度的完成百分比）
}

// SkillProfileResp 能力画像响应
type SkillProfileResp struct {
	UserID     int64             `json:"user_id"`    // 用户ID
	Dimensions []*SkillDimension `json:"dimensions"` // 各维度得分
	UpdatedAt  string            `json:"updated_at"` // 最后更新时间（RFC3339格式）
}
