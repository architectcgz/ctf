package http

import (
	"time"

	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestentity "ctf-platform/internal/module/contest/entity"
)

type PageResult[T any] struct {
	List  []T   `json:"list"`
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Size  int   `json:"page_size"`
}

type TeamResp struct {
	ID          int64     `json:"id"`
	ContestID   int64     `json:"contest_id"`
	Name        string    `json:"name"`
	CaptainID   int64     `json:"captain_id"`
	InviteCode  string    `json:"invite_code"`
	MaxMembers  int       `json:"max_members"`
	MemberCount int       `json:"member_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type TeamMemberResp struct {
	UserID   int64     `json:"user_id"`
	Username string    `json:"username"`
	JoinedAt time.Time `json:"joined_at"`
}

type MyTeamResp struct {
	ID         int64             `json:"id"`
	Name       string            `json:"name"`
	InviteCode string            `json:"invite_code"`
	CaptainID  int64             `json:"captain_user_id"`
	Members    []*TeamMemberResp `json:"members"`
}

type ContestChallengeResp struct {
	ID          int64     `json:"id"`
	ContestID   int64     `json:"contest_id"`
	ChallengeID int64     `json:"challenge_id"`
	Title       string    `json:"title,omitempty"`
	Category    string    `json:"category,omitempty"`
	Difficulty  string    `json:"difficulty,omitempty"`
	Points      int       `json:"points"`
	Order       int       `json:"order"`
	IsVisible   bool      `json:"is_visible"`
	CreatedAt   time.Time `json:"created_at"`
}

type ContestChallengeInfo struct {
	ID             int64  `json:"id"`
	ChallengeID    int64  `json:"challenge_id"`
	AWDChallengeID *int64 `json:"awd_challenge_id,omitempty"`
	AWDServiceID   *int64 `json:"awd_service_id,omitempty"`
	Title          string `json:"title"`
	Category       string `json:"category"`
	Difficulty     string `json:"difficulty"`
	Points         int    `json:"points"`
	Order          int    `json:"order"`
	SolvedCount    int64  `json:"solved_count"`
	IsSolved       bool   `json:"is_solved"`
}

type ContestRegistrationResp struct {
	ID         int64      `json:"id"`
	ContestID  int64      `json:"contest_id"`
	UserID     int64      `json:"user_id"`
	Username   string     `json:"username"`
	TeamID     *int64     `json:"team_id,omitempty"`
	Status     string     `json:"status"`
	ReviewedBy *int64     `json:"reviewed_by,omitempty"`
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

type ContestAnnouncementResp struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type ContestSolvedProgressItem struct {
	ContestChallengeID int64     `json:"contest_challenge_id"`
	SolvedAt           time.Time `json:"solved_at"`
	PointsEarned       int       `json:"points_earned"`
}

type ContestMyProgressResp struct {
	ContestID int64                        `json:"contest_id"`
	TeamID    *int64                       `json:"team_id,omitempty"`
	Solved    []*ContestSolvedProgressItem `json:"solved"`
}

type ContestListSummaryResp struct {
	DraftCount       int64 `json:"draft_count"`
	RegisteringCount int64 `json:"registering_count"`
	RunningCount     int64 `json:"running_count"`
	FrozenCount      int64 `json:"frozen_count"`
	EndedCount       int64 `json:"ended_count"`
}

type ContestPageResp struct {
	List     []*contestcmd.ContestResp `json:"list"`
	Total    int64                     `json:"total"`
	Page     int                       `json:"page"`
	PageSize int                       `json:"page_size"`
	Summary  ContestListSummaryResp    `json:"summary"`
}

type ScoreboardContestInfo struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	StartedAt time.Time `json:"started_at"`
	EndsAt    time.Time `json:"ends_at"`
}

type ScoreboardItem struct {
	Rank             int        `json:"rank"`
	TeamID           int64      `json:"team_id"`
	TeamName         string     `json:"team_name"`
	Score            float64    `json:"score"`
	SolvedCount      int        `json:"solved_count"`
	LastSubmissionAt *time.Time `json:"last_submission_at,omitempty"`
}

type ScoreboardPage struct {
	List     []*ScoreboardItem `json:"list"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	PageSize int               `json:"page_size"`
}

type ScoreboardResp struct {
	Contest    *ScoreboardContestInfo `json:"contest"`
	Scoreboard *ScoreboardPage        `json:"scoreboard"`
	Frozen     bool                   `json:"frozen"`
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

type AWDTeamServiceResp struct {
	ID                int64                        `json:"id"`
	RoundID           int64                        `json:"round_id"`
	TeamID            int64                        `json:"team_id"`
	TeamName          string                       `json:"team_name"`
	ServiceID         int64                        `json:"service_id"`
	ServiceName       string                       `json:"service_name,omitempty"`
	AWDChallengeID    int64                        `json:"awd_challenge_id"`
	AWDChallengeTitle string                       `json:"awd_challenge_title,omitempty"`
	ServiceStatus     string                       `json:"service_status"`
	CheckResult       map[string]any               `json:"check_result"`
	CheckerType       contestentity.AWDCheckerType `json:"checker_type,omitempty"`
	AttackReceived    int                          `json:"attack_received"`
	SLAScore          int                          `json:"sla_score"`
	DefenseScore      int                          `json:"defense_score"`
	AttackScore       int                          `json:"attack_score"`
	UpdatedAt         time.Time                    `json:"updated_at"`
}

type AWDAttackLogResp struct {
	ID             int64     `json:"id"`
	RoundID        int64     `json:"round_id"`
	AttackerTeamID int64     `json:"attacker_team_id"`
	AttackerTeam   string    `json:"attacker_team"`
	VictimTeamID   int64     `json:"victim_team_id"`
	VictimTeam     string    `json:"victim_team"`
	ServiceID      int64     `json:"service_id"`
	AWDChallengeID int64     `json:"awd_challenge_id"`
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

type AWDReadinessItemResp struct {
	ServiceID       int64                        `json:"service_id"`
	AWDChallengeID  int64                        `json:"awd_challenge_id"`
	Title           string                       `json:"title"`
	CheckerType     contestentity.AWDCheckerType `json:"checker_type"`
	ValidationState string                       `json:"validation_state"`
	LastPreviewAt   *time.Time                   `json:"last_preview_at"`
	LastAccessURL   *string                      `json:"last_access_url"`
	BlockingReason  string                       `json:"blocking_reason"`
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
	AWDChallengeID    int64  `json:"awd_challenge_id"`
	AWDChallengeTitle string `json:"awd_challenge_title"`
	RequestCount      int    `json:"request_count"`
	ErrorCount        int    `json:"error_count"`
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
	ID                int64     `json:"id"`
	ContestID         int64     `json:"contest_id"`
	RoundID           int64     `json:"round_id"`
	AttackerTeamID    int64     `json:"attacker_team_id"`
	AttackerTeam      string    `json:"-"`
	AttackerTeamName  string    `json:"attacker_team_name"`
	VictimTeamID      int64     `json:"victim_team_id"`
	VictimTeam        string    `json:"-"`
	VictimTeamName    string    `json:"victim_team_name"`
	ServiceID         int64     `json:"service_id"`
	AWDChallengeID    int64     `json:"awd_challenge_id"`
	AWDChallengeTitle string    `json:"awd_challenge_title"`
	Method            string    `json:"method"`
	Path              string    `json:"path"`
	StatusCode        int       `json:"status_code"`
	StatusGroup       string    `json:"status_group"`
	IsError           bool      `json:"is_error"`
	Source            string    `json:"source"`
	RequestID         string    `json:"request_id,omitempty"`
	OccurredAt        time.Time `json:"occurred_at"`
}

type AWDTrafficEventPageResp struct {
	List     []*AWDTrafficEventResp `json:"list"`
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
}

type ContestAWDWorkspaceResp struct {
	ContestID    int64                                 `json:"contest_id"`
	CurrentRound *AWDRoundResp                         `json:"current_round,omitempty"`
	MyTeam       *ContestAWDWorkspaceTeamResp          `json:"my_team,omitempty"`
	Services     []*ContestAWDWorkspaceServiceResp     `json:"services"`
	Targets      []*ContestAWDWorkspaceTargetTeamResp  `json:"targets"`
	RecentEvents []*ContestAWDWorkspaceRecentEventResp `json:"recent_events"`
}

type ContestAWDWorkspaceTeamResp struct {
	TeamID   int64  `json:"team_id"`
	TeamName string `json:"team_name"`
}

type ContestAWDWorkspaceServiceResp struct {
	ServiceID            int64                        `json:"service_id"`
	AWDChallengeID       int64                        `json:"awd_challenge_id"`
	InstanceID           int64                        `json:"instance_id,omitempty"`
	InstanceStatus       string                       `json:"instance_status,omitempty"`
	AccessURL            string                       `json:"access_url,omitempty"`
	ServiceStatus        string                       `json:"service_status,omitempty"`
	OperationStatus      string                       `json:"operation_status,omitempty"`
	OperationType        string                       `json:"operation_type,omitempty"`
	OperationReason      string                       `json:"operation_reason,omitempty"`
	OperationSLABillable *bool                        `json:"operation_sla_billable,omitempty"`
	CheckerType          contestentity.AWDCheckerType `json:"checker_type,omitempty"`
	AttackReceived       int                          `json:"attack_received"`
	SLAScore             int                          `json:"sla_score"`
	DefenseScore         int                          `json:"defense_score"`
	AttackScore          int                          `json:"attack_score"`
	DefenseConnection    *AWDDefenseConnectionResp    `json:"defense_connection,omitempty"`
	UpdatedAt            *time.Time                   `json:"updated_at,omitempty"`
}

type AWDDefenseConnectionResp struct {
	EntryMode         string `json:"entry_mode,omitempty"`
	WorkspaceStatus   string `json:"workspace_status,omitempty"`
	WorkspaceRevision int64  `json:"workspace_revision,omitempty"`
}

type ContestAWDWorkspaceTargetTeamResp struct {
	TeamID   int64                                   `json:"team_id"`
	TeamName string                                  `json:"team_name"`
	Services []*ContestAWDWorkspaceTargetServiceResp `json:"services"`
}

type ContestAWDWorkspaceTargetServiceResp struct {
	ServiceID      int64 `json:"service_id"`
	AWDChallengeID int64 `json:"awd_challenge_id"`
	Reachable      bool  `json:"reachable"`
}

type ContestAWDWorkspaceRecentEventResp struct {
	ID             int64     `json:"id"`
	Direction      string    `json:"direction"`
	ServiceID      int64     `json:"service_id"`
	AWDChallengeID int64     `json:"awd_challenge_id"`
	PeerTeamID     int64     `json:"peer_team_id"`
	PeerTeamName   string    `json:"peer_team_name"`
	IsSuccess      bool      `json:"is_success"`
	ScoreGained    int       `json:"score_gained"`
	CreatedAt      time.Time `json:"created_at"`
}
