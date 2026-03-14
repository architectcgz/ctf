package dto

import "time"

type AddContestChallengeReq struct {
	ChallengeID int64 `json:"challenge_id" binding:"required"`
	Points      int   `json:"points" binding:"omitempty,min=1"`
	Order       int   `json:"order" binding:"omitempty,min=0"`
	IsVisible   *bool `json:"is_visible"`
}

type UpdateContestChallengeReq struct {
	Points    *int  `json:"points" binding:"omitempty,min=1"`
	Order     *int  `json:"order" binding:"omitempty,min=0"`
	IsVisible *bool `json:"is_visible"`
}

type ContestChallengeResp struct {
	ID          int64     `json:"id"`
	ContestID   int64     `json:"contest_id"`
	ChallengeID int64     `json:"challenge_id"`
	Title       string    `json:"title,omitempty"`
	Category    string    `json:"category,omitempty"`
	Difficulty  string    `json:"difficulty,omitempty"`
	Points      int       `json:"points"`
	Order       int       `json:"order"`
	IsVisible   bool      `json:"is_visible"`
	CreatedAt   time.Time `json:"created_at"`
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
