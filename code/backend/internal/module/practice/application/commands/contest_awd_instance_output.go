package commands

import (
	"time"

	instancecontracts "ctf-platform/internal/module/instance/contracts"
)

type AdminAWDInstanceTeamResp struct {
	TeamID    int64  `json:"team_id"`
	TeamName  string `json:"team_name"`
	CaptainID int64  `json:"captain_id"`
}

type AdminAWDInstanceServiceResp struct {
	ServiceID      int64  `json:"service_id"`
	AWDChallengeID int64  `json:"awd_challenge_id"`
	DisplayName    string `json:"display_name"`
	IsVisible      bool   `json:"is_visible"`
}

type AdminAWDInstanceItemResp struct {
	TeamID    int64                           `json:"team_id"`
	ServiceID int64                           `json:"service_id"`
	Instance  *instancecontracts.InstanceResp `json:"instance,omitempty"`
}

type AdminAWDScopeControlResp struct {
	ScopeType   string     `json:"scope_type"`
	ControlType string     `json:"control_type"`
	TeamID      int64      `json:"team_id"`
	ServiceID   *int64     `json:"service_id,omitempty"`
	Enabled     bool       `json:"enabled"`
	Reason      string     `json:"reason,omitempty"`
	UpdatedBy   *int64     `json:"updated_by,omitempty"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

type AdminAWDInstanceOrchestrationResp struct {
	ContestID int64                          `json:"contest_id"`
	Teams     []*AdminAWDInstanceTeamResp    `json:"teams"`
	Services  []*AdminAWDInstanceServiceResp `json:"services"`
	Instances []*AdminAWDInstanceItemResp    `json:"instances"`
	Controls  []*AdminAWDScopeControlResp    `json:"controls"`
}

type AdminAWDInstancePrewarmItemResp struct {
	TeamID       int64                           `json:"team_id"`
	ServiceID    int64                           `json:"service_id"`
	Outcome      string                          `json:"outcome"`
	Instance     *instancecontracts.InstanceResp `json:"instance,omitempty"`
	ErrorMessage string                          `json:"error_message,omitempty"`
}

type AdminAWDInstancePrewarmSummaryResp struct {
	Total   int `json:"total"`
	Started int `json:"started"`
	Reused  int `json:"reused"`
	Failed  int `json:"failed"`
}

type AdminAWDInstancePrewarmResp struct {
	ContestID int64                              `json:"contest_id"`
	Results   []*AdminAWDInstancePrewarmItemResp `json:"results"`
	Summary   AdminAWDInstancePrewarmSummaryResp `json:"summary"`
}
