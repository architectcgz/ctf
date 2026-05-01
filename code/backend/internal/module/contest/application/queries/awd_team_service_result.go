package queries

import "time"

type AWDTeamServiceResult struct {
	ID                int64
	RoundID           int64
	TeamID            int64
	TeamName          string
	ServiceID         int64
	ServiceName       string
	AWDChallengeID    int64
	AWDChallengeTitle string
	ServiceStatus     string
	CheckResult       map[string]any
	CheckerType       string
	AttackReceived    int
	SLAScore          int
	DefenseScore      int
	AttackScore       int
	UpdatedAt         time.Time
}
