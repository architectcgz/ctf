package dto

import "time"

type GetTeacherAWDReviewArchiveReq struct {
	RoundNumber *int   `form:"round" binding:"omitempty,min=1"`
	TeamID      *int64 `form:"team_id"`
}

type TeacherAWDReviewContestListResp struct {
	Contests []TeacherAWDReviewContestResp `json:"contests"`
}

type TeacherAWDReviewContestResp struct {
	ID               int64      `json:"id"`
	Title            string     `json:"title"`
	Mode             string     `json:"mode"`
	Status           string     `json:"status"`
	CurrentRound     *int       `json:"current_round,omitempty"`
	RoundCount       int        `json:"round_count"`
	TeamCount        int        `json:"team_count"`
	LatestEvidenceAt *time.Time `json:"latest_evidence_at,omitempty"`
	ExportReady      bool       `json:"export_ready"`
}

type TeacherAWDReviewArchiveResp struct {
	GeneratedAt   time.Time                       `json:"generated_at"`
	Scope         TeacherAWDReviewScopeResp       `json:"scope"`
	Contest       TeacherAWDReviewContestMetaResp `json:"contest"`
	Overview      *TeacherAWDReviewOverviewResp   `json:"overview,omitempty"`
	Rounds        []TeacherAWDReviewRoundResp     `json:"rounds"`
	SelectedRound *TeacherAWDSelectedRoundResp    `json:"selected_round,omitempty"`
}

type TeacherAWDReviewScopeResp struct {
	SnapshotType  string `json:"snapshot_type"`
	RequestedBy   int64  `json:"requested_by"`
	RequestedRole string `json:"requested_role,omitempty"`
	RequestedID   int64  `json:"requested_id"`
}

type TeacherAWDReviewContestMetaResp struct {
	ID               int64      `json:"id"`
	Title            string     `json:"title"`
	Mode             string     `json:"mode"`
	Status           string     `json:"status"`
	CurrentRound     *int       `json:"current_round,omitempty"`
	RoundCount       int        `json:"round_count"`
	TeamCount        int        `json:"team_count"`
	LatestEvidenceAt *time.Time `json:"latest_evidence_at,omitempty"`
	ExportReady      bool       `json:"export_ready"`
}

type TeacherAWDReviewOverviewResp struct {
	RoundCount       int        `json:"round_count"`
	TeamCount        int        `json:"team_count"`
	ServiceCount     int        `json:"service_count"`
	AttackCount      int        `json:"attack_count"`
	TrafficCount     int        `json:"traffic_count"`
	LatestEvidenceAt *time.Time `json:"latest_evidence_at,omitempty"`
}

type TeacherAWDReviewRoundResp struct {
	ID           int64      `json:"id"`
	ContestID    int64      `json:"contest_id"`
	RoundNumber  int        `json:"round_number"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	AttackScore  int        `json:"attack_score"`
	DefenseScore int        `json:"defense_score"`
	ServiceCount int        `json:"service_count"`
	AttackCount  int        `json:"attack_count"`
	TrafficCount int        `json:"traffic_count"`
}

type TeacherAWDSelectedRoundResp struct {
	Round    TeacherAWDReviewRoundResp     `json:"round"`
	Teams    []TeacherAWDReviewTeamResp    `json:"teams"`
	Services []TeacherAWDReviewServiceResp `json:"services"`
	Attacks  []TeacherAWDReviewAttackResp  `json:"attacks"`
	Traffic  []TeacherAWDReviewTrafficResp `json:"traffic"`
}

type TeacherAWDReviewTeamResp struct {
	TeamID      int64      `json:"team_id"`
	TeamName    string     `json:"team_name"`
	CaptainID   int64      `json:"captain_id"`
	TotalScore  int        `json:"total_score"`
	MemberCount int        `json:"member_count"`
	LastSolveAt *time.Time `json:"last_solve_at,omitempty"`
}

type TeacherAWDReviewServiceResp struct {
	ID             int64     `json:"id"`
	RoundID        int64     `json:"round_id"`
	TeamID         int64     `json:"team_id"`
	TeamName       string    `json:"team_name"`
	ServiceID      int64     `json:"service_id"`
	ChallengeID    int64     `json:"challenge_id"`
	ChallengeTitle string    `json:"challenge_title"`
	ServiceStatus  string    `json:"service_status"`
	AttackReceived int       `json:"attack_received"`
	SLAScore       int       `json:"sla_score"`
	DefenseScore   int       `json:"defense_score"`
	AttackScore    int       `json:"attack_score"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type TeacherAWDReviewAttackResp struct {
	ID               int64     `json:"id"`
	RoundID          int64     `json:"round_id"`
	AttackerTeamID   int64     `json:"attacker_team_id"`
	AttackerTeamName string    `json:"attacker_team_name"`
	VictimTeamID     int64     `json:"victim_team_id"`
	VictimTeamName   string    `json:"victim_team_name"`
	ServiceID        int64     `json:"service_id"`
	ChallengeID      int64     `json:"challenge_id"`
	ChallengeTitle   string    `json:"challenge_title"`
	AttackType       string    `json:"attack_type"`
	Source           string    `json:"source"`
	SubmittedFlag    string    `json:"submitted_flag,omitempty"`
	IsSuccess        bool      `json:"is_success"`
	ScoreGained      int       `json:"score_gained"`
	CreatedAt        time.Time `json:"created_at"`
}

type TeacherAWDReviewTrafficResp struct {
	ID               int64     `json:"id"`
	ContestID        int64     `json:"contest_id"`
	RoundID          int64     `json:"round_id"`
	AttackerTeamID   int64     `json:"attacker_team_id"`
	AttackerTeamName string    `json:"attacker_team_name"`
	VictimTeamID     int64     `json:"victim_team_id"`
	VictimTeamName   string    `json:"victim_team_name"`
	ServiceID        int64     `json:"service_id"`
	ChallengeID      int64     `json:"challenge_id"`
	ChallengeTitle   string    `json:"challenge_title"`
	Method           string    `json:"method"`
	Path             string    `json:"path"`
	StatusCode       int       `json:"status_code"`
	Source           string    `json:"source"`
	CreatedAt        time.Time `json:"created_at"`
}
