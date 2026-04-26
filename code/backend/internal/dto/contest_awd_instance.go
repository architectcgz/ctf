package dto

type AdminAWDInstanceTeamResp struct {
	TeamID    int64  `json:"team_id"`
	TeamName  string `json:"team_name"`
	CaptainID int64  `json:"captain_id"`
}

type AdminAWDInstanceServiceResp struct {
	ServiceID   int64  `json:"service_id"`
	ChallengeID int64  `json:"challenge_id"`
	DisplayName string `json:"display_name"`
	IsVisible   bool   `json:"is_visible"`
}

type AdminAWDInstanceItemResp struct {
	TeamID    int64         `json:"team_id"`
	ServiceID int64         `json:"service_id"`
	Instance  *InstanceResp `json:"instance,omitempty"`
}

type AdminAWDInstanceOrchestrationResp struct {
	ContestID int64                          `json:"contest_id"`
	Teams     []*AdminAWDInstanceTeamResp    `json:"teams"`
	Services  []*AdminAWDInstanceServiceResp `json:"services"`
	Instances []*AdminAWDInstanceItemResp    `json:"instances"`
}

type StartAdminContestAWDInstanceReq struct {
	TeamID    int64 `json:"team_id" binding:"required"`
	ServiceID int64 `json:"service_id" binding:"required"`
}
