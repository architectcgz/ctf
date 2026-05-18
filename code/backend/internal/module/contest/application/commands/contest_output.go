package commands

import (
	"time"

	contestentity "ctf-platform/internal/module/contest/entity"
)

type ContestResp struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Mode        string     `json:"mode"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	FreezeTime  *time.Time `json:"freeze_time,omitempty"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
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

type AWDCheckerPreviewContextResp struct {
	ServiceID      int64  `json:"service_id"`
	AccessURL      string `json:"access_url"`
	PreviewFlag    string `json:"preview_flag"`
	RoundNumber    int    `json:"round_number"`
	TeamID         int64  `json:"team_id"`
	AWDChallengeID int64  `json:"awd_challenge_id"`
}

type AWDCheckerPreviewResp struct {
	CheckerType    contestentity.AWDCheckerType `json:"checker_type,omitempty"`
	ServiceStatus  string                       `json:"service_status"`
	CheckResult    map[string]any               `json:"check_result"`
	PreviewContext AWDCheckerPreviewContextResp `json:"preview_context"`
	PreviewToken   string                       `json:"preview_token,omitempty"`
}

type ContestAWDServiceResp struct {
	ID                int64                                   `json:"id"`
	ContestID         int64                                   `json:"contest_id"`
	AWDChallengeID    int64                                   `json:"awd_challenge_id"`
	Title             string                                  `json:"title,omitempty"`
	Category          string                                  `json:"category,omitempty"`
	Difficulty        string                                  `json:"difficulty,omitempty"`
	DisplayName       string                                  `json:"display_name"`
	Order             int                                     `json:"order"`
	IsVisible         bool                                    `json:"is_visible"`
	ScoreConfig       map[string]any                          `json:"score_config,omitempty"`
	RuntimeConfig     map[string]any                          `json:"runtime_config,omitempty"`
	ValidationState   contestentity.AWDCheckerValidationState `json:"validation_state"`
	LastPreviewAt     *time.Time                              `json:"last_preview_at,omitempty"`
	LastPreviewResult *AWDCheckerPreviewResp                  `json:"last_preview_result,omitempty"`
	CreatedAt         time.Time                               `json:"created_at"`
	UpdatedAt         time.Time                               `json:"updated_at"`
}

type AWDCheckerRunResp struct {
	Round    *AWDRoundResp         `json:"round"`
	Services []*AWDTeamServiceResp `json:"services"`
}
