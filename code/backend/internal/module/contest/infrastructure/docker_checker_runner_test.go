package infrastructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/config"
	contestports "ctf-platform/internal/module/contest/ports"
)

func TestDockerCheckerRunnerBuildsLockedDownContainerSpec(t *testing.T) {
	t.Parallel()

	runner := NewDockerCheckerRunnerWithClient(nil, checkerSandboxConfigForTest())
	spec, err := runner.buildContainerSpec(contestports.CheckerRunJob{
		Runtime: "python3",
		Entry:   "docker/check/check.py",
		Args:    []string{"{{TARGET_URL}}"},
		Env: map[string]string{
			"FLAG": "flag{redacted}",
		},
		Metadata: contestports.CheckerRunMetadata{
			ContestID:   10,
			ServiceID:   20,
			TeamID:      30,
			RoundNumber: 4,
		},
	}, "/tmp/checker-work")
	if err != nil {
		t.Fatalf("buildContainerSpec() error = %v", err)
	}

	if spec.HostConfig.Privileged {
		t.Fatal("checker sandbox must not be privileged")
	}
	if !spec.HostConfig.ReadonlyRootfs {
		t.Fatal("checker sandbox must use readonly rootfs")
	}
	if !spec.ContainerConfig.NetworkDisabled {
		t.Fatal("checker sandbox must disable network when no target network is explicit")
	}
	if got := string(spec.HostConfig.NetworkMode); got != "none" {
		t.Fatalf("NetworkMode = %q, want none", got)
	}
	if len(spec.HostConfig.CapDrop) != 1 || spec.HostConfig.CapDrop[0] != "ALL" {
		t.Fatalf("CapDrop = %v, want [ALL]", spec.HostConfig.CapDrop)
	}
	if !containsString(spec.HostConfig.SecurityOpt, "no-new-privileges:true") {
		t.Fatalf("SecurityOpt = %v, want no-new-privileges:true", spec.HostConfig.SecurityOpt)
	}
	if spec.HostConfig.Resources.Memory != 128*1024*1024 {
		t.Fatalf("Memory = %d, want 128MiB", spec.HostConfig.Resources.Memory)
	}
	if spec.HostConfig.Resources.PidsLimit == nil || *spec.HostConfig.Resources.PidsLimit != 64 {
		t.Fatalf("PidsLimit = %v, want 64", spec.HostConfig.Resources.PidsLimit)
	}
	if len(spec.HostConfig.Resources.Ulimits) != 1 || spec.HostConfig.Resources.Ulimits[0].Name != "nofile" || spec.HostConfig.Resources.Ulimits[0].Soft != 128 {
		t.Fatalf("Ulimits = %+v, want nofile=128", spec.HostConfig.Resources.Ulimits)
	}
	if len(spec.HostConfig.Mounts) != 1 {
		t.Fatalf("Mounts = %+v, want exactly one checker mount", spec.HostConfig.Mounts)
	}
	mount := spec.HostConfig.Mounts[0]
	if !mount.ReadOnly || mount.Target != "/checker" || mount.Source != "/tmp/checker-work" {
		t.Fatalf("checker mount = %+v, want read-only /tmp/checker-work:/checker", mount)
	}
	if strings.Contains(mount.Source, "docker.sock") || strings.Contains(mount.Target, "docker.sock") {
		t.Fatalf("checker sandbox must not mount docker socket: %+v", mount)
	}
	if spec.ContainerConfig.User != "65532:65532" {
		t.Fatalf("User = %q, want 65532:65532", spec.ContainerConfig.User)
	}
	if got := strings.Join(spec.ContainerConfig.Cmd, " "); got != "python3 /checker/docker/check/check.py {{TARGET_URL}}" {
		t.Fatalf("command = %q, want python3 /checker/docker/check/check.py {{TARGET_URL}}", got)
	}
	if spec.ContainerConfig.Labels["ctf.project"] != "ctf" {
		t.Fatalf("missing ctf project label: %+v", spec.ContainerConfig.Labels)
	}
	if spec.ContainerConfig.Labels["managed-by"] != "ctf-platform" {
		t.Fatalf("missing managed-by label: %+v", spec.ContainerConfig.Labels)
	}
	if spec.ContainerConfig.Labels["ctf.role"] != "checker-sandbox" {
		t.Fatalf("missing checker-sandbox label: %+v", spec.ContainerConfig.Labels)
	}
}

func TestDockerCheckerRunnerEnablesOnlyExplicitTargetNetwork(t *testing.T) {
	t.Parallel()

	runner := NewDockerCheckerRunnerWithClient(nil, checkerSandboxConfigForTest())
	spec, err := runner.buildContainerSpec(contestports.CheckerRunJob{
		Entry:           "check.py",
		NetworkMode:     "ctf-awd-target-10",
		TargetAllowlist: []string{"10.10.0.23:8080"},
	}, "/tmp/checker-work")
	if err != nil {
		t.Fatalf("buildContainerSpec() error = %v", err)
	}

	if spec.ContainerConfig.NetworkDisabled {
		t.Fatal("expected network to be enabled for explicit target network")
	}
	if got := string(spec.HostConfig.NetworkMode); got != "ctf-awd-target-10" {
		t.Fatalf("NetworkMode = %q, want ctf-awd-target-10", got)
	}
	if !containsString(spec.ContainerConfig.Env, "CHECKER_TARGET_ALLOWLIST=10.10.0.23:8080") {
		t.Fatalf("Env = %v, want CHECKER_TARGET_ALLOWLIST", spec.ContainerConfig.Env)
	}
}

func TestDockerCheckerRunnerRejectsUnsafeCheckerFilePath(t *testing.T) {
	t.Parallel()

	if _, err := cleanCheckerFilePath("../secret.py"); err == nil {
		t.Fatal("expected parent traversal path to be rejected")
	}
	if _, err := cleanCheckerFilePath("/etc/passwd"); err == nil {
		t.Fatal("expected absolute path to be rejected")
	}
}

func TestParseCheckerJSONOutput(t *testing.T) {
	t.Parallel()

	status, reason, err := parseCheckerJSONOutput(`{"status":"ok","reason":"flag_roundtrip_passed"}`)
	if err != nil {
		t.Fatalf("parseCheckerJSONOutput() error = %v", err)
	}
	if status != contestports.CheckerRunStatusOK || reason != "flag_roundtrip_passed" {
		t.Fatalf("status/reason = %s/%s, want ok/flag_roundtrip_passed", status, reason)
	}

	if _, _, err := parseCheckerJSONOutput(`not-json`); err == nil {
		t.Fatal("expected invalid JSON to fail")
	}
}

func TestMaterializeCheckerFilesUsesConfiguredHostRoot(t *testing.T) {
	t.Parallel()

	hostRoot := filepath.Join(t.TempDir(), "checker-sandboxes")
	workDir, err := materializeCheckerFiles([]contestports.CheckerRunFile{
		{Path: "docker/check/check.py", Content: []byte("print('ok')\n"), Mode: 0o500},
	}, hostRoot)
	if err != nil {
		t.Fatalf("materializeCheckerFiles() error = %v", err)
	}
	defer func() {
		_ = os.RemoveAll(workDir)
	}()

	if !strings.HasPrefix(workDir, hostRoot+string(filepath.Separator)) {
		t.Fatalf("workDir = %q, want under %q", workDir, hostRoot)
	}
	if _, err := os.Stat(filepath.Join(workDir, "docker/check/check.py")); err != nil {
		t.Fatalf("expected checker file to be materialized: %v", err)
	}
}

func checkerSandboxConfigForTest() config.CheckerSandboxConfig {
	return config.CheckerSandboxConfig{
		Image:            "python:3.12-alpine",
		User:             "65532:65532",
		HostWorkRoot:     "",
		WorkDir:          "/checker",
		Timeout:          10 * time.Second,
		CPUQuota:         0.5,
		MemoryBytes:      128 * 1024 * 1024,
		PidsLimit:        64,
		NofileLimit:      128,
		OutputLimitBytes: 32768,
	}
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
