package dto

import (
	"time"

	"ctf-platform/internal/model"
)

type CreateAWDRoundReq struct {
	RoundNumber    int     `json:"round_number" binding:"required,min=1"`
	Status         *string `json:"status" binding:"omitempty,oneof=pending running finished"`
	AttackScore    *int    `json:"attack_score" binding:"omitempty,min=0"`
	DefenseScore   *int    `json:"defense_score" binding:"omitempty,min=0"`
	ForceOverride  *bool   `json:"force_override"`
	OverrideReason *string `json:"override_reason" binding:"omitempty,max=500"`
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
	ServiceID     int64          `json:"service_id" binding:"required,min=1"`
	ServiceStatus string         `json:"service_status" binding:"required,oneof=up down compromised"`
	CheckResult   map[string]any `json:"check_result"`
}

type AWDTeamServiceResp struct {
	ID             int64                `json:"id"`
	RoundID        int64                `json:"round_id"`
	TeamID         int64                `json:"team_id"`
	TeamName       string               `json:"team_name"`
	ServiceID      int64                `json:"service_id"`
	ChallengeID    int64                `json:"challenge_id"`
	ServiceStatus  string               `json:"service_status"`
	CheckResult    map[string]any       `json:"check_result"`
	CheckerType    model.AWDCheckerType `json:"checker_type,omitempty"`
	AttackReceived int                  `json:"attack_received"`
	SLAScore       int                  `json:"sla_score"`
	DefenseScore   int                  `json:"defense_score"`
	AttackScore    int                  `json:"attack_score"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

type CreateAWDAttackLogReq struct {
	AttackerTeamID int64  `json:"attacker_team_id" binding:"required,min=1"`
	VictimTeamID   int64  `json:"victim_team_id" binding:"required,min=1"`
	ServiceID      int64  `json:"service_id" binding:"required,min=1"`
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
	ServiceID      int64     `json:"service_id"`
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
	SLAScore                int    `json:"sla_score"`
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

type RunCurrentAWDCheckerReq struct {
	ForceOverride  *bool   `json:"force_override"`
	OverrideReason *string `json:"override_reason" binding:"omitempty,max=500"`
}

type AWDReadinessResp struct {
	ContestID                int64                   `json:"contest_id"`
	Ready                    bool                    `json:"ready"`
	TotalChallenges          int                     `json:"total_challenges"`
	PassedChallenges         int                     `json:"passed_challenges"`
	PendingChallenges        int                     `json:"pending_challenges"`
	FailedChallenges         int                     `json:"failed_challenges"`
	StaleChallenges          int                     `json:"stale_challenges"`
	MissingCheckerChallenges int                     `json:"missing_checker_challenges"`
	BlockingCount            int                     `json:"blocking_count"`
	BlockingActions          []string                `json:"blocking_actions"`
	GlobalBlockingReasons    []string                `json:"global_blocking_reasons"`
	Items                    []*AWDReadinessItemResp `json:"items"`
}

type AWDReadinessItemResp struct {
	ServiceID       int64                `json:"service_id"`
	ChallengeID     int64                `json:"challenge_id"`
	Title           string               `json:"title"`
	CheckerType     model.AWDCheckerType `json:"checker_type"`
	ValidationState string               `json:"validation_state"`
	LastPreviewAt   *time.Time           `json:"last_preview_at"`
	LastAccessURL   *string              `json:"last_access_url"`
	BlockingReason  string               `json:"blocking_reason"`
}

type PreviewAWDCheckerReq struct {
	ServiceID        int64          `json:"service_id" binding:"omitempty,min=1"`
	ChallengeID      int64          `json:"challenge_id" binding:"omitempty,min=1"`
	CheckerType      string         `json:"checker_type" binding:"required,oneof=legacy_probe http_standard"`
	CheckerConfig    map[string]any `json:"checker_config"`
	AccessURL        string         `json:"access_url" binding:"omitempty,max=1024"`
	PreviewFlag      string         `json:"preview_flag" binding:"omitempty,max=512"`
	PreviewRequestID string         `json:"preview_request_id" binding:"omitempty,max=200"`
}

type AWDCheckerPreviewContextResp struct {
	ServiceID   int64  `json:"service_id"`
	AccessURL   string `json:"access_url"`
	PreviewFlag string `json:"preview_flag"`
	RoundNumber int    `json:"round_number"`
	TeamID      int64  `json:"team_id"`
	ChallengeID int64  `json:"challenge_id"`
}

type AWDCheckerPreviewResp struct {
	CheckerType    model.AWDCheckerType         `json:"checker_type,omitempty"`
	ServiceStatus  string                       `json:"service_status"`
	CheckResult    map[string]any               `json:"check_result"`
	PreviewContext AWDCheckerPreviewContextResp `json:"preview_context"`
	PreviewToken   string                       `json:"preview_token,omitempty"`
}

type ListAWDTrafficEventsReq struct {
	AttackerTeamID int64  `form:"attacker_team_id" binding:"omitempty,min=1"`
	VictimTeamID   int64  `form:"victim_team_id" binding:"omitempty,min=1"`
	ServiceID      int64  `form:"service_id" binding:"omitempty,min=1"`
	ChallengeID    int64  `form:"challenge_id" binding:"omitempty,min=1"`
	StatusGroup    string `form:"status_group" binding:"omitempty,oneof=success redirect client_error server_error"`
	PathKeyword    string `form:"path_keyword" binding:"omitempty,max=200"`
	Page           int    `form:"page" binding:"omitempty,min=1"`
	Size           int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type AWDTrafficTrendBucketResp struct {
	BucketStart  time.Time `json:"bucket_start_at"`
	RequestCount int       `json:"request_count"`
	ErrorCount   int       `json:"error_count"`
}

type AWDTrafficTopTeamResp struct {
	TeamID       int64  `json:"team_id"`
	TeamName     string `json:"team_name"`
	RequestCount int    `json:"request_count"`
	ErrorCount   int    `json:"error_count"`
}

type AWDTrafficTopChallengeResp struct {
	ChallengeID    int64  `json:"challenge_id"`
	ChallengeTitle string `json:"challenge_title"`
	RequestCount   int    `json:"request_count"`
	ErrorCount     int    `json:"error_count"`
}

type AWDTrafficTopPathResp struct {
	Path           string `json:"path"`
	RequestCount   int    `json:"request_count"`
	ErrorCount     int    `json:"error_count"`
	LastStatusCode int    `json:"last_status_code"`
}

type AWDTrafficSummaryResp struct {
	Round               *AWDRoundResp                 `json:"round"`
	ContestID           int64                         `json:"contest_id"`
	RoundID             int64                         `json:"round_id"`
	TotalRequests       int                           `json:"total_request_count"`
	ActiveAttackerTeams int                           `json:"active_attacker_team_count"`
	TargetedTeams       int                           `json:"victim_team_count"`
	ErrorRequests       int                           `json:"error_request_count"`
	UniquePathCount     int                           `json:"unique_path_count"`
	LatestEventAt       *time.Time                    `json:"latest_event_at,omitempty"`
	Trend               []*AWDTrafficTrendBucketResp  `json:"trend_buckets"`
	TopAttackers        []*AWDTrafficTopTeamResp      `json:"top_attackers"`
	TopVictims          []*AWDTrafficTopTeamResp      `json:"top_victims"`
	TopChallenges       []*AWDTrafficTopChallengeResp `json:"top_challenges"`
	TopPaths            []*AWDTrafficTopPathResp      `json:"top_paths"`
	TopErrorPaths       []*AWDTrafficTopPathResp      `json:"top_error_paths"`
}

type AWDTrafficEventResp struct {
	ID               int64     `json:"id"`
	ContestID        int64     `json:"contest_id"`
	RoundID          int64     `json:"round_id"`
	AttackerTeamID   int64     `json:"attacker_team_id"`
	AttackerTeam     string    `json:"-"`
	AttackerTeamName string    `json:"attacker_team_name"`
	VictimTeamID     int64     `json:"victim_team_id"`
	VictimTeam       string    `json:"-"`
	VictimTeamName   string    `json:"victim_team_name"`
	ServiceID        int64     `json:"service_id"`
	ChallengeID      int64     `json:"challenge_id"`
	ChallengeTitle   string    `json:"challenge_title"`
	Method           string    `json:"method"`
	Path             string    `json:"path"`
	StatusCode       int       `json:"status_code"`
	StatusGroup      string    `json:"status_group"`
	IsError          bool      `json:"is_error"`
	Source           string    `json:"source"`
	RequestID        string    `json:"request_id,omitempty"`
	OccurredAt       time.Time `json:"occurred_at"`
}

type AWDTrafficEventPageResp struct {
	List     []*AWDTrafficEventResp `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}
