package http

import (
	"time"
)

// ProgressResp is the student-facing progress response for practice APIs.
type ProgressResp struct {
	TotalScore      int              `json:"total_score"`
	TotalSolved     int              `json:"total_solved"`
	Rank            int              `json:"rank"`
	CategoryStats   []CategoryStat   `json:"category_stats"`
	DifficultyStats []DifficultyStat `json:"difficulty_stats"`
}

type CategoryStat struct {
	Category string `json:"category"`
	Solved   int    `json:"solved"`
	Total    int    `json:"total"`
}

type DifficultyStat struct {
	Difficulty string `json:"difficulty"`
	Solved     int    `json:"solved"`
	Total      int    `json:"total"`
}

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
