package model

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
	ChallengeID    int64          `gorm:"column:challenge_id;not null;uniqueIndex:uk_awd_team_services"`
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
	ID             int64     `gorm:"column:id;primaryKey"`
	RoundID        int64     `gorm:"column:round_id;not null;index"`
	AttackerTeamID int64     `gorm:"column:attacker_team_id;not null;index"`
	VictimTeamID   int64     `gorm:"column:victim_team_id;not null;index"`
	ChallengeID    int64     `gorm:"column:challenge_id;not null"`
	AttackType     string    `gorm:"column:attack_type;size:32;not null"`
	Source         string    `gorm:"column:source;size:32;not null;default:legacy"`
	SubmittedFlag  string    `gorm:"column:submitted_flag;size:512"`
	IsSuccess      bool      `gorm:"column:is_success;not null;default:false;index"`
	ScoreGained    int       `gorm:"column:score_gained;not null;default:0"`
	CreatedAt      time.Time `gorm:"column:created_at"`
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
	ChallengeID    int64     `gorm:"column:challenge_id;not null;index"`
	Method         string    `gorm:"column:method;size:16;not null"`
	Path           string    `gorm:"column:path;size:1024;not null"`
	StatusCode     int       `gorm:"column:status_code;not null"`
	Source         string    `gorm:"column:source;size:32;not null;default:runtime_proxy;index"`
	CreatedAt      time.Time `gorm:"column:created_at;index:idx_awd_traffic_round_created,priority:2,sort:desc"`
}

func (AWDTrafficEvent) TableName() string {
	return "awd_traffic_events"
}
