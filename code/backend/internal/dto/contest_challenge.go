package dto

import "time"

type AddContestChallengeReq struct {
	ChallengeID int64 `json:"challenge_id" binding:"required"`
	Points      int   `json:"points" binding:"omitempty,min=1"`
	Order       int   `json:"order" binding:"omitempty,min=0"`
}

type UpdateContestChallengeReq struct {
	Points int `json:"points" binding:"required,min=1"`
}

type ContestChallengeResp struct {
	ID          int64     `json:"id"`
	ContestID   int64     `json:"contest_id"`
	ChallengeID int64     `json:"challenge_id"`
	Points      int       `json:"points"`
	Order       int       `json:"order"`
	CreatedAt   time.Time `json:"created_at"`
}
