package dto

import (
	"time"

	"ctf-platform/internal/model"
)

type CreateContestAWDServiceReq struct {
	TemplateID             int64          `json:"template_id" binding:"required,min=1"`
	Points                 int            `json:"points" binding:"required,min=1"`
	DisplayName            string         `json:"display_name" binding:"omitempty,max=128"`
	Order                  int            `json:"order" binding:"omitempty,min=0"`
	IsVisible              *bool          `json:"is_visible"`
	CheckerType            *string        `json:"checker_type" binding:"omitempty,oneof=legacy_probe http_standard"`
	CheckerConfig          map[string]any `json:"checker_config"`
	AWDSLAScore            *int           `json:"awd_sla_score" binding:"omitempty,min=0"`
	AWDDefenseScore        *int           `json:"awd_defense_score" binding:"omitempty,min=0"`
	AWDCheckerPreviewToken *string        `json:"awd_checker_preview_token" binding:"omitempty,max=200"`
}

type UpdateContestAWDServiceReq struct {
	TemplateID             *int64         `json:"template_id" binding:"omitempty,min=1"`
	Points                 *int           `json:"points" binding:"omitempty,min=1"`
	DisplayName            *string        `json:"display_name" binding:"omitempty,max=128"`
	Order                  *int           `json:"order" binding:"omitempty,min=0"`
	IsVisible              *bool          `json:"is_visible"`
	CheckerType            *string        `json:"checker_type" binding:"omitempty,oneof=legacy_probe http_standard"`
	CheckerConfig          map[string]any `json:"checker_config"`
	AWDSLAScore            *int           `json:"awd_sla_score" binding:"omitempty,min=0"`
	AWDDefenseScore        *int           `json:"awd_defense_score" binding:"omitempty,min=0"`
	AWDCheckerPreviewToken *string        `json:"awd_checker_preview_token" binding:"omitempty,max=200"`
}

type ContestAWDServiceResp struct {
	ID                int64                           `json:"id"`
	ContestID         int64                           `json:"contest_id"`
	ChallengeID       int64                           `json:"challenge_id"`
	TemplateID        *int64                          `json:"template_id,omitempty"`
	Title             string                          `json:"title,omitempty"`
	Category          string                          `json:"category,omitempty"`
	Difficulty        string                          `json:"difficulty,omitempty"`
	DisplayName       string                          `json:"display_name"`
	Order             int                             `json:"order"`
	IsVisible         bool                            `json:"is_visible"`
	ScoreConfig       map[string]any                  `json:"score_config,omitempty"`
	RuntimeConfig     map[string]any                  `json:"runtime_config,omitempty"`
	ValidationState   model.AWDCheckerValidationState `json:"validation_state"`
	LastPreviewAt     *time.Time                      `json:"last_preview_at,omitempty"`
	LastPreviewResult *AWDCheckerPreviewResp          `json:"last_preview_result,omitempty"`
	CreatedAt         time.Time                       `json:"created_at"`
	UpdatedAt         time.Time                       `json:"updated_at"`
}
