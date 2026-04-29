package ports

import (
	"context"
	"time"
)

type CheckerRunner interface {
	RunChecker(ctx context.Context, job CheckerRunJob) (CheckerRunResult, error)
}

type CheckerRunJob struct {
	Runtime         string
	Image           string
	Entry           string
	Args            []string
	Env             map[string]string
	Files           []CheckerRunFile
	OutputMode      string
	NetworkMode     string
	TargetAllowlist []string
	Timeout         time.Duration
	Limits          CheckerRunLimits
	Metadata        CheckerRunMetadata
}

type CheckerRunFile struct {
	Path    string
	Content []byte
	Mode    int
}

type CheckerRunLimits struct {
	CPUQuota         float64
	MemoryBytes      int64
	PidsLimit        int64
	NofileLimit      int64
	OutputLimitBytes int64
}

type CheckerRunMetadata struct {
	ContestID   int64
	ServiceID   int64
	TeamID      int64
	RoundNumber int
}

type CheckerRunResult struct {
	Status           CheckerRunStatus
	Reason           string
	ExitCode         int64
	Stdout           string
	Stderr           string
	Duration         time.Duration
	OutputLimitHit   bool
	ResourceLimitHit string
	StartedAt        time.Time
	FinishedAt       time.Time
}

type CheckerRunStatus string

const (
	CheckerRunStatusOK     CheckerRunStatus = "ok"
	CheckerRunStatusFailed CheckerRunStatus = "failed"
)

const (
	CheckerReasonPassed              = "checker_passed"
	CheckerReasonFailed              = "checker_failed"
	CheckerReasonTimeout             = "checker_timeout"
	CheckerReasonOutputLimitExceeded = "checker_output_limit_exceeded"
	CheckerReasonInvalidOutput       = "invalid_checker_output"
	CheckerReasonSandboxError        = "checker_sandbox_error"
)
