package queries

import "time"

type ContestChallengeResult struct {
	ID          int64
	ContestID   int64
	ChallengeID int64
	Title       string
	Category    string
	Difficulty  string
	Points      int
	Order       int
	IsVisible   bool
	CreatedAt   time.Time
}
