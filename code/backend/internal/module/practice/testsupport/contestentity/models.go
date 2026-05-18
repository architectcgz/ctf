package contestentity

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

const (
	ContestModeJeopardy = "jeopardy"
	ContestModeAWD      = "awd"

	ContestStatusDraft        = "draft"
	ContestStatusRegistration = "registration"
	ContestStatusRunning      = "running"
	ContestStatusFrozen       = "frozen"
	ContestStatusEnded        = "ended"
)

type Contest struct {
	ID                            int64          `gorm:"column:id;primaryKey"`
	Title                         string         `gorm:"column:title"`
	Description                   string         `gorm:"column:description;type:text"`
	Mode                          string         `gorm:"column:mode"`
	StartTime                     time.Time      `gorm:"column:start_time"`
	EndTime                       time.Time      `gorm:"column:end_time"`
	FreezeTime                    *time.Time     `gorm:"column:freeze_time"`
	PausedSeconds                 int64          `gorm:"column:paused_seconds"`
	RuntimeRecoveryKey            string         `gorm:"column:runtime_recovery_key;size:191"`
	RuntimeRecoveryAppliedSeconds int64          `gorm:"column:runtime_recovery_applied_seconds"`
	Status                        string         `gorm:"column:status"`
	StatusVersion                 int64          `gorm:"column:status_version"`
	CreatedAt                     time.Time      `gorm:"column:created_at"`
	UpdatedAt                     time.Time      `gorm:"column:updated_at"`
	DeletedAt                     gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Contest) TableName() string {
	return "contests"
}

type AWDCheckerType string

const (
	AWDCheckerTypeLegacyProbe  AWDCheckerType = "legacy_probe"
	AWDCheckerTypeHTTPStandard AWDCheckerType = "http_standard"
	AWDCheckerTypeTCPStandard  AWDCheckerType = "tcp_standard"
	AWDCheckerTypeScript       AWDCheckerType = "script_checker"
)

type AWDCheckerValidationState string

const (
	AWDCheckerValidationStatePending AWDCheckerValidationState = "pending"
	AWDCheckerValidationStatePassed  AWDCheckerValidationState = "passed"
	AWDCheckerValidationStateFailed  AWDCheckerValidationState = "failed"
	AWDCheckerValidationStateStale   AWDCheckerValidationState = "stale"
)

type ContestChallenge struct {
	ID           int64          `gorm:"column:id;primaryKey"`
	ContestID    int64          `gorm:"column:contest_id"`
	ChallengeID  int64          `gorm:"column:challenge_id"`
	Points       int            `gorm:"column:points"`
	ContestScore *int           `gorm:"column:contest_score"`
	Order        int            `gorm:"column:order"`
	IsVisible    bool           `gorm:"column:is_visible"`
	FirstBloodBy *int64         `gorm:"column:first_blood_by"`
	CreatedAt    time.Time      `gorm:"column:created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (ContestChallenge) TableName() string {
	return "contest_challenges"
}

type ContestAWDService struct {
	ID                int64          `gorm:"column:id;primaryKey"`
	ContestID         int64          `gorm:"column:contest_id"`
	AWDChallengeID    int64          `gorm:"column:awd_challenge_id"`
	DisplayName       string         `gorm:"column:display_name"`
	ServiceSnapshot   string         `gorm:"column:service_snapshot;type:text;default:'{}'"`
	Order             int            `gorm:"column:order"`
	IsVisible         bool           `gorm:"column:is_visible"`
	ScoreConfig       string         `gorm:"column:score_config;type:text;default:'{}'"`
	RuntimeConfig     string         `gorm:"column:runtime_config;type:text;default:'{}'"`
	ValidationState   string         `gorm:"column:awd_checker_validation_state;size:24;not null;default:'pending'"`
	LastPreviewAt     *time.Time     `gorm:"column:awd_checker_last_preview_at"`
	LastPreviewResult string         `gorm:"column:awd_checker_last_preview_result;type:text;default:''"`
	CreatedAt         time.Time      `gorm:"column:created_at"`
	UpdatedAt         time.Time      `gorm:"column:updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (ContestAWDService) TableName() string {
	return "contest_awd_services"
}

const (
	ContestRegistrationStatusPending  = "pending"
	ContestRegistrationStatusApproved = "approved"
	ContestRegistrationStatusRejected = "rejected"
)

type ContestRegistration struct {
	ID         int64      `gorm:"column:id;primaryKey"`
	ContestID  int64      `gorm:"column:contest_id;not null;uniqueIndex:uk_contest_reg_user"`
	UserID     int64      `gorm:"column:user_id;not null;uniqueIndex:uk_contest_reg_user"`
	TeamID     *int64     `gorm:"column:team_id"`
	Status     string     `gorm:"column:status;size:16;not null;default:'pending'"`
	ReviewedBy *int64     `gorm:"column:reviewed_by"`
	ReviewedAt *time.Time `gorm:"column:reviewed_at"`
	CreatedAt  time.Time  `gorm:"column:created_at;not null"`
	UpdatedAt  time.Time  `gorm:"column:updated_at;not null"`
}

func (ContestRegistration) TableName() string {
	return "contest_registrations"
}

type Team struct {
	ID          int64          `gorm:"column:id;primaryKey"`
	ContestID   int64          `gorm:"column:contest_id;index"`
	Name        string         `gorm:"column:name"`
	CaptainID   int64          `gorm:"column:captain_id"`
	InviteCode  string         `gorm:"column:invite_code;uniqueIndex"`
	MaxMembers  int            `gorm:"column:max_members;default:4"`
	TotalScore  int            `gorm:"column:total_score;default:0"`
	LastSolveAt *time.Time     `gorm:"column:last_solve_at"`
	CreatedAt   time.Time      `gorm:"column:created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (Team) TableName() string {
	return "teams"
}

type TeamMember struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	ContestID int64     `gorm:"column:contest_id"`
	TeamID    int64     `gorm:"column:team_id;index:idx_team_user"`
	UserID    int64     `gorm:"column:user_id;index:idx_team_user"`
	JoinedAt  time.Time `gorm:"column:joined_at"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (TeamMember) TableName() string {
	return "team_members"
}

const (
	SubmissionReviewStatusNotRequired = "not_required"
	SubmissionReviewStatusPending     = "pending"
	SubmissionReviewStatusApproved    = "approved"
	SubmissionReviewStatusRejected    = "rejected"
)

type Submission struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	UserID        int64      `gorm:"column:user_id;not null;index:idx_user_challenge"`
	ChallengeID   int64      `gorm:"column:challenge_id;not null;index:idx_user_challenge"`
	ContestID     *int64     `gorm:"column:contest_id;index"`
	TeamID        *int64     `gorm:"column:team_id"`
	Flag          string     `gorm:"column:flag;size:500"`
	IsCorrect     bool       `gorm:"column:is_correct;not null"`
	ReviewStatus  string     `gorm:"column:review_status;size:32;default:'not_required';index:idx_submissions_review_status"`
	ReviewedBy    *int64     `gorm:"column:reviewed_by"`
	ReviewedAt    *time.Time `gorm:"column:reviewed_at"`
	ReviewComment string     `gorm:"column:review_comment"`
	Score         int        `gorm:"column:score;default:0"`
	SubmittedAt   time.Time  `gorm:"column:submitted_at;not null;index"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null"`
}

func (Submission) TableName() string {
	return "submissions"
}

type AWDServiceType string

const (
	AWDServiceTypeWebHTTP        AWDServiceType = "web_http"
	AWDServiceTypeBinaryTCP      AWDServiceType = "binary_tcp"
	AWDServiceTypeMultiContainer AWDServiceType = "multi_container"
)

type AWDDeploymentMode string

const (
	AWDDeploymentModeSingleContainer AWDDeploymentMode = "single_container"
	AWDDeploymentModeTopology        AWDDeploymentMode = "topology"
)

type ContestAWDServiceSnapshot struct {
	Name             string            `json:"name"`
	Category         string            `json:"category"`
	Difficulty       string            `json:"difficulty"`
	Description      string            `json:"description,omitempty"`
	ServiceType      AWDServiceType    `json:"service_type,omitempty"`
	DeploymentMode   AWDDeploymentMode `json:"deployment_mode,omitempty"`
	FlagMode         string            `json:"flag_mode,omitempty"`
	FlagConfig       map[string]any    `json:"flag_config,omitempty"`
	DefenseEntryMode string            `json:"defense_entry_mode,omitempty"`
	AccessConfig     map[string]any    `json:"access_config,omitempty"`
	RuntimeConfig    map[string]any    `json:"runtime_config,omitempty"`
}

func EncodeContestAWDServiceSnapshot(snapshot ContestAWDServiceSnapshot) (string, error) {
	raw, err := json.Marshal(snapshot)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func DecodeContestAWDServiceSnapshot(raw string) (ContestAWDServiceSnapshot, error) {
	if raw == "" {
		return ContestAWDServiceSnapshot{}, nil
	}
	var snapshot ContestAWDServiceSnapshot
	if err := json.Unmarshal([]byte(raw), &snapshot); err != nil {
		return ContestAWDServiceSnapshot{}, err
	}
	if snapshot.FlagConfig == nil {
		snapshot.FlagConfig = map[string]any{}
	}
	if snapshot.AccessConfig == nil {
		snapshot.AccessConfig = map[string]any{}
	}
	if snapshot.RuntimeConfig == nil {
		snapshot.RuntimeConfig = map[string]any{}
	}
	return snapshot, nil
}
