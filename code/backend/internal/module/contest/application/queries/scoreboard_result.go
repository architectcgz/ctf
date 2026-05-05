package queries

import "time"

type ScoreboardContestResult struct {
	ID        int64
	Title     string
	Status    string
	StartedAt time.Time
	EndsAt    time.Time
}

type ScoreboardItemResult struct {
	Rank             int
	TeamID           int64
	TeamName         string
	Score            float64
	SolvedCount      int
	LastSubmissionAt *time.Time
}

type ScoreboardPageResult struct {
	List     []*ScoreboardItemResult
	Total    int64
	Page     int
	PageSize int
}

type ScoreboardResult struct {
	Contest    *ScoreboardContestResult
	Scoreboard *ScoreboardPageResult
	Frozen     bool
}
