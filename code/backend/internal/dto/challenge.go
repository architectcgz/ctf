package dto

import "time"

type ChallengeHintReq struct {
	Level      int    `json:"level" binding:"required,min=1"`
	Title      string `json:"title" binding:"omitempty,max=128"`
	CostPoints int    `json:"cost_points" binding:"omitempty,min=0"`
	Content    string `json:"content" binding:"required"`
}

type CreateChallengeReq struct {
	Title         string             `json:"title" binding:"required"`
	Description   string             `json:"description" binding:"required"`
	Category      string             `json:"category" binding:"required"`
	Difficulty    string             `json:"difficulty" binding:"required,oneof=beginner easy medium hard insane"`
	Points        int                `json:"points" binding:"required,min=1"`
	ImageID       int64              `json:"image_id"`
	AttachmentURL string             `json:"attachment_url" binding:"omitempty,max=2048"`
	Hints         []ChallengeHintReq `json:"hints" binding:"omitempty,dive"`
}

type UpdateChallengeReq struct {
	Title         string             `json:"title"`
	Description   string             `json:"description"`
	Category      string             `json:"category"`
	Difficulty    string             `json:"difficulty" binding:"omitempty,oneof=beginner easy medium hard insane"`
	Points        int                `json:"points" binding:"omitempty,min=1"`
	ImageID       *int64             `json:"image_id"`
	AttachmentURL *string            `json:"attachment_url" binding:"omitempty,max=2048"`
	Hints         []ChallengeHintReq `json:"hints" binding:"omitempty,dive"`
}

type ChallengeHintAdminResp struct {
	ID         int64  `json:"id"`
	Level      int    `json:"level"`
	Title      string `json:"title,omitempty"`
	CostPoints int    `json:"cost_points,omitempty"`
	Content    string `json:"content"`
}

type ChallengeResp struct {
	ID            int64                     `json:"id"`
	Title         string                    `json:"title"`
	Description   string                    `json:"description"`
	Category      string                    `json:"category"`
	Difficulty    string                    `json:"difficulty"`
	Points        int                       `json:"points"`
	ImageID       int64                     `json:"image_id"`
	AttachmentURL string                    `json:"attachment_url,omitempty"`
	Hints         []*ChallengeHintAdminResp `json:"hints,omitempty"`
	Status        string                    `json:"status"`
	CreatedAt     time.Time                 `json:"created_at"`
	UpdatedAt     time.Time                 `json:"updated_at"`
}

type ChallengeQuery struct {
	Category   string `form:"category"`
	Difficulty string `form:"difficulty"`
	Status     string `form:"status"`
	Keyword    string `form:"keyword"`
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=created_at difficulty"`
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Size       int    `form:"size" binding:"omitempty,min=1,max=100"`
}

// ChallengeListItem 学员视图靶场列表项
type ChallengeListItem struct {
	ID            int64     `json:"id"`
	Title         string    `json:"title"`
	Category      string    `json:"category"`
	Difficulty    string    `json:"difficulty"`
	Points        int       `json:"points"`
	SolvedCount   int64     `json:"solved_count"`
	TotalAttempts int64     `json:"total_attempts"`
	IsSolved      bool      `json:"is_solved"`
	CreatedAt     time.Time `json:"created_at"`
}

// ChallengeDetailResp 学员视图靶场详情
type ChallengeDetailResp struct {
	ID            int64                `json:"id"`
	Title         string               `json:"title"`
	Description   string               `json:"description"`
	Category      string               `json:"category"`
	Difficulty    string               `json:"difficulty"`
	Points        int                  `json:"points"`
	NeedTarget    bool                 `json:"need_target"`
	AttachmentURL string               `json:"attachment_url,omitempty"`
	Hints         []*ChallengeHintResp `json:"hints"`
	SolvedCount   int64                `json:"solved_count"`
	TotalAttempts int64                `json:"total_attempts"`
	IsSolved      bool                 `json:"is_solved"`
	CreatedAt     time.Time            `json:"created_at"`
}

type ChallengeHintResp struct {
	ID         int64  `json:"id"`
	Level      int    `json:"level"`
	Title      string `json:"title,omitempty"`
	CostPoints int    `json:"cost_points,omitempty"`
	IsUnlocked bool   `json:"is_unlocked"`
	Content    string `json:"content,omitempty"`
}

type UnlockHintResp struct {
	Hint *ChallengeHintResp `json:"hint"`
}

type ConfigureFlagReq struct {
	FlagType   string `json:"flag_type" binding:"required,oneof=static dynamic"`
	Flag       string `json:"flag" binding:"required_if=FlagType static"`
	FlagPrefix string `json:"flag_prefix" binding:"omitempty,max=32"`
}

type FlagResp struct {
	FlagType   string `json:"flag_type"`
	FlagPrefix string `json:"flag_prefix,omitempty"`
	Configured bool   `json:"configured"`
}
