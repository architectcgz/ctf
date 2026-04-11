package dto

import (
	"time"

	"ctf-platform/internal/model"
)

type AddContestChallengeReq struct {
	ChallengeID      int64                `json:"challenge_id" binding:"required"`
	Points           int                  `json:"points" binding:"omitempty,min=1"`
	Order            int                  `json:"order" binding:"omitempty,min=0"`
	IsVisible        *bool                `json:"is_visible"`
	AWDCheckerType   model.AWDCheckerType `json:"awd_checker_type" binding:"omitempty,oneof=legacy_probe http_standard"`
	AWDCheckerConfig map[string]any       `json:"awd_checker_config"`
	AWDSLAScore      int                  `json:"awd_sla_score" binding:"omitempty,min=0"`
	AWDDefenseScore  int                  `json:"awd_defense_score" binding:"omitempty,min=0"`
}

type UpdateContestChallengeReq struct {
	Points           *int           `json:"points" binding:"omitempty,min=1"`
	Order            *int           `json:"order" binding:"omitempty,min=0"`
	IsVisible        *bool          `json:"is_visible"`
	AWDCheckerType   *string        `json:"awd_checker_type" binding:"omitempty,oneof=legacy_probe http_standard"`
	AWDCheckerConfig map[string]any `json:"awd_checker_config"`
	AWDSLAScore      *int           `json:"awd_sla_score" binding:"omitempty,min=0"`
	AWDDefenseScore  *int           `json:"awd_defense_score" binding:"omitempty,min=0"`
}

type ContestChallengeResp struct {
	ID               int64                `json:"id"`
	ContestID        int64                `json:"contest_id"`
	ChallengeID      int64                `json:"challenge_id"`
	Title            string               `json:"title,omitempty"`
	Category         string               `json:"category,omitempty"`
	Difficulty       string               `json:"difficulty,omitempty"`
	Points           int                  `json:"points"`
	Order            int                  `json:"order"`
	IsVisible        bool                 `json:"is_visible"`
	AWDCheckerType   model.AWDCheckerType `json:"awd_checker_type,omitempty"`
	AWDCheckerConfig map[string]any       `json:"awd_checker_config,omitempty"`
	AWDSLAScore      int                  `json:"awd_sla_score"`
	AWDDefenseScore  int                  `json:"awd_defense_score"`
	CreatedAt        time.Time            `json:"created_at"`
}

type ContestChallengeInfo struct {
	ID          int64  `json:"id"`
	ChallengeID int64  `json:"challenge_id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Difficulty  string `json:"difficulty"`
	Points      int    `json:"points"`
	Order       int    `json:"order"`
	SolvedCount int64  `json:"solved_count"`
	IsSolved    bool   `json:"is_solved"`
}
