package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
	artifactRoot := t.TempDir()
	t.Setenv("AWD_CHECKER_ARTIFACT_DIR", artifactRoot)
	artifactContent := []byte("print('ok')\n")
	artifactPath := filepath.Join(artifactRoot, "script-checker", "check.py")
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0o700); err != nil {
		t.Fatalf("create artifact dir: %v", err)
	}
	if err := os.WriteFile(artifactPath, artifactContent, 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	artifactHash := sha256.Sum256(artifactContent)
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
			"env": {"CUSTOM_FLAG": "{{FLAG}}", "CHECKER_TOKEN": "{{CHECKER_TOKEN}}"},
			"output": "json",
			"artifact": {
				"entry": "docker/check/check.py",
				"storage_path": "` + artifactPath + `",
				"sha256": "` + hex.EncodeToString(artifactHash[:]) + `",
				"size": 12
			}
		}`,
		CheckerTokenEnv: "CHECKER_TOKEN",
		CheckerToken:    "preview-checker-token",
		AccessURL:       "http://10.10.0.23:8080",
		PreviewFlag:     "flag{preview}",
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
	if job.Env["TARGET_URL"] != "http://10.10.0.23:8080" || job.Env["CUSTOM_FLAG"] != "flag{preview}" || job.Env["CHECKER_TOKEN"] != "preview-checker-token" {
		t.Fatalf("Env = %#v", job.Env)
	}
	if len(job.TargetAllowlist) != 1 || job.TargetAllowlist[0] != "10.10.0.23:8080" {
		t.Fatalf("TargetAllowlist = %#v", job.TargetAllowlist)
	}
	if len(job.Files) != 1 || job.Files[0].Path != "docker/check/check.py" || string(job.Files[0].Content) != "print('ok')\n" {
		t.Fatalf("Files = %#v", job.Files)
	}
}

func TestLoadAWDScriptCheckerArtifactRejectsOutsideRoot(t *testing.T) {
	artifactRoot := t.TempDir()
	t.Setenv("AWD_CHECKER_ARTIFACT_DIR", artifactRoot)
	outsidePath := filepath.Join(t.TempDir(), "check.py")
	if err := os.WriteFile(outsidePath, []byte("print('ok')\n"), 0o600); err != nil {
		t.Fatalf("write outside artifact: %v", err)
	}

	_, ok, err := loadAWDScriptCheckerArtifacts(awdScriptCheckerConfig{
		Entry: "docker/check/check.py",
		Artifact: awdScriptCheckerArtifactConfig{
			Entry:       "docker/check/check.py",
			StoragePath: outsidePath,
		},
	})
	if err == nil {
		t.Fatalf("expected outside artifact root error")
	}
	if ok {
		t.Fatalf("ok = true, want false")
	}
}

func TestLoadAWDScriptCheckerArtifactLoadsMultipleFiles(t *testing.T) {
	artifactRoot := t.TempDir()
	t.Setenv("AWD_CHECKER_ARTIFACT_DIR", artifactRoot)
	checkContent := []byte("import protocol\n")
	protocolContent := []byte("def ok(): return True\n")
	checkPath := filepath.Join(artifactRoot, "script-checker", "check.py")
	protocolPath := filepath.Join(artifactRoot, "script-checker", "protocol.py")
	if err := os.MkdirAll(filepath.Dir(checkPath), 0o700); err != nil {
		t.Fatalf("create artifact dir: %v", err)
	}
	if err := os.WriteFile(checkPath, checkContent, 0o600); err != nil {
		t.Fatalf("write check artifact: %v", err)
	}
	if err := os.WriteFile(protocolPath, protocolContent, 0o600); err != nil {
		t.Fatalf("write protocol artifact: %v", err)
	}
	checkHash := sha256.Sum256(checkContent)
	protocolHash := sha256.Sum256(protocolContent)

	files, ok, err := loadAWDScriptCheckerArtifacts(awdScriptCheckerConfig{
		Entry: "docker/check/check.py",
		Artifact: awdScriptCheckerArtifactConfig{
			Entry: "docker/check/check.py",
			Files: []awdScriptCheckerArtifactFileConfig{
				{
					Path:        "docker/check/check.py",
					StoragePath: checkPath,
					SHA256:      hex.EncodeToString(checkHash[:]),
					Size:        int64(len(checkContent)),
				},
				{
					Path:        "docker/check/protocol.py",
					StoragePath: protocolPath,
					SHA256:      hex.EncodeToString(protocolHash[:]),
					Size:        int64(len(protocolContent)),
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("loadAWDScriptCheckerArtifacts() error = %v", err)
	}
	if !ok {
		t.Fatal("ok = false, want true")
	}
	if len(files) != 2 {
		t.Fatalf("files = %#v", files)
	}
	if files[0].Path != "docker/check/check.py" || string(files[0].Content) != string(checkContent) {
		t.Fatalf("unexpected first file: %#v", files[0])
	}
	if files[1].Path != "docker/check/protocol.py" || string(files[1].Content) != string(protocolContent) {
		t.Fatalf("unexpected second file: %#v", files[1])
	}
}

func TestAWDRoundUpdaterScriptCheckerRedactsFlagInFailedAudit(t *testing.T) {
	artifactRoot := t.TempDir()
	t.Setenv("AWD_CHECKER_ARTIFACT_DIR", artifactRoot)
	artifactContent := []byte("print('fail')\n")
	artifactPath := filepath.Join(artifactRoot, "script-checker", "check.py")
	if err := os.MkdirAll(filepath.Dir(artifactPath), 0o700); err != nil {
		t.Fatalf("create artifact dir: %v", err)
	}
	if err := os.WriteFile(artifactPath, artifactContent, 0o600); err != nil {
		t.Fatalf("write artifact: %v", err)
	}
	artifactHash := sha256.Sum256(artifactContent)
	runner := &fakeCheckerRunner{
		result: contestports.CheckerRunResult{
			Status:   contestports.CheckerRunStatusFailed,
			Reason:   contestports.CheckerReasonFailed,
			ExitCode: 1,
			Stdout:   "stdout leaked flag{preview}",
			Stderr:   "stderr leaked flag{preview}",
			Duration: 23 * time.Millisecond,
		},
	}
	updater := NewAWDRoundUpdater(nil, nil, config.ContestAWDConfig{
		CheckerTimeout: time.Second,
		CheckerSandbox: config.CheckerSandboxConfig{
			Timeout:          10 * time.Second,
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
			"output": "json",
			"artifact": {
				"entry": "docker/check/check.py",
				"storage_path": "` + artifactPath + `",
				"sha256": "` + hex.EncodeToString(artifactHash[:]) + `",
				"size": 14,
				"digest": "artifact-digest-1"
			}
		}`,
		AccessURL:   "http://10.10.0.23:8080",
		PreviewFlag: "flag{preview}",
	})
	if err != nil {
		t.Fatalf("PreviewServiceCheck() error = %v", err)
	}
	if strings.Contains(resp.CheckResult, "flag{preview}") {
		t.Fatalf("CheckResult leaked flag: %s", resp.CheckResult)
	}
	var result map[string]any
	if err := json.Unmarshal([]byte(resp.CheckResult), &result); err != nil {
		t.Fatalf("unmarshal check result: %v", err)
	}
	targets, ok := result["targets"].([]any)
	if !ok || len(targets) != 1 {
		t.Fatalf("unexpected targets: %#v", result["targets"])
	}
	target, ok := targets[0].(map[string]any)
	if !ok {
		t.Fatalf("unexpected target: %#v", targets[0])
	}
	audit, ok := target["audit"].(map[string]any)
	if !ok {
		t.Fatalf("missing audit: %#v", target)
	}
	if audit["checker_type"] != string(model.AWDCheckerTypeScript) || audit["service_id"] != float64(2001) || audit["artifact_digest"] != "artifact-digest-1" {
		t.Fatalf("unexpected audit: %#v", audit)
	}
	if !strings.Contains(fmt.Sprint(audit["stderr"]), "[redacted]") {
		t.Fatalf("audit stderr was not redacted: %#v", audit)
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
