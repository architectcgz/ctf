package dto

import "time"

type CreateContestAWDServiceReq struct {
	ChallengeID int64  `json:"challenge_id" binding:"required,min=1"`
	TemplateID  int64  `json:"template_id" binding:"required,min=1"`
	DisplayName string `json:"display_name" binding:"omitempty,max=128"`
	Order       int    `json:"order" binding:"omitempty,min=0"`
	IsVisible   *bool  `json:"is_visible"`
}

type UpdateContestAWDServiceReq struct {
	TemplateID  *int64  `json:"template_id" binding:"omitempty,min=1"`
	DisplayName *string `json:"display_name" binding:"omitempty,max=128"`
	Order       *int    `json:"order" binding:"omitempty,min=0"`
	IsVisible   *bool   `json:"is_visible"`
}

type ContestAWDServiceResp struct {
	ID            int64          `json:"id"`
	ContestID     int64          `json:"contest_id"`
	ChallengeID   int64          `json:"challenge_id"`
	TemplateID    *int64         `json:"template_id,omitempty"`
	DisplayName   string         `json:"display_name"`
	Order         int            `json:"order"`
	IsVisible     bool           `json:"is_visible"`
	ScoreConfig   map[string]any `json:"score_config,omitempty"`
	RuntimeConfig map[string]any `json:"runtime_config,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}
