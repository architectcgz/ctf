package dto

import (
	"time"

	"ctf-platform/internal/model"
)

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
	ServiceID            int64                     `json:"service_id"`
	AWDChallengeID       int64                     `json:"awd_challenge_id"`
	InstanceID           int64                     `json:"instance_id,omitempty"`
	InstanceStatus       string                    `json:"instance_status,omitempty"`
	AccessURL            string                    `json:"access_url,omitempty"`
	ServiceStatus        string                    `json:"service_status,omitempty"`
	OperationStatus      string                    `json:"operation_status,omitempty"`
	OperationType        string                    `json:"operation_type,omitempty"`
	OperationReason      string                    `json:"operation_reason,omitempty"`
	OperationSLABillable *bool                     `json:"operation_sla_billable,omitempty"`
	CheckerType          model.AWDCheckerType      `json:"checker_type,omitempty"`
	AttackReceived       int                       `json:"attack_received"`
	SLAScore             int                       `json:"sla_score"`
	DefenseScore         int                       `json:"defense_score"`
	AttackScore          int                       `json:"attack_score"`
	DefenseConnection    *AWDDefenseConnectionResp `json:"defense_connection,omitempty"`
	UpdatedAt            *time.Time                `json:"updated_at,omitempty"`
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
