package dto

import "time"

type CreateAWDRoundReq struct {
	RoundNumber  int     `json:"round_number" binding:"required,min=1"`
	Status       *string `json:"status" binding:"omitempty,oneof=pending running finished"`
	AttackScore  *int    `json:"attack_score" binding:"omitempty,min=0"`
	DefenseScore *int    `json:"defense_score" binding:"omitempty,min=0"`
}

type AWDRoundResp struct {
	ID           int64      `json:"id"`
	ContestID    int64      `json:"contest_id"`
	RoundNumber  int        `json:"round_number"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	EndedAt      *time.Time `json:"ended_at,omitempty"`
	AttackScore  int        `json:"attack_score"`
	DefenseScore int        `json:"defense_score"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type UpsertAWDServiceCheckReq struct {
	TeamID        int64          `json:"team_id" binding:"required,min=1"`
	ChallengeID   int64          `json:"challenge_id" binding:"required,min=1"`
	ServiceStatus string         `json:"service_status" binding:"required,oneof=up down compromised"`
	CheckResult   map[string]any `json:"check_result"`
}

type AWDTeamServiceResp struct {
	ID             int64          `json:"id"`
	RoundID        int64          `json:"round_id"`
	TeamID         int64          `json:"team_id"`
	TeamName       string         `json:"team_name"`
	ChallengeID    int64          `json:"challenge_id"`
	ServiceStatus  string         `json:"service_status"`
	CheckResult    map[string]any `json:"check_result"`
	AttackReceived int            `json:"attack_received"`
	DefenseScore   int            `json:"defense_score"`
	AttackScore    int            `json:"attack_score"`
	UpdatedAt      time.Time      `json:"updated_at"`
}

type CreateAWDAttackLogReq struct {
	AttackerTeamID int64  `json:"attacker_team_id" binding:"required,min=1"`
	VictimTeamID   int64  `json:"victim_team_id" binding:"required,min=1"`
	ChallengeID    int64  `json:"challenge_id" binding:"required,min=1"`
	AttackType     string `json:"attack_type" binding:"required,oneof=flag_capture service_exploit"`
	SubmittedFlag  string `json:"submitted_flag" binding:"omitempty,max=512"`
	IsSuccess      bool   `json:"is_success"`
}

type SubmitAWDAttackReq struct {
	VictimTeamID int64  `json:"victim_team_id" binding:"required,min=1"`
	Flag         string `json:"flag" binding:"required,max=512"`
}

type AWDAttackLogResp struct {
	ID             int64     `json:"id"`
	RoundID        int64     `json:"round_id"`
	AttackerTeamID int64     `json:"attacker_team_id"`
	AttackerTeam   string    `json:"attacker_team"`
	VictimTeamID   int64     `json:"victim_team_id"`
	VictimTeam     string    `json:"victim_team"`
	ChallengeID    int64     `json:"challenge_id"`
	AttackType     string    `json:"attack_type"`
	Source         string    `json:"source"`
	SubmittedFlag  string    `json:"submitted_flag,omitempty"`
	IsSuccess      bool      `json:"is_success"`
	ScoreGained    int       `json:"score_gained"`
	CreatedAt      time.Time `json:"created_at"`
}

type AWDRoundSummaryItem struct {
	TeamID                  int64  `json:"team_id"`
	TeamName                string `json:"team_name"`
	ServiceUpCount          int    `json:"service_up_count"`
	ServiceDownCount        int    `json:"service_down_count"`
	ServiceCompromisedCount int    `json:"service_compromised_count"`
	DefenseScore            int    `json:"defense_score"`
	AttackScore             int    `json:"attack_score"`
	SuccessfulAttackCount   int    `json:"successful_attack_count"`
	SuccessfulBreachCount   int    `json:"successful_breach_count"`
	UniqueAttackersAgainst  int    `json:"unique_attackers_against"`
	TotalScore              int    `json:"total_score"`
}

type AWDRoundMetrics struct {
	TotalServiceCount         int `json:"total_service_count"`
	ServiceUpCount            int `json:"service_up_count"`
	ServiceDownCount          int `json:"service_down_count"`
	ServiceCompromisedCount   int `json:"service_compromised_count"`
	AttackedServiceCount      int `json:"attacked_service_count"`
	DefenseSuccessCount       int `json:"defense_success_count"`
	TotalAttackCount          int `json:"total_attack_count"`
	SuccessfulAttackCount     int `json:"successful_attack_count"`
	FailedAttackCount         int `json:"failed_attack_count"`
	SchedulerCheckCount       int `json:"scheduler_check_count"`
	ManualCurrentRoundChecks  int `json:"manual_current_round_check_count"`
	ManualSelectedRoundChecks int `json:"manual_selected_round_check_count"`
	ManualServiceCheckCount   int `json:"manual_service_check_count"`
	SubmissionAttackCount     int `json:"submission_attack_count"`
	ManualAttackLogCount      int `json:"manual_attack_log_count"`
	LegacyAttackLogCount      int `json:"legacy_attack_log_count"`
}

type AWDRoundSummaryResp struct {
	Round   *AWDRoundResp          `json:"round"`
	Metrics *AWDRoundMetrics       `json:"metrics,omitempty"`
	Items   []*AWDRoundSummaryItem `json:"items"`
}

type AWDCheckerRunResp struct {
	Round    *AWDRoundResp         `json:"round"`
	Services []*AWDTeamServiceResp `json:"services"`
}
