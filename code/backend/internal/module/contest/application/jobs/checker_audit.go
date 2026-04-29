package jobs

import (
	"strings"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

const (
	awdCheckerAuditOutputLimit = 2048
	awdCheckerRedactedValue    = "[redacted]"
)

type awdCheckerAuditRecord struct {
	ContestID       int64                `json:"contest_id,omitempty"`
	ServiceID       int64                `json:"service_id,omitempty"`
	TeamID          int64                `json:"team_id,omitempty"`
	RoundNumber     int                  `json:"round_number,omitempty"`
	CheckerType     model.AWDCheckerType `json:"checker_type,omitempty"`
	ArtifactDigest  string               `json:"artifact_digest,omitempty"`
	DurationMS      int64                `json:"duration_ms,omitempty"`
	ErrorCode       string               `json:"error_code,omitempty"`
	ExitCode        int64                `json:"exit_code,omitempty"`
	Stdout          string               `json:"stdout,omitempty"`
	Stderr          string               `json:"stderr,omitempty"`
	OutputTruncated bool                 `json:"output_truncated,omitempty"`
}

func buildAWDCheckerAuditRecord(
	job contestports.CheckerRunJob,
	checkerType model.AWDCheckerType,
	artifactDigest string,
	result contestports.CheckerRunResult,
	errorCode string,
	secrets ...string,
) *awdCheckerAuditRecord {
	stdout, stdoutTruncated := redactAndTruncateAWDCheckerText(result.Stdout, awdCheckerAuditOutputLimit, secrets...)
	stderr, stderrTruncated := redactAndTruncateAWDCheckerText(result.Stderr, awdCheckerAuditOutputLimit, secrets...)
	duration := result.Duration
	if duration <= 0 && !result.StartedAt.IsZero() && !result.FinishedAt.IsZero() {
		duration = result.FinishedAt.Sub(result.StartedAt)
	}
	return &awdCheckerAuditRecord{
		ContestID:       job.Metadata.ContestID,
		ServiceID:       job.Metadata.ServiceID,
		TeamID:          job.Metadata.TeamID,
		RoundNumber:     job.Metadata.RoundNumber,
		CheckerType:     checkerType,
		ArtifactDigest:  strings.TrimSpace(artifactDigest),
		DurationMS:      duration.Milliseconds(),
		ErrorCode:       strings.TrimSpace(errorCode),
		ExitCode:        result.ExitCode,
		Stdout:          stdout,
		Stderr:          stderr,
		OutputTruncated: stdoutTruncated || stderrTruncated || result.OutputLimitHit,
	}
}

func redactAndTruncateAWDCheckerText(value string, limit int, secrets ...string) (string, bool) {
	redacted := strings.TrimSpace(value)
	changed := false
	for _, secret := range secrets {
		secret = strings.TrimSpace(secret)
		if secret == "" {
			continue
		}
		if strings.Contains(redacted, secret) {
			changed = true
		}
		redacted = strings.ReplaceAll(redacted, secret, awdCheckerRedactedValue)
	}
	if limit <= 0 {
		return redacted, false
	}
	if len(redacted) <= limit {
		return redacted, false
	}
	suffix := "...[truncated]"
	if limit <= len(suffix) {
		return suffix[:limit], true
	}
	output := strings.TrimSpace(redacted[:limit])
	if changed && !strings.Contains(output, awdCheckerRedactedValue) && limit > len(awdCheckerRedactedValue)+1 {
		output = strings.TrimSpace(output[:limit-len(awdCheckerRedactedValue)-1]) + " " + awdCheckerRedactedValue
	}
	return output + suffix, true
}

func sanitizeAWDCheckerText(value string, secrets ...string) string {
	redacted, _ := redactAndTruncateAWDCheckerText(value, awdCheckerAuditOutputLimit, secrets...)
	if redacted == "" {
		return "unknown_checker_error"
	}
	return redacted
}

func awdCheckerDurationMS(value time.Duration) int64 {
	if value <= 0 {
		return 0
	}
	return value.Milliseconds()
}
