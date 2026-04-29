package jobs

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type fakeCheckerRunner struct {
	jobs   []contestports.CheckerRunJob
	result contestports.CheckerRunResult
	err    error
}

func (f *fakeCheckerRunner) RunChecker(_ context.Context, job contestports.CheckerRunJob) (contestports.CheckerRunResult, error) {
	f.jobs = append(f.jobs, job)
	if f.err != nil {
		return contestports.CheckerRunResult{}, f.err
	}
	return f.result, nil
}

func TestAWDRoundUpdaterPreviewScriptCheckerUsesSandboxRunner(t *testing.T) {
	t.Parallel()

	runner := &fakeCheckerRunner{
		result: contestports.CheckerRunResult{
			Status:   contestports.CheckerRunStatusOK,
			Reason:   contestports.CheckerReasonPassed,
			ExitCode: 0,
			Stdout:   `{"status":"ok","reason":"flag_roundtrip_passed"}`,
		},
	}
	updater := NewAWDRoundUpdater(nil, nil, config.ContestAWDConfig{
		CheckerTimeout: time.Second,
		CheckerSandbox: config.CheckerSandboxConfig{
			Timeout:          10 * time.Second,
			CPUQuota:         0.5,
			MemoryBytes:      128 * 1024 * 1024,
			PidsLimit:        64,
			NofileLimit:      128,
			OutputLimitBytes: 32768,
		},
	}, "", nil, nil)
	updater.SetCheckerRunner(runner)

	resp, err := updater.PreviewServiceCheck(context.Background(), contestports.AWDServicePreviewRequest{
		ServiceID:      2001,
		AWDChallengeID: 3001,
		CheckerType:    model.AWDCheckerTypeScript,
		CheckerConfig: `{
			"runtime": "python3",
			"entry": "docker/check/check.py",
			"timeout_sec": 7,
			"args": ["{{TARGET_URL}}", "{{FLAG}}"],
			"env": {"CUSTOM_FLAG": "{{FLAG}}"},
			"output": "json"
		}`,
		AccessURL:   "http://10.10.0.23:8080",
		PreviewFlag: "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewServiceCheck() error = %v", err)
	}
	if resp.ServiceStatus != model.AWDServiceStatusUp {
		t.Fatalf("ServiceStatus = %s, want up", resp.ServiceStatus)
	}
	if len(runner.jobs) != 1 {
		t.Fatalf("runner jobs = %d, want 1", len(runner.jobs))
	}
	job := runner.jobs[0]
	if job.Runtime != "python3" || job.Entry != "docker/check/check.py" || job.OutputMode != "json" {
		t.Fatalf("unexpected runner job contract: %+v", job)
	}
	if job.Timeout != 7*time.Second {
		t.Fatalf("Timeout = %s, want 7s", job.Timeout)
	}
	if len(job.Args) != 2 || job.Args[0] != "http://10.10.0.23:8080" || job.Args[1] != "flag{preview}" {
		t.Fatalf("Args = %#v", job.Args)
	}
	if job.Env["TARGET_URL"] != "http://10.10.0.23:8080" || job.Env["CUSTOM_FLAG"] != "flag{preview}" {
		t.Fatalf("Env = %#v", job.Env)
	}
	if len(job.TargetAllowlist) != 1 || job.TargetAllowlist[0] != "10.10.0.23:8080" {
		t.Fatalf("TargetAllowlist = %#v", job.TargetAllowlist)
	}
}

func TestAWDRoundUpdaterPreviewScriptCheckerRejectsMissingRunner(t *testing.T) {
	t.Parallel()

	updater := NewAWDRoundUpdater(nil, nil, config.ContestAWDConfig{CheckerTimeout: time.Second}, "", nil, nil)
	resp, err := updater.PreviewServiceCheck(context.Background(), contestports.AWDServicePreviewRequest{
		AWDChallengeID: 3001,
		CheckerType:    model.AWDCheckerTypeScript,
		CheckerConfig:  `{"runtime":"python3","entry":"docker/check/check.py"}`,
		AccessURL:      "http://10.10.0.23:8080",
		PreviewFlag:    "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewServiceCheck() error = %v", err)
	}
	if resp.ServiceStatus != model.AWDServiceStatusDown {
		t.Fatalf("ServiceStatus = %s, want down", resp.ServiceStatus)
	}
}
