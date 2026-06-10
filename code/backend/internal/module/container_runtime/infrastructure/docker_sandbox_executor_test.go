package infrastructure

import (
	"archive/tar"
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	networktypes "github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"ctf-platform/internal/config"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
)

func TestDockerSandboxExecutorBuildsLockedDownContainerSpec(t *testing.T) {
	t.Parallel()

	runner := NewDockerSandboxExecutorWithClient(nil, checkerSandboxConfigForTest())
	spec, err := runner.buildContainerSpec(runtimeports.SandboxExecJob{
		Runtime: "python3",
		Entry:   "docker/check/check.py",
		Args:    []string{"{{TARGET_URL}}"},
		Env: map[string]string{
			"FLAG": "flag{redacted}",
		},
		Labels: map[string]string{
			"ctf.checker.contest": "10",
			"ctf.checker.service": "20",
			"ctf.checker.team":    "30",
			"ctf.checker.round":   "4",
		},
	})
	if err != nil {
		t.Fatalf("buildContainerSpec() error = %v", err)
	}

	if spec.HostConfig.Privileged {
		t.Fatal("sandbox executor must not be privileged")
	}
	if !spec.HostConfig.ReadonlyRootfs {
		t.Fatal("sandbox executor must use readonly rootfs")
	}
	if !spec.ContainerConfig.NetworkDisabled {
		t.Fatal("sandbox executor must disable network when no target network is explicit")
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
	if len(spec.HostConfig.Mounts) != 0 {
		t.Fatalf("Mounts = %+v, want no host bind mounts", spec.HostConfig.Mounts)
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
	if spec.ContainerConfig.Labels[runtimecontracts.ComposeProjectLabelKey] != runtimecontracts.ProjectLabelValue {
		t.Fatalf("missing compose project label: %+v", spec.ContainerConfig.Labels)
	}
	if spec.ContainerConfig.Labels[runtimecontracts.ComposeServiceLabelKey] != runtimecontracts.ComposeServiceAWD {
		t.Fatalf("missing awd compose service label: %+v", spec.ContainerConfig.Labels)
	}
}

func TestDockerSandboxExecutorEnablesOnlyExplicitTargetNetwork(t *testing.T) {
	t.Parallel()

	runner := NewDockerSandboxExecutorWithClient(nil, checkerSandboxConfigForTest())
	spec, err := runner.buildContainerSpec(runtimeports.SandboxExecJob{
		Entry:           "check.py",
		NetworkMode:     "ctf-awd-target-10",
		TargetAllowlist: []string{"10.10.0.23:8080"},
	})
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

func TestDockerSandboxExecutorRejectsUnsafeCheckerFilePath(t *testing.T) {
	t.Parallel()

	if _, err := cleanSandboxFilePath("../secret.py"); err == nil {
		t.Fatal("expected parent traversal path to be rejected")
	}
	if _, err := cleanSandboxFilePath("/etc/passwd"); err == nil {
		t.Fatal("expected absolute path to be rejected")
	}
}

func TestParseCheckerJSONOutput(t *testing.T) {
	t.Parallel()

	status, reason, err := parseSandboxJSONOutput(`{"status":"ok","reason":"flag_roundtrip_passed"}`)
	if err != nil {
		t.Fatalf("parseSandboxJSONOutput() error = %v", err)
	}
	if status != runtimeports.SandboxExecStatusOK || reason != "flag_roundtrip_passed" {
		t.Fatalf("status/reason = %s/%s, want ok/flag_roundtrip_passed", status, reason)
	}

	if _, _, err := parseSandboxJSONOutput(`not-json`); err == nil {
		t.Fatal("expected invalid JSON to fail")
	}
}

func TestDockerSandboxExecutorCopiesCheckerFilesToContainerBeforeStart(t *testing.T) {
	t.Parallel()

	fake := &fakeDockerCheckerClient{
		logs: dockerCheckerLogStream(t, "checker-ok", ""),
	}
	runner := NewDockerSandboxExecutorWithClient(fake, checkerSandboxConfigForTest())

	result, err := runner.RunSandboxExec(t.Context(), runtimeports.SandboxExecJob{
		Runtime: "python3",
		Entry:   "docker/check/check.py",
		Files: []runtimeports.SandboxExecFile{
			{Path: "docker/check/check.py", Content: []byte("print('ok')\n"), Mode: 0o500},
		},
	})
	if err != nil {
		t.Fatalf("RunSandboxExec() error = %v", err)
	}
	if result.Status != runtimeports.SandboxExecStatusOK {
		t.Fatalf("result.Status = %s, want ok", result.Status)
	}
	if got := strings.Join(fake.calls, " -> "); got != "create -> copy -> start -> wait -> logs -> remove" {
		t.Fatalf("call order = %q, want create -> copy -> start -> wait -> logs -> remove", got)
	}
	if fake.copyDestination != "/checker" {
		t.Fatalf("CopyToContainer destination = %q, want /checker", fake.copyDestination)
	}
	file, ok := fake.copiedFiles["docker/check/check.py"]
	if !ok {
		t.Fatalf("copied files = %+v, want docker/check/check.py", fake.copiedFiles)
	}
	if string(file.content) != "print('ok')\n" {
		t.Fatalf("copied content = %q, want %q", string(file.content), "print('ok')\n")
	}
	if file.mode != 0o500 {
		t.Fatalf("copied mode = %o, want 0500", file.mode)
	}
}

func TestBuildCheckerFilesArchivePreservesPathsAndModes(t *testing.T) {
	t.Parallel()

	archive, err := buildSandboxFilesArchive([]runtimeports.SandboxExecFile{
		{Path: "docker/check/check.py", Content: []byte("print('ok')\n"), Mode: 0o500},
	})
	if err != nil {
		t.Fatalf("buildSandboxFilesArchive() error = %v", err)
	}

	tr := tar.NewReader(archive)
	entries := make(map[string]copiedCheckerFile)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("tar reader next error = %v", err)
		}
		content, err := io.ReadAll(tr)
		if err != nil {
			t.Fatalf("tar reader read error = %v", err)
		}
		entries[header.Name] = copiedCheckerFile{content: content, mode: header.Mode}
	}
	if _, ok := entries["docker"]; !ok {
		t.Fatalf("archive entries = %+v, want docker directory", entries)
	}
	file, ok := entries["docker/check/check.py"]
	if !ok {
		t.Fatalf("archive entries = %+v, want docker/check/check.py", entries)
	}
	if string(file.content) != "print('ok')\n" {
		t.Fatalf("archive content = %q, want %q", string(file.content), "print('ok')\n")
	}
	if file.mode != 0o500 {
		t.Fatalf("archive mode = %o, want 0500", file.mode)
	}
}

func checkerSandboxConfigForTest() config.CheckerSandboxConfig {
	return config.CheckerSandboxConfig{
		Image:            "python:3.12-alpine",
		User:             "65532:65532",
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

type fakeDockerCheckerClient struct {
	calls           []string
	logs            []byte
	copyDestination string
	copiedFiles     map[string]copiedCheckerFile
}

type copiedCheckerFile struct {
	content []byte
	mode    int64
}

func (f *fakeDockerCheckerClient) ContainerCreate(context.Context, *container.Config, *container.HostConfig, *networktypes.NetworkingConfig, *ocispec.Platform, string) (container.CreateResponse, error) {
	f.calls = append(f.calls, "create")
	return container.CreateResponse{ID: "checker-container"}, nil
}

func (f *fakeDockerCheckerClient) CopyToContainer(_ context.Context, _ string, dst string, content io.Reader, _ container.CopyToContainerOptions) error {
	f.calls = append(f.calls, "copy")
	f.copyDestination = dst

	tr := tar.NewReader(content)
	f.copiedFiles = make(map[string]copiedCheckerFile)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		payload, err := io.ReadAll(tr)
		if err != nil {
			return err
		}
		f.copiedFiles[header.Name] = copiedCheckerFile{
			content: payload,
			mode:    header.Mode,
		}
	}
}

func (f *fakeDockerCheckerClient) ContainerStart(context.Context, string, container.StartOptions) error {
	f.calls = append(f.calls, "start")
	return nil
}

func (f *fakeDockerCheckerClient) ContainerWait(context.Context, string, container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	f.calls = append(f.calls, "wait")
	waitCh := make(chan container.WaitResponse, 1)
	errCh := make(chan error, 1)
	waitCh <- container.WaitResponse{StatusCode: 0}
	return waitCh, errCh
}

func (f *fakeDockerCheckerClient) ContainerLogs(context.Context, string, container.LogsOptions) (io.ReadCloser, error) {
	f.calls = append(f.calls, "logs")
	return io.NopCloser(bytes.NewReader(f.logs)), nil
}

func (*fakeDockerCheckerClient) ContainerInspect(context.Context, string) (types.ContainerJSON, error) {
	return types.ContainerJSON{}, nil
}

func (f *fakeDockerCheckerClient) ContainerRemove(context.Context, string, container.RemoveOptions) error {
	f.calls = append(f.calls, "remove")
	return nil
}

func dockerCheckerLogStream(t *testing.T, stdout, stderr string) []byte {
	t.Helper()

	var stream bytes.Buffer
	if stdout != "" {
		writer := stdcopy.NewStdWriter(&stream, stdcopy.Stdout)
		if _, err := writer.Write([]byte(stdout)); err != nil {
			t.Fatalf("stdout log stream write error = %v", err)
		}
	}
	if stderr != "" {
		writer := stdcopy.NewStdWriter(&stream, stdcopy.Stderr)
		if _, err := writer.Write([]byte(stderr)); err != nil {
			t.Fatalf("stderr log stream write error = %v", err)
		}
	}
	return stream.Bytes()
}
