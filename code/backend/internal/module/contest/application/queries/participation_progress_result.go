package queries

import "time"

type ContestSolvedProgressResult struct {
	ContestChallengeID int64
	SolvedAt           time.Time
	PointsEarned       int
}

type ParticipationProgressResult struct {
	ContestID int64
	TeamID    *int64
	Solved    []*ContestSolvedProgressResult
}
