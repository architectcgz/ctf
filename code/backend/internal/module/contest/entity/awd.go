package entity

import "time"

const (
	AWDRoundStatusPending  = "pending"
	AWDRoundStatusRunning  = "running"
	AWDRoundStatusFinished = "finished"

	AWDServiceStatusUp          = "up"
	AWDServiceStatusDown        = "down"
	AWDServiceStatusCompromised = "compromised"

	AWDAttackTypeFlagCapture    = "flag_capture"
	AWDAttackTypeServiceExploit = "service_exploit"

	AWDAttackSourceLegacy     = "legacy"
	AWDAttackSourceManual     = "manual_attack_log"
	AWDAttackSourceSubmission = "submission"

	AWDTrafficSourceRuntimeProxy = "runtime_proxy"

	AWDDefenseWorkspaceStatusPending      = "pending"
	AWDDefenseWorkspaceStatusProvisioning = "provisioning"
	AWDDefenseWorkspaceStatusRunning      = "running"
	AWDDefenseWorkspaceStatusFailed       = "failed"

	AWDServiceOperationTypeStart    = "start"
	AWDServiceOperationTypeRestart  = "restart"
	AWDServiceOperationTypeRecover  = "recover"
	AWDServiceOperationTypeRecreate = "recreate"

	AWDServiceOperationRequestedByUser   = "user"
	AWDServiceOperationRequestedByAdmin  = "admin"
	AWDServiceOperationRequestedBySystem = "system"

	AWDServiceOperationStatusRequested    = "requested"
	AWDServiceOperationStatusProvisioning = "provisioning"
	AWDServiceOperationStatusRecovering   = "recovering"
	AWDServiceOperationStatusRecovered    = "recovered"
	AWDServiceOperationStatusSucceeded    = "succeeded"
	AWDServiceOperationStatusFailed       = "failed"

	AWDScopeControlScopeTeam        = "team"
	AWDScopeControlScopeTeamService = "team_service"

	AWDScopeControlTypeRetired                    = "retired"
	AWDScopeControlTypeServiceDisabled            = "service_disabled"
	AWDScopeControlTypeDesiredReconcileSuppressed = "desired_reconcile_suppressed"
)

type AWDRound struct {
	ID           int64      `gorm:"column:id;primaryKey"`
	ContestID    int64      `gorm:"column:contest_id;index;uniqueIndex:uk_awd_rounds,priority:1;not null"`
	RoundNumber  int        `gorm:"column:round_number;not null;uniqueIndex:uk_awd_rounds,priority:2"`
	Status       string     `gorm:"column:status;size:16;not null;default:pending;index"`
	StartedAt    *time.Time `gorm:"column:started_at"`
	EndedAt      *time.Time `gorm:"column:ended_at"`
	AttackScore  int        `gorm:"column:attack_score;not null;default:50"`
	DefenseScore int        `gorm:"column:defense_score;not null;default:50"`
	CreatedAt    time.Time  `gorm:"column:created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at"`
}

func (AWDRound) TableName() string {
	return "awd_rounds"
}

type AWDTeamService struct {
	ID             int64          `gorm:"column:id;primaryKey"`
	RoundID        int64          `gorm:"column:round_id;not null;uniqueIndex:uk_awd_team_services"`
	TeamID         int64          `gorm:"column:team_id;not null;index:idx_awd_ts_team;uniqueIndex:uk_awd_team_services"`
	ServiceID      int64          `gorm:"column:service_id;not null;index:idx_awd_ts_round_team_service;uniqueIndex:uk_awd_team_services"`
	AWDChallengeID int64          `gorm:"column:awd_challenge_id;not null;index"`
	ServiceStatus  string         `gorm:"column:service_status;size:16;not null;default:up"`
	CheckResult    string         `gorm:"column:check_result;type:text;not null;default:'{}'"`
	CheckerType    AWDCheckerType `gorm:"column:checker_type;size:32;not null;default:''"`
	AttackReceived int            `gorm:"column:attack_received;not null;default:0"`
	SLAScore       int            `gorm:"column:sla_score;not null;default:0"`
	DefenseScore   int            `gorm:"column:defense_score;not null;default:0"`
	AttackScore    int            `gorm:"column:attack_score;not null;default:0"`
	CreatedAt      time.Time      `gorm:"column:created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at"`
}

func (AWDTeamService) TableName() string {
	return "awd_team_services"
}

type AWDAttackLog struct {
	ID                int64     `gorm:"column:id;primaryKey"`
	RoundID           int64     `gorm:"column:round_id;not null;index"`
	AttackerTeamID    int64     `gorm:"column:attacker_team_id;not null;index"`
	VictimTeamID      int64     `gorm:"column:victim_team_id;not null;index"`
	ServiceID         int64     `gorm:"column:service_id;not null;index:idx_awd_attack_round_service_success,priority:4"`
	AWDChallengeID    int64     `gorm:"column:awd_challenge_id;not null"`
	AttackType        string    `gorm:"column:attack_type;size:32;not null"`
	Source            string    `gorm:"column:source;size:32;not null;default:legacy"`
	SubmittedFlag     string    `gorm:"column:submitted_flag;size:512"`
	SubmittedByUserID *int64    `gorm:"column:submitted_by_user_id"`
	IsSuccess         bool      `gorm:"column:is_success;not null;default:false;index;index:idx_awd_attack_round_service_success,priority:5"`
	ScoreGained       int       `gorm:"column:score_gained;not null;default:0"`
	CreatedAt         time.Time `gorm:"column:created_at"`
}

func (AWDAttackLog) TableName() string {
	return "awd_attack_logs"
}

type AWDTrafficEvent struct {
	ID             int64     `gorm:"column:id;primaryKey"`
	ContestID      int64     `gorm:"column:contest_id;not null;index"`
	RoundID        int64     `gorm:"column:round_id;not null;index:idx_awd_traffic_round_created,priority:1;index:idx_awd_traffic_attacker,priority:1;index:idx_awd_traffic_victim,priority:1"`
	AttackerTeamID int64     `gorm:"column:attacker_team_id;not null;index:idx_awd_traffic_attacker,priority:2"`
	VictimTeamID   int64     `gorm:"column:victim_team_id;not null;index:idx_awd_traffic_victim,priority:2"`
	ServiceID      int64     `gorm:"column:service_id;not null;index"`
	AWDChallengeID int64     `gorm:"column:awd_challenge_id;not null;index"`
	Method         string    `gorm:"column:method;size:16;not null"`
	Path           string    `gorm:"column:path;size:1024;not null"`
	StatusCode     int       `gorm:"column:status_code;not null"`
	Source         string    `gorm:"column:source;size:32;not null;default:runtime_proxy;index"`
	CreatedAt      time.Time `gorm:"column:created_at;index:idx_awd_traffic_round_created,priority:2,sort:desc"`
}

func (AWDTrafficEvent) TableName() string {
	return "awd_traffic_events"
}

type AWDDefenseWorkspace struct {
	ID                int64     `gorm:"column:id;primaryKey"`
	ContestID         int64     `gorm:"column:contest_id;not null;uniqueIndex:uk_awd_defense_workspaces_scope,priority:1"`
	TeamID            int64     `gorm:"column:team_id;not null;uniqueIndex:uk_awd_defense_workspaces_scope,priority:2"`
	ServiceID         int64     `gorm:"column:service_id;not null;uniqueIndex:uk_awd_defense_workspaces_scope,priority:3"`
	InstanceID        int64     `gorm:"column:instance_id;not null;index:idx_awd_defense_workspaces_instance_id"`
	WorkspaceRevision int64     `gorm:"column:workspace_revision;not null;default:1"`
	Status            string    `gorm:"column:status;size:24;not null;default:'pending'"`
	ContainerID       string    `gorm:"column:container_id;size:64;not null;default:''"`
	SeedSignature     string    `gorm:"column:seed_signature;type:text;not null;default:''"`
	CreatedAt         time.Time `gorm:"column:created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at"`
}

func (AWDDefenseWorkspace) TableName() string {
	return "awd_defense_workspaces"
}

type AWDServiceOperation struct {
	ID            int64      `gorm:"column:id;primaryKey"`
	ContestID     int64      `gorm:"column:contest_id;not null;index:idx_awd_service_operations_scope,priority:1"`
	TeamID        int64      `gorm:"column:team_id;not null;index:idx_awd_service_operations_scope,priority:2"`
	ServiceID     int64      `gorm:"column:service_id;not null;index:idx_awd_service_operations_scope,priority:3"`
	InstanceID    int64      `gorm:"column:instance_id;not null;index"`
	OperationType string     `gorm:"column:operation_type;size:24;not null"`
	RequestedBy   string     `gorm:"column:requested_by;size:16;not null"`
	RequestedByID *int64     `gorm:"column:requested_by_id"`
	Reason        string     `gorm:"column:reason;size:128;not null;default:''"`
	SLABillable   bool       `gorm:"column:sla_billable;not null"`
	Status        string     `gorm:"column:status;size:24;not null"`
	ErrorMessage  string     `gorm:"column:error_message;type:text;not null;default:''"`
	StartedAt     time.Time  `gorm:"column:started_at;not null;index:idx_awd_service_operations_window,priority:1"`
	FinishedAt    *time.Time `gorm:"column:finished_at;index:idx_awd_service_operations_window,priority:2"`
	CreatedAt     time.Time  `gorm:"column:created_at"`
	UpdatedAt     time.Time  `gorm:"column:updated_at"`
}

func (AWDServiceOperation) TableName() string {
	return "awd_service_operations"
}

type AWDScopeControl struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	ContestID   int64     `gorm:"column:contest_id;not null;index:idx_awd_scope_controls_scope,priority:1;uniqueIndex:uk_awd_scope_controls"`
	TeamID      int64     `gorm:"column:team_id;not null;index:idx_awd_scope_controls_scope,priority:2;uniqueIndex:uk_awd_scope_controls"`
	ScopeType   string    `gorm:"column:scope_type;size:24;not null;index:idx_awd_scope_controls_scope,priority:3;uniqueIndex:uk_awd_scope_controls"`
	ServiceID   int64     `gorm:"column:service_id;not null;default:0;index:idx_awd_scope_controls_scope,priority:4;uniqueIndex:uk_awd_scope_controls"`
	ControlType string    `gorm:"column:control_type;size:48;not null;uniqueIndex:uk_awd_scope_controls"`
	Reason      string    `gorm:"column:reason;type:text;not null;default:''"`
	UpdatedBy   *int64    `gorm:"column:updated_by"`
	CreatedAt   time.Time `gorm:"column:created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at"`
}

func (AWDScopeControl) TableName() string {
	return "awd_scope_controls"
}

type AWDProxyTrafficEventInput struct {
	ContestID      int64
	AttackerTeamID int64
	VictimTeamID   int64
	ServiceID      int64
	AWDChallengeID int64
	Method         string
	Path           string
	StatusCode     int
}
