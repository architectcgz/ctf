package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	contestports "ctf-platform/internal/module/contest/ports"
)

const (
	checkerLabelContestID     = "ctf.checker.contest"
	checkerLabelServiceID     = "ctf.checker.service"
	checkerLabelTeamID        = "ctf.checker.team"
	checkerLabelRoundNumber   = "ctf.checker.round"
	checkerLabelRuntimeNodeID = "ctf.checker.runtime_node"
)

var (
	errCheckerSandboxExecutorUnavailable = errors.New("checker sandbox executor is unavailable")
	errCheckerRunnerUnavailable          = errors.New("checker runner is unavailable")
)

type SandboxCheckerRunner struct {
	executor runtimeports.SandboxExecutor
}

var _ contestports.CheckerRunner = (*SandboxCheckerRunner)(nil)

func NewSandboxCheckerRunner(executor runtimeports.SandboxExecutor) *SandboxCheckerRunner {
	return &SandboxCheckerRunner{executor: executor}
}

func (r *SandboxCheckerRunner) RunChecker(ctx context.Context, job contestports.CheckerRunJob) (contestports.CheckerRunResult, error) {
	if r == nil || r.executor == nil {
		return contestports.CheckerRunResult{}, errCheckerSandboxExecutorUnavailable
	}
	result, err := r.executor.RunSandboxExec(ctx, sandboxExecJobFromCheckerRunJob(job))
	if err != nil {
		return contestports.CheckerRunResult{}, err
	}
	return checkerRunResultFromSandboxExecResult(result), nil
}

type CheckerSandboxExecutor struct {
	runner contestports.CheckerRunner
}

var _ runtimeports.SandboxExecutor = (*CheckerSandboxExecutor)(nil)

func NewCheckerSandboxExecutor(runner contestports.CheckerRunner) *CheckerSandboxExecutor {
	return &CheckerSandboxExecutor{runner: runner}
}

func (e *CheckerSandboxExecutor) RunSandboxExec(ctx context.Context, job runtimeports.SandboxExecJob) (runtimeports.SandboxExecResult, error) {
	if e == nil || e.runner == nil {
		return runtimeports.SandboxExecResult{}, errCheckerRunnerUnavailable
	}
	result, err := e.runner.RunChecker(ctx, checkerRunJobFromSandboxExecJob(job))
	if err != nil {
		return runtimeports.SandboxExecResult{}, err
	}
	return sandboxExecResultFromCheckerRunResult(result), nil
}

func sandboxExecJobFromCheckerRunJob(job contestports.CheckerRunJob) runtimeports.SandboxExecJob {
	return runtimeports.SandboxExecJob{
		Runtime:         job.Runtime,
		Image:           job.Image,
		Entry:           job.Entry,
		Args:            append([]string(nil), job.Args...),
		Env:             cloneStringMap(job.Env),
		Files:           sandboxExecFilesFromCheckerRunFiles(job.Files),
		OutputMode:      job.OutputMode,
		NetworkMode:     job.NetworkMode,
		TargetAllowlist: append([]string(nil), job.TargetAllowlist...),
		Timeout:         job.Timeout,
		Limits: runtimeports.SandboxExecLimits{
			CPUQuota:         job.Limits.CPUQuota,
			MemoryBytes:      job.Limits.MemoryBytes,
			PidsLimit:        job.Limits.PidsLimit,
			NofileLimit:      job.Limits.NofileLimit,
			OutputLimitBytes: job.Limits.OutputLimitBytes,
		},
		Labels: checkerRunMetadataLabels(job.Metadata),
	}
}

func checkerRunJobFromSandboxExecJob(job runtimeports.SandboxExecJob) contestports.CheckerRunJob {
	return contestports.CheckerRunJob{
		Runtime:         job.Runtime,
		Image:           job.Image,
		Entry:           job.Entry,
		Args:            append([]string(nil), job.Args...),
		Env:             cloneStringMap(job.Env),
		Files:           checkerRunFilesFromSandboxExecFiles(job.Files),
		OutputMode:      job.OutputMode,
		NetworkMode:     job.NetworkMode,
		TargetAllowlist: append([]string(nil), job.TargetAllowlist...),
		Timeout:         job.Timeout,
		Limits: contestports.CheckerRunLimits{
			CPUQuota:         job.Limits.CPUQuota,
			MemoryBytes:      job.Limits.MemoryBytes,
			PidsLimit:        job.Limits.PidsLimit,
			NofileLimit:      job.Limits.NofileLimit,
			OutputLimitBytes: job.Limits.OutputLimitBytes,
		},
		Metadata: checkerRunMetadataFromLabels(job.Labels),
	}
}

func sandboxExecFilesFromCheckerRunFiles(files []contestports.CheckerRunFile) []runtimeports.SandboxExecFile {
	if len(files) == 0 {
		return nil
	}
	result := make([]runtimeports.SandboxExecFile, 0, len(files))
	for _, file := range files {
		result = append(result, runtimeports.SandboxExecFile{
			Path:    file.Path,
			Content: append([]byte(nil), file.Content...),
			Mode:    file.Mode,
		})
	}
	return result
}

func checkerRunFilesFromSandboxExecFiles(files []runtimeports.SandboxExecFile) []contestports.CheckerRunFile {
	if len(files) == 0 {
		return nil
	}
	result := make([]contestports.CheckerRunFile, 0, len(files))
	for _, file := range files {
		result = append(result, contestports.CheckerRunFile{
			Path:    file.Path,
			Content: append([]byte(nil), file.Content...),
			Mode:    file.Mode,
		})
	}
	return result
}

func checkerRunMetadataLabels(metadata contestports.CheckerRunMetadata) map[string]string {
	return map[string]string{
		checkerLabelContestID:     fmt.Sprintf("%d", metadata.ContestID),
		checkerLabelServiceID:     fmt.Sprintf("%d", metadata.ServiceID),
		checkerLabelTeamID:        fmt.Sprintf("%d", metadata.TeamID),
		checkerLabelRoundNumber:   fmt.Sprintf("%d", metadata.RoundNumber),
		checkerLabelRuntimeNodeID: fmt.Sprintf("%d", metadata.RuntimeNodeID),
	}
}

func checkerRunMetadataFromLabels(labels map[string]string) contestports.CheckerRunMetadata {
	return contestports.CheckerRunMetadata{
		ContestID:     parseInt64Label(labels, checkerLabelContestID),
		ServiceID:     parseInt64Label(labels, checkerLabelServiceID),
		TeamID:        parseInt64Label(labels, checkerLabelTeamID),
		RoundNumber:   int(parseInt64Label(labels, checkerLabelRoundNumber)),
		RuntimeNodeID: parseInt64Label(labels, checkerLabelRuntimeNodeID),
	}
}

func parseInt64Label(labels map[string]string, key string) int64 {
	value := strings.TrimSpace(labels[key])
	if value == "" {
		return 0
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}

func sandboxExecResultFromCheckerRunResult(result contestports.CheckerRunResult) runtimeports.SandboxExecResult {
	return runtimeports.SandboxExecResult{
		Status:           sandboxExecStatusFromCheckerRunStatus(result.Status),
		Reason:           sandboxExecReasonFromCheckerReason(result.Reason),
		ExitCode:         result.ExitCode,
		Stdout:           result.Stdout,
		Stderr:           result.Stderr,
		Duration:         result.Duration,
		OutputLimitHit:   result.OutputLimitHit,
		ResourceLimitHit: result.ResourceLimitHit,
		StartedAt:        result.StartedAt,
		FinishedAt:       result.FinishedAt,
	}
}

func checkerRunResultFromSandboxExecResult(result runtimeports.SandboxExecResult) contestports.CheckerRunResult {
	return contestports.CheckerRunResult{
		Status:           checkerRunStatusFromSandboxExecStatus(result.Status),
		Reason:           checkerReasonFromSandboxExecReason(result.Reason),
		ExitCode:         result.ExitCode,
		Stdout:           result.Stdout,
		Stderr:           result.Stderr,
		Duration:         result.Duration,
		OutputLimitHit:   result.OutputLimitHit,
		ResourceLimitHit: result.ResourceLimitHit,
		StartedAt:        result.StartedAt,
		FinishedAt:       result.FinishedAt,
	}
}

func sandboxExecStatusFromCheckerRunStatus(status contestports.CheckerRunStatus) runtimeports.SandboxExecStatus {
	if status == contestports.CheckerRunStatusOK {
		return runtimeports.SandboxExecStatusOK
	}
	return runtimeports.SandboxExecStatusFailed
}

func checkerRunStatusFromSandboxExecStatus(status runtimeports.SandboxExecStatus) contestports.CheckerRunStatus {
	if status == runtimeports.SandboxExecStatusOK {
		return contestports.CheckerRunStatusOK
	}
	return contestports.CheckerRunStatusFailed
}

func sandboxExecReasonFromCheckerReason(reason string) string {
	switch reason {
	case contestports.CheckerReasonPassed:
		return runtimeports.SandboxExecReasonPassed
	case contestports.CheckerReasonFailed:
		return runtimeports.SandboxExecReasonFailed
	case contestports.CheckerReasonTimeout:
		return runtimeports.SandboxExecReasonTimeout
	case contestports.CheckerReasonOutputLimitExceeded:
		return runtimeports.SandboxExecReasonOutputLimitExceeded
	case contestports.CheckerReasonInvalidOutput:
		return runtimeports.SandboxExecReasonInvalidOutput
	case contestports.CheckerReasonSandboxError:
		return runtimeports.SandboxExecReasonSandboxError
	default:
		return reason
	}
}

func checkerReasonFromSandboxExecReason(reason string) string {
	switch reason {
	case runtimeports.SandboxExecReasonPassed:
		return contestports.CheckerReasonPassed
	case runtimeports.SandboxExecReasonFailed:
		return contestports.CheckerReasonFailed
	case runtimeports.SandboxExecReasonTimeout:
		return contestports.CheckerReasonTimeout
	case runtimeports.SandboxExecReasonOutputLimitExceeded:
		return contestports.CheckerReasonOutputLimitExceeded
	case runtimeports.SandboxExecReasonInvalidOutput:
		return contestports.CheckerReasonInvalidOutput
	case runtimeports.SandboxExecReasonSandboxError:
		return contestports.CheckerReasonSandboxError
	default:
		return reason
	}
}

func cloneStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]string, len(values))
	for key, value := range values {
		result[key] = value
	}
	return result
}
