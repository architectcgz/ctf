package queries

import "time"

type AWDTrafficTrendBucketResult struct {
	BucketStart  time.Time
	RequestCount int
	ErrorCount   int
}

type AWDTrafficTopTeamResult struct {
	TeamID       int64
	TeamName     string
	RequestCount int
	ErrorCount   int
}

type AWDTrafficTopChallengeResult struct {
	AWDChallengeID    int64
	AWDChallengeTitle string
	RequestCount      int
	ErrorCount        int
}

type AWDTrafficTopPathResult struct {
	Path           string
	RequestCount   int
	ErrorCount     int
	LastStatusCode int
}

type AWDTrafficSummaryResult struct {
	Round               *AWDRoundResult
	ContestID           int64
	RoundID             int64
	TotalRequests       int
	ActiveAttackerTeams int
	TargetedTeams       int
	ErrorRequests       int
	UniquePathCount     int
	LatestEventAt       *time.Time
	Trend               []*AWDTrafficTrendBucketResult
	TopAttackers        []*AWDTrafficTopTeamResult
	TopVictims          []*AWDTrafficTopTeamResult
	TopChallenges       []*AWDTrafficTopChallengeResult
	TopPaths            []*AWDTrafficTopPathResult
	TopErrorPaths       []*AWDTrafficTopPathResult
}
