package queries

import "time"

type AWDWorkspaceResult struct {
	ContestID    int64                            `json:"contest_id"`
	CurrentRound *AWDRoundResult                  `json:"current_round,omitempty"`
	MyTeam       *AWDWorkspaceTeamResult          `json:"my_team,omitempty"`
	Services     []*AWDWorkspaceServiceResult     `json:"services"`
	Targets      []*AWDWorkspaceTargetTeamResult  `json:"targets"`
	RecentEvents []*AWDWorkspaceRecentEventResult `json:"recent_events"`
}

type AWDWorkspaceTeamResult struct {
	TeamID   int64  `json:"team_id"`
	TeamName string `json:"team_name"`
}

type AWDWorkspaceServiceResult struct {
	ServiceID            int64                  `json:"service_id"`
	AWDChallengeID       int64                  `json:"awd_challenge_id"`
	InstanceID           int64                  `json:"instance_id,omitempty"`
	InstanceStatus       string                 `json:"instance_status,omitempty"`
	AccessURL            string                 `json:"access_url,omitempty"`
	ServiceStatus        string                 `json:"service_status,omitempty"`
	OperationStatus      string                 `json:"operation_status,omitempty"`
	OperationType        string                 `json:"operation_type,omitempty"`
	OperationReason      string                 `json:"operation_reason,omitempty"`
	OperationSLABillable *bool                  `json:"operation_sla_billable,omitempty"`
	CheckerType          string                 `json:"checker_type,omitempty"`
	AttackReceived       int                    `json:"attack_received"`
	SLAScore             int                    `json:"sla_score"`
	DefenseScore         int                    `json:"defense_score"`
	AttackScore          int                    `json:"attack_score"`
	DefenseScope         *AWDDefenseScopeResult `json:"defense_scope,omitempty"`
	UpdatedAt            *time.Time             `json:"updated_at,omitempty"`
}

type AWDDefenseScopeResult struct {
	EditablePaths    []string `json:"editable_paths,omitempty"`
	ProtectedPaths   []string `json:"protected_paths,omitempty"`
	ServiceContracts []string `json:"service_contracts,omitempty"`
}

type AWDWorkspaceTargetTeamResult struct {
	TeamID   int64                              `json:"team_id"`
	TeamName string                             `json:"team_name"`
	Services []*AWDWorkspaceTargetServiceResult `json:"services"`
}

type AWDWorkspaceTargetServiceResult struct {
	ServiceID      int64 `json:"service_id"`
	AWDChallengeID int64 `json:"awd_challenge_id"`
	Reachable      bool  `json:"reachable"`
}

type AWDWorkspaceRecentEventResult struct {
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
