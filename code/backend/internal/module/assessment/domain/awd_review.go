package domain

import "time"

type TeacherAWDReviewContestCard struct {
	ID               int64
	Title            string
	Mode             string
	Status           string
	CurrentRound     *int
	RoundCount       int
	TeamCount        int
	LatestEvidenceAt *time.Time
	ExportReady      bool
}

type TeacherAWDReviewContestMeta struct {
	ID               int64
	Title            string
	Mode             string
	Status           string
	CurrentRound     *int
	RoundCount       int
	TeamCount        int
	LatestEvidenceAt *time.Time
	ExportReady      bool
}

type TeacherAWDReviewRoundSummary struct {
	ID           int64
	ContestID    int64
	RoundNumber  int
	Status       string
	StartedAt    *time.Time
	EndedAt      *time.Time
	AttackScore  int
	DefenseScore int
}

type TeacherAWDReviewTeamSummary struct {
	TeamID      int64
	TeamName    string
	CaptainID   int64
	TotalScore  int
	MemberCount int
	LastSolveAt *time.Time
}

type TeacherAWDReviewServiceRecord struct {
	ID                int64
	RoundID           int64
	TeamID            int64
	TeamName          string
	ServiceID         int64
	AWDChallengeID    int64
	AWDChallengeTitle string
	ServiceStatus     string
	AttackReceived    int
	SLAScore          int
	DefenseScore      int
	AttackScore       int
	UpdatedAt         time.Time
}

type TeacherAWDReviewAttackRecord struct {
	ID                int64
	RoundID           int64
	AttackerTeamID    int64
	AttackerTeamName  string
	VictimTeamID      int64
	VictimTeamName    string
	ServiceID         int64
	AWDChallengeID    int64
	AWDChallengeTitle string
	AttackType        string
	Source            string
	SubmittedFlag     string
	IsSuccess         bool
	ScoreGained       int
	CreatedAt         time.Time
}

type TeacherAWDReviewTrafficRecord struct {
	ID                int64
	ContestID         int64
	RoundID           int64
	AttackerTeamID    int64
	AttackerTeamName  string
	VictimTeamID      int64
	VictimTeamName    string
	ServiceID         int64
	AWDChallengeID    int64
	AWDChallengeTitle string
	Method            string
	Path              string
	StatusCode        int
	Source            string
	CreatedAt         time.Time
}
