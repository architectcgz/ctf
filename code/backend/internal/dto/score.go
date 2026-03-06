package dto

// UserScoreInfo 用户得分信息
type UserScoreInfo struct {
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	TotalScore  int    `json:"total_score"`
	SolvedCount int    `json:"solved_count"`
	Rank        int    `json:"rank"`
}

// RankingItem 排行榜项
type RankingItem struct {
	Rank        int    `json:"rank"`
	UserID      int64  `json:"user_id"`
	Username    string `json:"username"`
	TotalScore  int    `json:"total_score"`
	SolvedCount int    `json:"solved_count"`
}
