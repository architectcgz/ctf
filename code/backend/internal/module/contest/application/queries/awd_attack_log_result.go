package queries

import "time"

type AWDAttackLogResult struct {
	ID             int64
	RoundID        int64
	AttackerTeamID int64
	AttackerTeam   string
	VictimTeamID   int64
	VictimTeam     string
	ServiceID      int64
	AWDChallengeID int64
	AttackType     string
	Source         string
	SubmittedFlag  string
	IsSuccess      bool
	ScoreGained    int
	CreatedAt      time.Time
}
