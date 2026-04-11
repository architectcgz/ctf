package dto

import (
	"time"

	"ctf-platform/internal/model"
)

type ChallengeHintReq struct {
	Level   int    `json:"level" binding:"required,min=1"`
	Title   string `json:"title" binding:"omitempty,max=128"`
	Content string `json:"content" binding:"required"`
}

type CreateChallengeReq struct {
	Title           string                `json:"title" binding:"required"`
	Description     string                `json:"description" binding:"required"`
	Category        string                `json:"category" binding:"required"`
	Difficulty      string                `json:"difficulty" binding:"required,oneof=beginner easy medium hard insane"`
	Points          int                   `json:"points" binding:"required,min=1"`
	ImageID         int64                 `json:"image_id"`
	AttachmentURL   string                `json:"attachment_url" binding:"omitempty,max=2048"`
	InstanceSharing model.InstanceSharing `json:"instance_sharing" binding:"omitempty,oneof=per_user per_team shared"`
	Hints           []ChallengeHintReq    `json:"hints" binding:"omitempty,dive"`
}

type UpdateChallengeReq struct {
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	Category        string                `json:"category"`
	Difficulty      string                `json:"difficulty" binding:"omitempty,oneof=beginner easy medium hard insane"`
	Points          int                   `json:"points" binding:"omitempty,min=1"`
	ImageID         *int64                `json:"image_id"`
	AttachmentURL   *string               `json:"attachment_url" binding:"omitempty,max=2048"`
	InstanceSharing model.InstanceSharing `json:"instance_sharing" binding:"omitempty,oneof=per_user per_team shared"`
	Hints           []ChallengeHintReq    `json:"hints" binding:"omitempty,dive"`
}

type ChallengeHintAdminResp struct {
	ID      int64  `json:"id"`
	Level   int    `json:"level"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content"`
}

type ChallengeResp struct {
	ID              int64                     `json:"id"`
	Title           string                    `json:"title"`
	Description     string                    `json:"description"`
	Category        string                    `json:"category"`
	Difficulty      string                    `json:"difficulty"`
	Points          int                       `json:"points"`
	ImageID         int64                     `json:"image_id"`
	AttachmentURL   string                    `json:"attachment_url,omitempty"`
	InstanceSharing model.InstanceSharing     `json:"instance_sharing"`
	Hints           []*ChallengeHintAdminResp `json:"hints,omitempty"`
	Status          string                    `json:"status"`
	CreatedBy       *int64                    `json:"created_by,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

type ChallengeQuery struct {
	Category   string `form:"category"`
	Difficulty string `form:"difficulty"`
	Status     string `form:"status"`
	CreatedBy  *int64 `form:"created_by"`
	Keyword    string `form:"keyword"`
	SortBy     string `form:"sort_by" binding:"omitempty,oneof=created_at difficulty"`
	Page       int    `form:"page" binding:"omitempty,min=1"`
	Size       int    `form:"page_size" binding:"omitempty,min=1,max=100"`
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
	ID              int64                 `json:"id"`
	Title           string                `json:"title"`
	Description     string                `json:"description"`
	Category        string                `json:"category"`
	Difficulty      string                `json:"difficulty"`
	Points          int                   `json:"points"`
	NeedTarget      bool                  `json:"need_target"`
	FlagType        string                `json:"flag_type"`
	InstanceSharing model.InstanceSharing `json:"instance_sharing"`
	AttachmentURL   string                `json:"attachment_url,omitempty"`
	Hints           []*ChallengeHintResp  `json:"hints"`
	SolvedCount     int64                 `json:"solved_count"`
	TotalAttempts   int64                 `json:"total_attempts"`
	IsSolved        bool                  `json:"is_solved"`
	CreatedAt       time.Time             `json:"created_at"`
}

type ChallengeHintResp struct {
	ID      int64  `json:"id"`
	Level   int    `json:"level"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
}

type ConfigureFlagReq struct {
	FlagType   string `json:"flag_type" binding:"required,oneof=static dynamic regex manual_review"`
	Flag       string `json:"flag" binding:"required_if=FlagType static"`
	FlagRegex  string `json:"flag_regex" binding:"required_if=FlagType regex"`
	FlagPrefix string `json:"flag_prefix" binding:"omitempty,max=32"`
}

type FlagResp struct {
	FlagType   string `json:"flag_type"`
	FlagRegex  string `json:"flag_regex,omitempty"`
	FlagPrefix string `json:"flag_prefix,omitempty"`
	Configured bool   `json:"configured"`
}

type ChallengeSelfCheckStepResp struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

type ChallengeSelfCheckPhaseResp struct {
	Passed    bool                         `json:"passed"`
	StartedAt time.Time                    `json:"started_at"`
	EndedAt   time.Time                    `json:"ended_at"`
	Steps     []ChallengeSelfCheckStepResp `json:"steps"`
}

type ChallengeSelfCheckRuntimeResp struct {
	Passed         bool                         `json:"passed"`
	StartedAt      time.Time                    `json:"started_at"`
	EndedAt        time.Time                    `json:"ended_at"`
	AccessURL      string                       `json:"access_url,omitempty"`
	ContainerCount int                          `json:"container_count"`
	NetworkCount   int                          `json:"network_count"`
	Steps          []ChallengeSelfCheckStepResp `json:"steps"`
}

type ChallengeSelfCheckResp struct {
	ChallengeID int64                         `json:"challenge_id"`
	Precheck    ChallengeSelfCheckPhaseResp   `json:"precheck"`
	Runtime     ChallengeSelfCheckRuntimeResp `json:"runtime"`
}

type ChallengePublishCheckJobResp struct {
	ID             int64                   `json:"id"`
	ChallengeID    int64                   `json:"challenge_id"`
	RequestedBy    int64                   `json:"requested_by"`
	Status         string                  `json:"status"`
	Active         bool                    `json:"active"`
	RequestSource  string                  `json:"request_source"`
	FailureSummary string                  `json:"failure_summary,omitempty"`
	StartedAt      *time.Time              `json:"started_at,omitempty"`
	FinishedAt     *time.Time              `json:"finished_at,omitempty"`
	PublishedAt    *time.Time              `json:"published_at,omitempty"`
	CreatedAt      time.Time               `json:"created_at"`
	UpdatedAt      time.Time               `json:"updated_at"`
	Result         *ChallengeSelfCheckResp `json:"result,omitempty"`
}
