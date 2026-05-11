package application

import (
	"context"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	instanceports "ctf-platform/internal/module/instance/ports"
	"ctf-platform/pkg/errcode"
)

func TestAWDDefenseWorkbenchServiceReadsAndListsEditablePaths(t *testing.T) {
	service := NewAWDDefenseWorkbenchService(
		stubAWDDefenseWorkbenchScopeReader{scope: testAWDDefenseSSHScope()},
		stubAWDDefenseWorkbenchRuntime{
			files: map[string][]byte{
				"/home/student/challenge_app.py": []byte("print('vuln')"),
			},
			entries: []instanceports.ContainerDirectoryEntry{
				{Name: "app.py", Type: "file", Size: 13},
				{Name: "ctf_runtime.py", Type: "file", Size: 64},
				{Name: "requirements.txt", Type: "file", Size: 128},
				{Name: "challenge_app.py", Type: "file", Size: 42},
				{Name: "templates", Type: "dir"},
			},
		},
		AWDDefenseWorkbenchConfig{ReadOnlyEnabled: true, Root: "/home/student"},
	)

	fileResp, err := service.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/challenge_app.py")
	if err != nil {
		t.Fatalf("ReadAWDDefenseFile() error = %v", err)
	}
	if fileResp.Path != "docker/challenge_app.py" || fileResp.Content != "print('vuln')" {
		t.Fatalf("unexpected file response: %+v", fileResp)
	}

	dirResp, err := service.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker")
	if err != nil {
		t.Fatalf("ListAWDDefenseDirectory() error = %v", err)
	}
	if dirResp.Path != "docker" || len(dirResp.Entries) != 2 {
		t.Fatalf("unexpected directory response: %+v", dirResp)
	}
	if dirResp.Entries[0].Path != "docker/templates" || dirResp.Entries[0].Type != "dir" {
		t.Fatalf("expected editable templates directory, got %+v", dirResp.Entries)
	}
	if dirResp.Entries[1].Path != "docker/challenge_app.py" || dirResp.Entries[1].Name != "challenge_app.py" {
		t.Fatalf("unexpected directory response: %+v", dirResp)
	}
}

func TestAWDDefenseWorkbenchServiceRejectsWhenReadOnlyDisabled(t *testing.T) {
	service := NewAWDDefenseWorkbenchService(
		stubAWDDefenseWorkbenchScopeReader{scope: testAWDDefenseSSHScope()},
		stubAWDDefenseWorkbenchRuntime{},
		AWDDefenseWorkbenchConfig{ReadOnlyEnabled: false, Root: "/home/student"},
	)
	if _, err := service.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/challenge_app.py"); err == nil {
		t.Fatal("expected disabled workbench read to fail")
	}
	if _, err := service.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker"); err == nil {
		t.Fatal("expected disabled workbench list to fail")
	}
}

func TestAWDDefenseWorkbenchServiceRejectsSensitivePaths(t *testing.T) {
	service := NewAWDDefenseWorkbenchService(
		stubAWDDefenseWorkbenchScopeReader{scope: testAWDDefenseSSHScope()},
		stubAWDDefenseWorkbenchRuntime{},
		AWDDefenseWorkbenchConfig{ReadOnlyEnabled: true, Root: "/home/student"},
	)

	blocked := []string{".env", ".env.local", "prod.env", ".ssh/id_rsa", "proc/self/environ", "sys/kernel", "dev/null", "run/secrets", "var/run/docker.sock"}
	for _, path := range blocked {
		if _, err := service.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, path); err == nil {
			t.Fatalf("expected read path %q to fail", path)
		}
	}
	if _, err := service.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/app.py"); err == nil {
		t.Fatal("expected protected path to fail")
	}
	if _, err := service.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "../app.py"); err == nil {
		t.Fatal("expected traversal path to fail")
	}
	if _, err := service.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/check"); err == nil {
		t.Fatal("expected protected directory to fail")
	}
}

func TestAWDDefenseWorkbenchServiceRejectsInvalidRoot(t *testing.T) {
	roots := []string{"", ".", "/", "app"}
	for _, root := range roots {
		service := NewAWDDefenseWorkbenchService(
			stubAWDDefenseWorkbenchScopeReader{scope: testAWDDefenseSSHScope()},
			stubAWDDefenseWorkbenchRuntime{},
			AWDDefenseWorkbenchConfig{ReadOnlyEnabled: true, Root: root},
		)
		if _, err := service.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/challenge_app.py"); err == nil {
			t.Fatalf("expected root %q to fail", root)
		}
	}
}

func TestAWDDefenseWorkbenchServiceMapsDockerPathsToContainerRoot(t *testing.T) {
	runtime := &recordingAWDDefenseWorkbenchRuntime{fileContent: []byte("print('vuln')")}
	service := NewAWDDefenseWorkbenchService(
		stubAWDDefenseWorkbenchScopeReader{scope: testAWDDefenseSSHScopeWithEditable("docker/src/app.py")},
		runtime,
		AWDDefenseWorkbenchConfig{ReadOnlyEnabled: true, Root: "/home/student"},
	)

	if _, err := service.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/src/app.py"); err != nil {
		t.Fatalf("ReadAWDDefenseFile() error = %v", err)
	}
	if runtime.readPath != "/home/student/src/app.py" {
		t.Fatalf("expected rooted read path, got %q", runtime.readPath)
	}

	if _, err := service.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/src"); err != nil {
		t.Fatalf("ListAWDDefenseDirectory() error = %v", err)
	}
	if runtime.listPath != "/home/student/src" {
		t.Fatalf("expected rooted list path, got %q", runtime.listPath)
	}
}

func TestAWDDefenseWorkbenchServiceSavesOnlyEditableFiles(t *testing.T) {
	runtime := &recordingAWDDefenseWorkbenchRuntime{fileContent: []byte("print('old')")}
	service := NewAWDDefenseWorkbenchService(
		stubAWDDefenseWorkbenchScopeReader{scope: testAWDDefenseSSHScope()},
		runtime,
		AWDDefenseWorkbenchConfig{ReadOnlyEnabled: true, Root: "/home/student"},
	)

	resp, err := service.SaveAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, dto.AWDDefenseFileSaveReq{
		Path:    "docker/challenge_app.py",
		Content: "print('fixed')",
		Backup:  true,
	})
	if err != nil {
		t.Fatalf("SaveAWDDefenseFile() error = %v", err)
	}
	if runtime.readPath != "/home/student/challenge_app.py" {
		t.Fatalf("expected backup read from mapped editable path, got %q", runtime.readPath)
	}
	if len(runtime.writePaths) != 2 {
		t.Fatalf("expected backup and save writes, got %+v", runtime.writePaths)
	}
	if !strings.HasPrefix(runtime.writePaths[0], "/home/student/challenge_app.py.bak.") {
		t.Fatalf("expected mapped backup path, got %+v", runtime.writePaths)
	}
	if runtime.writePaths[1] != "/home/student/challenge_app.py" {
		t.Fatalf("expected mapped save path, got %+v", runtime.writePaths)
	}
	if resp.Path != "docker/challenge_app.py" || !strings.HasPrefix(resp.BackupPath, "docker/challenge_app.py.bak.") {
		t.Fatalf("unexpected save response: %+v", resp)
	}

	if _, err := service.SaveAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, dto.AWDDefenseFileSaveReq{
		Path:    "docker/app.py",
		Content: "print('nope')",
	}); err == nil {
		t.Fatal("expected protected path save to fail")
	}
}

func TestAWDDefenseWorkbenchServiceRunsScopedContainerCommand(t *testing.T) {
	runtime := &recordingAWDDefenseWorkbenchRuntime{commandOutput: []byte("ok")}
	service := NewAWDDefenseWorkbenchService(
		stubAWDDefenseWorkbenchScopeReader{scope: testAWDDefenseSSHScope()},
		runtime,
		AWDDefenseWorkbenchConfig{ReadOnlyEnabled: false, Root: "/home/student"},
	)

	resp, err := service.RunAWDDefenseCommand(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, dto.AWDDefenseCommandReq{
		Command: "ls",
	})
	if err != nil {
		t.Fatalf("RunAWDDefenseCommand() error = %v", err)
	}
	if resp.Command != "ls" || resp.Output != "ok" {
		t.Fatalf("unexpected command response: %+v", resp)
	}
	if runtime.commandContainerID != "container-12" {
		t.Fatalf("expected scoped container id, got %q", runtime.commandContainerID)
	}
	if len(runtime.commandArgs) != 3 || runtime.commandArgs[0] != "/bin/sh" || runtime.commandArgs[2] != "ls" {
		t.Fatalf("unexpected command args: %+v", runtime.commandArgs)
	}
	if runtime.commandLimit != 64*1024 {
		t.Fatalf("expected command limit 65536, got %d", runtime.commandLimit)
	}
}

func TestAWDDefenseWorkbenchServiceRejectsInvalidCommand(t *testing.T) {
	service := NewAWDDefenseWorkbenchService(
		stubAWDDefenseWorkbenchScopeReader{scope: testAWDDefenseSSHScope()},
		stubAWDDefenseWorkbenchRuntime{},
		AWDDefenseWorkbenchConfig{},
	)
	_, err := service.RunAWDDefenseCommand(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, dto.AWDDefenseCommandReq{Command: "  "})
	appErr, ok := err.(*errcode.AppError)
	if err == nil || !ok || appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params, got %v", err)
	}
}

type stubAWDDefenseWorkbenchScopeReader struct {
	scope *instanceports.AWDDefenseSSHScope
	err   error
}

func (s stubAWDDefenseWorkbenchScopeReader) FindAWDDefenseSSHScope(context.Context, int64, int64, int64) (*instanceports.AWDDefenseSSHScope, error) {
	return s.scope, s.err
}

type stubAWDDefenseWorkbenchRuntime struct {
	files   map[string][]byte
	entries []instanceports.ContainerDirectoryEntry
}

func (s stubAWDDefenseWorkbenchRuntime) ReadFileFromContainer(_ context.Context, _ string, filePath string, _ int64) ([]byte, error) {
	return append([]byte(nil), s.files[filePath]...), nil
}

func (s stubAWDDefenseWorkbenchRuntime) ListDirectoryFromContainer(context.Context, string, string, int) ([]instanceports.ContainerDirectoryEntry, error) {
	return append([]instanceports.ContainerDirectoryEntry(nil), s.entries...), nil
}

func (s stubAWDDefenseWorkbenchRuntime) WriteFileToContainer(context.Context, string, string, []byte) error {
	return nil
}

func (s stubAWDDefenseWorkbenchRuntime) ExecContainerCommand(context.Context, string, []string, []byte, int64) ([]byte, error) {
	return nil, nil
}

type recordingAWDDefenseWorkbenchRuntime struct {
	readPath           string
	listPath           string
	fileContent        []byte
	writePaths         []string
	commandContainerID string
	commandArgs        []string
	commandLimit       int64
	commandOutput      []byte
}

func (s *recordingAWDDefenseWorkbenchRuntime) ReadFileFromContainer(_ context.Context, _ string, filePath string, _ int64) ([]byte, error) {
	s.readPath = filePath
	return append([]byte(nil), s.fileContent...), nil
}

func (s *recordingAWDDefenseWorkbenchRuntime) ListDirectoryFromContainer(_ context.Context, _ string, dirPath string, _ int) ([]instanceports.ContainerDirectoryEntry, error) {
	s.listPath = dirPath
	return []instanceports.ContainerDirectoryEntry{
		{Name: "challenge_app.py", Type: "file", Size: 13},
		{Name: "templates", Type: "dir"},
	}, nil
}

func (s *recordingAWDDefenseWorkbenchRuntime) WriteFileToContainer(_ context.Context, _ string, filePath string, _ []byte) error {
	s.writePaths = append(s.writePaths, filePath)
	return nil
}

func (s *recordingAWDDefenseWorkbenchRuntime) ExecContainerCommand(_ context.Context, containerID string, command []string, _ []byte, limit int64) ([]byte, error) {
	s.commandContainerID = containerID
	s.commandArgs = append([]string(nil), command...)
	s.commandLimit = limit
	return append([]byte(nil), s.commandOutput...), nil
}

func testAWDDefenseSSHScope() *instanceports.AWDDefenseSSHScope {
	return testAWDDefenseSSHScopeWithEditable("docker/challenge_app.py", "docker/templates/mail.html")
}

func testAWDDefenseSSHScopeWithEditable(editablePaths ...string) *instanceports.AWDDefenseSSHScope {
	return &instanceports.AWDDefenseSSHScope{
		ContainerID:    "container-12",
		EditablePaths:  append([]string(nil), editablePaths...),
		ProtectedPaths: []string{"docker/app.py", "docker/ctf_runtime.py", "docker/check/check.py", "challenge.yml"},
	}
}

func TestAWDDefenseWorkbenchServiceUsesMappedDockerRoot(t *testing.T) {
	runtime := &recordingAWDDefenseWorkbenchRuntime{fileContent: []byte("print('vuln')")}
	service := NewAWDDefenseWorkbenchService(
		stubAWDDefenseWorkbenchScopeReader{scope: testAWDDefenseSSHScope()},
		runtime,
		AWDDefenseWorkbenchConfig{ReadOnlyEnabled: true, Root: "/home/student"},
	)

	if _, err := service.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/challenge_app.py"); err != nil {
		t.Fatalf("ReadAWDDefenseFile() error = %v", err)
	}
	if runtime.readPath != "/home/student/challenge_app.py" {
		t.Fatalf("expected mapped read path, got %q", runtime.readPath)
	}

	if _, err := service.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker"); err != nil {
		t.Fatalf("ListAWDDefenseDirectory() error = %v", err)
	}
	if runtime.listPath != "/home/student" {
		t.Fatalf("expected mapped docker directory path, got %q", runtime.listPath)
	}
}

func TestAWDDefenseWorkbenchServiceCommandDeadline(t *testing.T) {
	runtime := &deadlineCapturingAWDDefenseWorkbenchRuntime{}
	service := NewAWDDefenseWorkbenchService(
		stubAWDDefenseWorkbenchScopeReader{scope: testAWDDefenseSSHScope()},
		runtime,
		AWDDefenseWorkbenchConfig{},
	)

	if _, err := service.RunAWDDefenseCommand(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, dto.AWDDefenseCommandReq{
		Command: "pwd",
	}); err != nil {
		t.Fatalf("RunAWDDefenseCommand() error = %v", err)
	}
	if runtime.deadline.IsZero() {
		t.Fatal("expected command context deadline")
	}
	if remaining := time.Until(runtime.deadline); remaining > 6*time.Second || remaining < 0 {
		t.Fatalf("unexpected remaining deadline window: %s", remaining)
	}
}

type deadlineCapturingAWDDefenseWorkbenchRuntime struct {
	deadline time.Time
}

func (s *deadlineCapturingAWDDefenseWorkbenchRuntime) ReadFileFromContainer(context.Context, string, string, int64) ([]byte, error) {
	return nil, nil
}

func (s *deadlineCapturingAWDDefenseWorkbenchRuntime) ListDirectoryFromContainer(context.Context, string, string, int) ([]instanceports.ContainerDirectoryEntry, error) {
	return nil, nil
}

func (s *deadlineCapturingAWDDefenseWorkbenchRuntime) WriteFileToContainer(context.Context, string, string, []byte) error {
	return nil
}

func (s *deadlineCapturingAWDDefenseWorkbenchRuntime) ExecContainerCommand(ctx context.Context, _ string, _ []string, _ []byte, _ int64) ([]byte, error) {
	s.deadline, _ = ctx.Deadline()
	return []byte("ok"), nil
}
