package jobs

import "ctf-platform/internal/model"

type awdServiceCheckOutcome struct {
	serviceStatus string
	checkResult   string
	checkerType   model.AWDCheckerType
	slaScore      int
	defenseScore  int
}

type awdServiceCheckResult struct {
	CheckedAt            string                 `json:"checked_at"`
	CheckSource          string                 `json:"check_source,omitempty"`
	CheckerType          model.AWDCheckerType   `json:"checker_type,omitempty"`
	HealthPath           string                 `json:"health_path"`
	InstanceCount        int                    `json:"instance_count"`
	HealthyInstanceCount int                    `json:"healthy_instance_count"`
	FailedInstanceCount  int                    `json:"failed_instance_count"`
	StatusReason         string                 `json:"status_reason,omitempty"`
	Probe                string                 `json:"probe,omitempty"`
	LatencyMS            int64                  `json:"latency_ms,omitempty"`
	ErrorCode            string                 `json:"error_code,omitempty"`
	Error                string                 `json:"error,omitempty"`
	Targets              []awdCheckTargetResult `json:"targets,omitempty"`
}

type awdCheckTargetResult struct {
	AccessURL string                  `json:"access_url,omitempty"`
	Healthy   bool                    `json:"healthy"`
	Probe     string                  `json:"probe,omitempty"`
	LatencyMS int64                   `json:"latency_ms,omitempty"`
	ErrorCode string                  `json:"error_code,omitempty"`
	Error     string                  `json:"error,omitempty"`
	Attempts  []awdProbeAttemptResult `json:"attempts,omitempty"`
}
