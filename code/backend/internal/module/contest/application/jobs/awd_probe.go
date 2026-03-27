package jobs

type awdProbeAttemptResult struct {
	Probe     string `json:"probe"`
	Healthy   bool   `json:"healthy"`
	LatencyMS int64  `json:"latency_ms,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Error     string `json:"error,omitempty"`
}

type awdInstanceProbeResult struct {
	healthy   bool
	latencyMS int64
	probe     string
	errorCode string
	err       string
	attempts  []awdProbeAttemptResult
}

type awdCheckError struct {
	code    string
	message string
}

func (e awdCheckError) Error() string {
	return e.message
}
