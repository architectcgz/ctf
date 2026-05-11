package composition

import (
	"context"
	"ctf-platform/internal/authctx"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	instanceports "ctf-platform/internal/module/instance/ports"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func TestBuildRuntimeEngineProvidesReachableRuntimeInTestEnv(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newRootTestDependencies(t)
	cfg.Container = config.ContainerConfig{
		DefaultExposedPort: 80,
		PortRangeStart:     35000,
		PortRangeEnd:       35010,
		PublicHost:         "127.0.0.1",
	}

	root, err := BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}

	engine := buildRuntimeEngine(root)
	service := runtimecmd.NewProvisioningService(nil, engine, &cfg.Container, zap.NewNop())

	containerID, networkID, hostPort, _, err := service.CreateContainer(context.Background(), "ctf/test:v1", nil, 35001)
	if err != nil {
		t.Fatalf("CreateContainer() error = %v", err)
	}
	if containerID == "" {
		t.Fatal("expected non-empty container id")
	}
	if networkID == "" {
		t.Fatal("expected non-empty network id")
	}
	if hostPort != 35001 {
		t.Fatalf("expected host port 35001, got %d", hostPort)
	}

	client := &http.Client{Timeout: time.Second}
	resp, err := client.Get("http://127.0.0.1:35001")
	if err != nil {
		t.Fatalf("expected runtime access url to be reachable, got error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected runtime probe status 200, got %d", resp.StatusCode)
	}

	cleanup := runtimecmd.NewRuntimeCleanupService(engine, nil, zap.NewNop())
	if err := cleanup.RemoveContainer(context.Background(), containerID); err != nil {
		t.Fatalf("RemoveContainer() error = %v", err)
	}
	if engine != nil && networkID != "" {
		if err := engine.RemoveNetwork(context.Background(), networkID); err != nil {
			t.Fatalf("RemoveNetwork() error = %v", err)
		}
	}
}

func TestRuntimeHTTPServiceAdapterReturnsSSHAccessWithoutProfile(t *testing.T) {
	expiresAt := time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC)
	adapter := newRuntimeHTTPServiceAdapter(
		nil,
		nil,
		stubRuntimeHTTPProxyTickets{ticket: "ticket-secret", expiresAt: expiresAt},
		nil,
		0,
		0,
		true,
		"ssh.ctf.local",
		2222,
	)

	resp, err := adapter.IssueAWDDefenseSSHTicket(context.Background(), authctx.CurrentUser{
		UserID:   1001,
		Username: "student",
		Role:     "student",
	}, 5, 12)
	if err != nil {
		t.Fatalf("IssueAWDDefenseSSHTicket() error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected ssh access response")
	}
	if resp.Host != "ssh.ctf.local" ||
		resp.Port != 2222 ||
		resp.Username != "student+5+12" ||
		resp.Password != "ticket-secret" ||
		resp.Command != "ssh student+5+12@ssh.ctf.local -p 2222" ||
		resp.ExpiresAt != expiresAt.Format(time.RFC3339) {
		t.Fatalf("unexpected ssh access response: %+v", resp)
	}
}

func TestRuntimeHTTPServiceAdapterDelegatesAWDDefenseWorkbenchCalls(t *testing.T) {
	workbench := &stubRuntimeHTTPAWDDefenseWorkbenchService{}
	adapter := newRuntimeHTTPServiceAdapter(
		nil,
		nil,
		nil,
		workbench,
		0,
		0,
		false,
		"",
		0,
	)

	fileResp, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/challenge_app.py")
	if err != nil {
		t.Fatalf("ReadAWDDefenseFile() error = %v", err)
	}
	if fileResp.Path != "docker/challenge_app.py" || fileResp.Content != "print('delegated')" {
		t.Fatalf("unexpected file response: %+v", fileResp)
	}

	dirResp, err := adapter.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker")
	if err != nil {
		t.Fatalf("ListAWDDefenseDirectory() error = %v", err)
	}
	if dirResp.Path != "docker" || len(dirResp.Entries) != 1 {
		t.Fatalf("unexpected directory response: %+v", dirResp)
	}
	if dirResp.Entries[0].Path != "docker/challenge_app.py" || dirResp.Entries[0].Type != "file" {
		t.Fatalf("unexpected directory entry: %+v", dirResp.Entries)
	}

	saveResp, err := adapter.SaveAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, dto.AWDDefenseFileSaveReq{
		Path:    "docker/challenge_app.py",
		Content: "print('fixed')",
		Backup:  true,
	})
	if err != nil {
		t.Fatalf("SaveAWDDefenseFile() error = %v", err)
	}
	if saveResp.Path != "docker/challenge_app.py" || saveResp.Size != len("print('fixed')") {
		t.Fatalf("unexpected save response: %+v", saveResp)
	}

	commandResp, err := adapter.RunAWDDefenseCommand(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, dto.AWDDefenseCommandReq{
		Command: "ls",
	})
	if err != nil {
		t.Fatalf("RunAWDDefenseCommand() error = %v", err)
	}
	if commandResp.Command != "ls" || commandResp.Output != "delegated" {
		t.Fatalf("unexpected command response: %+v", commandResp)
	}

	if len(workbench.calls) != 4 {
		t.Fatalf("expected 4 delegated calls, got %+v", workbench.calls)
	}
	expectedCalls := []string{"read", "list", "save", "run"}
	for idx, call := range expectedCalls {
		if workbench.calls[idx] != call {
			t.Fatalf("calls[%d] = %q, want %q", idx, workbench.calls[idx], call)
		}
	}
}

func TestInstanceAWDDefenseWorkbenchRuntimeAdapterMapsDirectoryEntries(t *testing.T) {
	adapter := newInstanceAWDDefenseWorkbenchRuntime(stubRuntimeContainerFileRuntime{
		entries: []runtimeports.ContainerDirectoryEntry{
			{Name: "challenge_app.py", Type: "file", Size: 42},
			{Name: "templates", Type: "dir"},
		},
	})
	entries, err := adapter.ListDirectoryFromContainer(context.Background(), "container-12", "/home/student", 10)
	if err != nil {
		t.Fatalf("ListDirectoryFromContainer() error = %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %+v", entries)
	}
	if entries[0] != (instanceports.ContainerDirectoryEntry{Name: "challenge_app.py", Type: "file", Size: 42}) {
		t.Fatalf("unexpected first entry: %+v", entries[0])
	}
	if entries[1] != (instanceports.ContainerDirectoryEntry{Name: "templates", Type: "dir"}) {
		t.Fatalf("unexpected second entry: %+v", entries[1])
	}
}

type stubRuntimeHTTPProxyTickets struct {
	ticket    string
	expiresAt time.Time
}

func (s stubRuntimeHTTPProxyTickets) IssueTicket(context.Context, authctx.CurrentUser, int64) (string, time.Time, error) {
	return s.ticket, s.expiresAt, nil
}

func (s stubRuntimeHTTPProxyTickets) IssueAWDTargetTicket(context.Context, authctx.CurrentUser, int64, int64, int64) (string, time.Time, error) {
	return s.ticket, s.expiresAt, nil
}

func (s stubRuntimeHTTPProxyTickets) IssueAWDDefenseSSHTicket(context.Context, authctx.CurrentUser, int64, int64) (string, time.Time, error) {
	return s.ticket, s.expiresAt, nil
}

func (s stubRuntimeHTTPProxyTickets) ResolveTicket(context.Context, string) (*runtimeports.ProxyTicketClaims, error) {
	return nil, nil
}

func (s stubRuntimeHTTPProxyTickets) ResolveAWDTargetAccessURL(context.Context, *runtimeports.ProxyTicketClaims, int64, int64, int64) (string, error) {
	return "", nil
}

func (s stubRuntimeHTTPProxyTickets) MaxAge() int {
	return 900
}

type stubRuntimeHTTPProxyTicketReader struct {
	scope *runtimeports.AWDDefenseSSHScope
}

func (s stubRuntimeHTTPProxyTicketReader) FindByID(context.Context, int64) (*model.Instance, error) {
	return nil, nil
}

func (s stubRuntimeHTTPProxyTicketReader) FindAWDTargetProxyScope(context.Context, int64, int64, int64, int64) (*runtimeports.AWDTargetProxyScope, error) {
	return nil, nil
}

func (s stubRuntimeHTTPProxyTicketReader) FindAWDDefenseSSHScope(context.Context, int64, int64, int64) (*runtimeports.AWDDefenseSSHScope, error) {
	return s.scope, nil
}

type stubRuntimeHTTPAWDDefenseWorkbenchService struct {
	calls []string
}

func (s *stubRuntimeHTTPAWDDefenseWorkbenchService) ReadAWDDefenseFile(context.Context, authctx.CurrentUser, int64, int64, string) (*dto.AWDDefenseFileResp, error) {
	s.calls = append(s.calls, "read")
	return &dto.AWDDefenseFileResp{Path: "docker/challenge_app.py", Content: "print('delegated')", Size: 18}, nil
}

func (s *stubRuntimeHTTPAWDDefenseWorkbenchService) ListAWDDefenseDirectory(context.Context, authctx.CurrentUser, int64, int64, string) (*dto.AWDDefenseDirectoryResp, error) {
	s.calls = append(s.calls, "list")
	return &dto.AWDDefenseDirectoryResp{
		Path: "docker",
		Entries: []dto.AWDDefenseDirectoryEntryResp{
			{Name: "challenge_app.py", Path: "docker/challenge_app.py", Type: "file", Size: 18},
		},
	}, nil
}

func (s *stubRuntimeHTTPAWDDefenseWorkbenchService) SaveAWDDefenseFile(_ context.Context, _ authctx.CurrentUser, _ int64, _ int64, req dto.AWDDefenseFileSaveReq) (*dto.AWDDefenseFileSaveResp, error) {
	s.calls = append(s.calls, "save")
	return &dto.AWDDefenseFileSaveResp{Path: req.Path, Size: len(req.Content), BackupPath: req.Path + ".bak"}, nil
}

func (s *stubRuntimeHTTPAWDDefenseWorkbenchService) RunAWDDefenseCommand(_ context.Context, _ authctx.CurrentUser, _ int64, _ int64, req dto.AWDDefenseCommandReq) (*dto.AWDDefenseCommandResp, error) {
	s.calls = append(s.calls, "run")
	return &dto.AWDDefenseCommandResp{Command: req.Command, Output: "delegated"}, nil
}

type stubRuntimeContainerFileRuntime struct {
	entries []runtimeports.ContainerDirectoryEntry
}

func (s stubRuntimeContainerFileRuntime) ReadFileFromContainer(context.Context, string, string, int64) ([]byte, error) {
	return nil, nil
}

func (s stubRuntimeContainerFileRuntime) ListDirectoryFromContainer(context.Context, string, string, int) ([]runtimeports.ContainerDirectoryEntry, error) {
	return append([]runtimeports.ContainerDirectoryEntry(nil), s.entries...), nil
}

func (s stubRuntimeContainerFileRuntime) WriteFileToContainer(context.Context, string, string, []byte) error {
	return nil
}

func (s stubRuntimeContainerFileRuntime) ExecContainerCommand(context.Context, string, []string, []byte, int64) ([]byte, error) {
	return nil, nil
}
