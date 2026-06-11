package queries

import "time"

type TimelineEvent struct {
	Type        string    `json:"type"`
	ChallengeID int64     `json:"challenge_id"`
	Title       string    `json:"title"`
	Timestamp   time.Time `json:"timestamp"`
	IsCorrect   *bool     `json:"is_correct,omitempty"`
	Points      *int      `json:"points,omitempty"`
	Detail      string    `json:"detail,omitempty"`
}

type TimelineResp struct {
	Events []TimelineEvent `json:"events"`
}
