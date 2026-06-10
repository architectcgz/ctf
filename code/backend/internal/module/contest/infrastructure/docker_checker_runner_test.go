package infrastructure

import (
	"context"
	"testing"
	"time"

	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	contestports "ctf-platform/internal/module/contest/ports"
)

func TestSandboxCheckerRunnerMapsCheckerJobToSandboxExecutor(t *testing.T) {
	t.Parallel()

	executor := &fakeSandboxExecutor{
		result: runtimeports.SandboxExecResult{
			Status:   runtimeports.SandboxExecStatusOK,
			Reason:   runtimeports.SandboxExecReasonPassed,
			Stdout:   "checker-ok",
			Duration: 250 * time.Millisecond,
		},
	}
	runner := NewSandboxCheckerRunner(executor)

	result, err := runner.RunChecker(context.Background(), contestports.CheckerRunJob{
		Runtime:         "python3",
		Image:           "python:3.12-alpine",
		Entry:           "check.py",
		Args:            []string{"--target", "http://10.0.0.2"},
		Env:             map[string]string{"FLAG": "ctf{example}"},
		Files:           []contestports.CheckerRunFile{{Path: "check.py", Content: []byte("print(1)"), Mode: 0o500}},
		OutputMode:      "json",
		NetworkMode:     "ctf-awd-target",
		TargetAllowlist: []string{"10.0.0.2:80"},
		Timeout:         3 * time.Second,
		Limits:          contestports.CheckerRunLimits{CPUQuota: 0.25, MemoryBytes: 64},
		Metadata: contestports.CheckerRunMetadata{
			ContestID:   7,
			ServiceID:   9,
			TeamID:      11,
			RoundNumber: 3,
			NodeID:      13,
		},
	})
	if err != nil {
		t.Fatalf("RunChecker() error = %v", err)
	}
	if result.Status != contestports.CheckerRunStatusOK || result.Reason != contestports.CheckerReasonPassed || result.Stdout != "checker-ok" {
		t.Fatalf("unexpected checker result: %+v", result)
	}
	if executor.lastJob.Labels[checkerLabelContestID] != "7" || executor.lastJob.Labels[checkerLabelTeamID] != "11" || executor.lastJob.Labels[checkerLabelNodeID] != "13" {
		t.Fatalf("sandbox labels = %+v", executor.lastJob.Labels)
	}
	if len(executor.lastJob.Files) != 1 || string(executor.lastJob.Files[0].Content) != "print(1)" {
		t.Fatalf("sandbox files = %+v", executor.lastJob.Files)
	}
}

func TestSandboxCheckerRunnerMapsNeutralReasonsToCheckerReasons(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		sandbox    string
		wantReason string
	}{
		{name: "passed", sandbox: runtimeports.SandboxExecReasonPassed, wantReason: contestports.CheckerReasonPassed},
		{name: "failed", sandbox: runtimeports.SandboxExecReasonFailed, wantReason: contestports.CheckerReasonFailed},
		{name: "timeout", sandbox: runtimeports.SandboxExecReasonTimeout, wantReason: contestports.CheckerReasonTimeout},
		{name: "output limit", sandbox: runtimeports.SandboxExecReasonOutputLimitExceeded, wantReason: contestports.CheckerReasonOutputLimitExceeded},
		{name: "invalid output", sandbox: runtimeports.SandboxExecReasonInvalidOutput, wantReason: contestports.CheckerReasonInvalidOutput},
		{name: "sandbox error", sandbox: runtimeports.SandboxExecReasonSandboxError, wantReason: contestports.CheckerReasonSandboxError},
		{name: "custom passthrough", sandbox: "flag_roundtrip_failed", wantReason: "flag_roundtrip_failed"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runner := NewSandboxCheckerRunner(&fakeSandboxExecutor{
				result: runtimeports.SandboxExecResult{
					Status: runtimeports.SandboxExecStatusFailed,
					Reason: tc.sandbox,
				},
			})
			result, err := runner.RunChecker(context.Background(), contestports.CheckerRunJob{})
			if err != nil {
				t.Fatalf("RunChecker() error = %v", err)
			}
			if result.Reason != tc.wantReason {
				t.Fatalf("Reason = %q, want %q", result.Reason, tc.wantReason)
			}
		})
	}
}

type fakeSandboxExecutor struct {
	lastJob runtimeports.SandboxExecJob
	result  runtimeports.SandboxExecResult
}

func (f *fakeSandboxExecutor) RunSandboxExec(_ context.Context, job runtimeports.SandboxExecJob) (runtimeports.SandboxExecResult, error) {
	f.lastJob = job
	return f.result, nil
}
