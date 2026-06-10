package http

import "time"

type CreateContestReq struct {
	Title       string    `json:"title" binding:"required,min=1,max=200"`
	Description string    `json:"description" binding:"max=5000"`
	Mode        string    `json:"mode" binding:"required,oneof=jeopardy awd"`
	StartTime   time.Time `json:"start_time" binding:"required"`
	EndTime     time.Time `json:"end_time" binding:"required,gtfield=StartTime"`
}

type UpdateContestReq struct {
	Title          *string    `json:"title" binding:"omitempty,min=1,max=200"`
	Description    *string    `json:"description" binding:"omitempty,max=5000"`
	Mode           *string    `json:"mode" binding:"omitempty,oneof=jeopardy awd"`
	StartTime      *time.Time `json:"start_time"`
	EndTime        *time.Time `json:"end_time"`
	Status         *string    `json:"status" binding:"omitempty,oneof=draft registration running frozen ended"`
	ForceOverride  *bool      `json:"force_override"`
	OverrideReason *string    `json:"override_reason" binding:"omitempty,max=500"`
}

type ListContestsReq struct {
	Status    *string `form:"status" binding:"omitempty,oneof=draft registration running frozen ended"`
	Statuses  string  `form:"statuses" binding:"omitempty,max=128"`
	Mode      *string `form:"mode" binding:"omitempty,oneof=jeopardy awd"`
	SortKey   string  `form:"sort_key" binding:"omitempty,oneof=created_at start_time"`
	SortOrder string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
	Page      int     `form:"page" binding:"omitempty,min=1"`
	Size      int     `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type FreezeReq struct {
	MinutesBeforeEnd int `json:"minutes_before_end" binding:"required,min=1"`
}

type CreateContestAnnouncementReq struct {
	Title   string `json:"title" binding:"required,min=1,max=200"`
	Content string `json:"content" binding:"max=5000"`
}

type ContestRegistrationQuery struct {
	Status *string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	Page   int     `form:"page" binding:"omitempty,min=1"`
	Size   int     `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type ContestAnnouncementSyncQuery struct {
	AfterID int64 `form:"after_id" binding:"omitempty,min=1"`
}

type ReviewContestRegistrationReq struct {
	Status string `json:"status" binding:"required,oneof=approved rejected"`
}

type CreateTeamReq struct {
	Name       string `json:"name" binding:"required,min=2,max=50"`
	MaxMembers int    `json:"max_members" binding:"omitempty,min=2,max=10"`
}

type AddContestChallengeReq struct {
	ChallengeID int64 `json:"challenge_id" binding:"required"`
	Points      int   `json:"points" binding:"omitempty,min=1"`
	Order       int   `json:"order" binding:"omitempty,min=0"`
	IsVisible   *bool `json:"is_visible"`
}

type UpdateContestChallengeReq struct {
	Points    *int  `json:"points" binding:"omitempty,min=1"`
	Order     *int  `json:"order" binding:"omitempty,min=0"`
	IsVisible *bool `json:"is_visible"`
}

type CreateAWDRoundReq struct {
	RoundNumber    int     `json:"round_number" binding:"required,min=1"`
	Status         *string `json:"status" binding:"omitempty,oneof=pending running finished"`
	AttackScore    *int    `json:"attack_score" binding:"omitempty,min=0,max=100"`
	DefenseScore   *int    `json:"defense_score" binding:"omitempty,min=0,max=10"`
	ForceOverride  *bool   `json:"force_override"`
	OverrideReason *string `json:"override_reason" binding:"omitempty,max=500"`
}

type UpsertAWDServiceCheckReq struct {
	TeamID        int64          `json:"team_id" binding:"required,min=1"`
	ServiceID     int64          `json:"service_id" binding:"required,min=1"`
	ServiceStatus string         `json:"service_status" binding:"required,oneof=up down compromised"`
	CheckResult   map[string]any `json:"check_result"`
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

type RunCurrentAWDCheckerReq struct {
	ForceOverride  *bool   `json:"force_override"`
	OverrideReason *string `json:"override_reason" binding:"omitempty,max=500"`
}

type PreviewAWDCheckerReq struct {
	ServiceID        int64          `json:"service_id" binding:"omitempty,min=1"`
	AWDChallengeID   int64          `json:"awd_challenge_id" binding:"omitempty,min=1"`
	CheckerType      string         `json:"checker_type" binding:"required,oneof=legacy_probe http_standard tcp_standard script_checker"`
	CheckerConfig    map[string]any `json:"checker_config"`
	AccessURL        string         `json:"access_url" binding:"omitempty,max=1024"`
	PreviewFlag      string         `json:"preview_flag" binding:"omitempty,max=512"`
	PreviewRequestID string         `json:"preview_request_id" binding:"omitempty,max=200"`
}

type CreateContestAWDServiceReq struct {
	AWDChallengeID         int64          `json:"awd_challenge_id" binding:"required,min=1"`
	Points                 int            `json:"points" binding:"required,min=1,max=500"`
	DisplayName            string         `json:"display_name" binding:"omitempty,max=128"`
	Order                  int            `json:"order" binding:"omitempty,min=0"`
	IsVisible              *bool          `json:"is_visible"`
	CheckerType            *string        `json:"checker_type" binding:"omitempty,oneof=legacy_probe http_standard tcp_standard script_checker"`
	CheckerConfig          map[string]any `json:"checker_config"`
	AWDSLAScore            *int           `json:"awd_sla_score" binding:"omitempty,min=0,max=5"`
	AWDDefenseScore        *int           `json:"awd_defense_score" binding:"omitempty,min=0,max=5"`
	AWDCheckerPreviewToken *string        `json:"awd_checker_preview_token" binding:"omitempty,max=200"`
}

type UpdateContestAWDServiceReq struct {
	AWDChallengeID         *int64         `json:"awd_challenge_id" binding:"omitempty,min=1"`
	Points                 *int           `json:"points" binding:"omitempty,min=1,max=500"`
	DisplayName            *string        `json:"display_name" binding:"omitempty,max=128"`
	Order                  *int           `json:"order" binding:"omitempty,min=0"`
	IsVisible              *bool          `json:"is_visible"`
	CheckerType            *string        `json:"checker_type" binding:"omitempty,oneof=legacy_probe http_standard tcp_standard script_checker"`
	CheckerConfig          map[string]any `json:"checker_config"`
	AWDSLAScore            *int           `json:"awd_sla_score" binding:"omitempty,min=0,max=5"`
	AWDDefenseScore        *int           `json:"awd_defense_score" binding:"omitempty,min=0,max=5"`
	AWDCheckerPreviewToken *string        `json:"awd_checker_preview_token" binding:"omitempty,max=200"`
}

type ListAWDTrafficEventsReq struct {
	AttackerTeamID int64  `form:"attacker_team_id" binding:"omitempty,min=1"`
	VictimTeamID   int64  `form:"victim_team_id" binding:"omitempty,min=1"`
	ServiceID      int64  `form:"service_id" binding:"omitempty,min=1"`
	AWDChallengeID int64  `form:"awd_challenge_id" binding:"omitempty,min=1"`
	StatusGroup    string `form:"status_group" binding:"omitempty,oneof=success redirect client_error server_error"`
	PathKeyword    string `form:"path_keyword" binding:"omitempty,max=200"`
	Page           int    `form:"page" binding:"omitempty,min=1"`
	Size           int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}
