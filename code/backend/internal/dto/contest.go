package dto

import "time"

type ScoreboardItem struct {
	TeamID   int64   `json:"team_id"`
	TeamName string  `json:"team_name"`
	Score    float64 `json:"score"`
	Rank     int     `json:"rank"`
}

type ScoreboardResp struct {
	ContestID  int64             `json:"contest_id"`
	IsFrozen   bool              `json:"is_frozen"`
	FreezeTime *time.Time        `json:"freeze_time,omitempty"`
	Items      []*ScoreboardItem `json:"items"`
}

type TeamRankResp struct {
	TeamID int64   `json:"team_id"`
	Rank   int     `json:"rank"`
	Score  float64 `json:"score"`
}

type FreezeReq struct {
	MinutesBeforeEnd int `json:"minutes_before_end" binding:"required,min=1"`
}
