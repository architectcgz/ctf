package http

type StartAdminContestAWDInstanceReq struct {
	TeamID    int64 `json:"team_id" binding:"required"`
	ServiceID int64 `json:"service_id" binding:"required"`
}

type PrewarmAdminContestAWDInstancesReq struct {
	TeamID *int64 `json:"team_id,omitempty"`
}

type SetAdminContestAWDTeamRetiredReq struct {
	Retired *bool  `json:"retired" binding:"required"`
	Reason  string `json:"reason" binding:"max=256"`
}

type SetAdminContestAWDServiceDisabledReq struct {
	Disabled *bool  `json:"disabled" binding:"required"`
	Reason   string `json:"reason" binding:"max=256"`
}

type SetAdminContestAWDDesiredReconcileSuppressedReq struct {
	Suppressed *bool  `json:"suppressed" binding:"required"`
	Reason     string `json:"reason" binding:"max=256"`
}
