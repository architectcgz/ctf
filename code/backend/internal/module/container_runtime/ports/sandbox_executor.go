package ports

import (
	"context"
	"time"
)

type SandboxExecutor interface {
	RunSandboxExec(ctx context.Context, job SandboxExecJob) (SandboxExecResult, error)
}

type SandboxExecJob struct {
	Runtime         string
	Image           string
	Entry           string
	Args            []string
	Env             map[string]string
	Files           []SandboxExecFile
	OutputMode      string
	NetworkMode     string
	TargetAllowlist []string
	Timeout         time.Duration
	Limits          SandboxExecLimits
	Labels          map[string]string
}

type SandboxExecFile struct {
	Path    string
	Content []byte
	Mode    int
}

type SandboxExecLimits struct {
	CPUQuota         float64
	MemoryBytes      int64
	PidsLimit        int64
	NofileLimit      int64
	OutputLimitBytes int64
}

type SandboxExecResult struct {
	Status           SandboxExecStatus
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

type SandboxExecStatus string

const (
	SandboxExecStatusOK     SandboxExecStatus = "ok"
	SandboxExecStatusFailed SandboxExecStatus = "failed"
)
