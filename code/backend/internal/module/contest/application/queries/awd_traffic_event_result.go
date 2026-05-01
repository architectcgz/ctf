package queries

import "time"

type ListAWDTrafficEventsInput struct {
	AttackerTeamID int64
	VictimTeamID   int64
	ServiceID      int64
	AWDChallengeID int64
	StatusGroup    string
	PathKeyword    string
	Page           int
	Size           int
}

type AWDTrafficEventPageResult struct {
	List     []AWDTrafficEventResult
	Total    int64
	Page     int
	PageSize int
}

type AWDTrafficEventResult struct {
	ID                int64
	ContestID         int64
	RoundID           int64
	AttackerTeamID    int64
	AttackerTeam      string
	AttackerTeamName  string
	VictimTeamID      int64
	VictimTeam        string
	VictimTeamName    string
	ServiceID         int64
	AWDChallengeID    int64
	AWDChallengeTitle string
	Method            string
	Path              string
	StatusCode        int
	StatusGroup       string
	IsError           bool
	Source            string
	OccurredAt        time.Time
}
