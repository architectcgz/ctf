package composition

import (
	"context"
	"ctf-platform/internal/authctx"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
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

func TestRuntimeHTTPServiceAdapterReadsAWDDefenseWorkbenchWhenEnabled(t *testing.T) {
	adapter := newRuntimeHTTPServiceAdapter(
		nil,
		nil,
		nil,
		stubRuntimeHTTPProxyTicketReader{scope: testAWDDefenseSSHScope()},
		stubRuntimeHTTPDefenseWorkbench{
			files: map[string][]byte{
				"/home/student/challenge_app.py": []byte("print('vuln')"),
			},
			entries: []runtimeports.ContainerDirectoryEntry{
				{Name: "app.py", Type: "file", Size: 13},
				{Name: "ctf_runtime.py", Type: "file", Size: 64},
				{Name: "requirements.txt", Type: "file", Size: 128},
				{Name: "challenge_app.py", Type: "file", Size: 42},
				{Name: "templates", Type: "dir"},
			},
		},
		0,
		false,
		"",
		0,
		true,
		"/home/student",
	)

	fileResp, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/challenge_app.py")
	if err != nil {
		t.Fatalf("ReadAWDDefenseFile() error = %v", err)
	}
	if fileResp.Path != "docker/challenge_app.py" || fileResp.Content != "print('vuln')" {
		t.Fatalf("unexpected file response: %+v", fileResp)
	}

	dirResp, err := adapter.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker")
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

func TestRuntimeHTTPServiceAdapterRejectsAWDDefenseWorkbenchWhenDisabled(t *testing.T) {
	adapter := newRuntimeHTTPServiceAdapter(
		nil,
		nil,
		nil,
		stubRuntimeHTTPProxyTicketReader{scope: testAWDDefenseSSHScope()},
		stubRuntimeHTTPDefenseWorkbench{},
		0,
		false,
		"",
		0,
		false,
		"/home/student",
	)

	if _, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/challenge_app.py"); err == nil {
		t.Fatal("expected disabled workbench read to fail")
	}
	if _, err := adapter.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker"); err == nil {
		t.Fatal("expected disabled workbench list to fail")
	}
}

func TestRuntimeHTTPServiceAdapterRejectsAWDDefenseSensitivePaths(t *testing.T) {
	adapter := newRuntimeHTTPServiceAdapter(
		nil,
		nil,
		nil,
		stubRuntimeHTTPProxyTicketReader{scope: testAWDDefenseSSHScope()},
		stubRuntimeHTTPDefenseWorkbench{},
		0,
		false,
		"",
		0,
		true,
		"/home/student",
	)

	blocked := []string{".env", ".env.local", "prod.env", ".ssh/id_rsa", "proc/self/environ", "sys/kernel", "dev/null", "run/secrets", "var/run/docker.sock"}
	for _, path := range blocked {
		if _, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, path); err == nil {
			t.Fatalf("expected read path %q to fail", path)
		}
	}
	if _, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/app.py"); err == nil {
		t.Fatal("expected protected path to fail")
	}
	if _, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "../app.py"); err == nil {
		t.Fatal("expected traversal path to fail")
	}
	if _, err := adapter.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/check"); err == nil {
		t.Fatal("expected protected directory to fail")
	}
}

func TestRuntimeHTTPServiceAdapterRejectsAWDDefenseWorkbenchWithoutAbsoluteRoot(t *testing.T) {
	roots := []string{"", ".", "/", "app"}
	for _, root := range roots {
		adapter := newRuntimeHTTPServiceAdapter(
			nil,
			nil,
			nil,
			stubRuntimeHTTPProxyTicketReader{scope: testAWDDefenseSSHScope()},
			stubRuntimeHTTPDefenseWorkbench{},
			0,
			false,
			"",
			0,
			true,
			root,
		)
		if _, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/challenge_app.py"); err == nil {
			t.Fatalf("expected root %q to fail", root)
		}
	}
}

func TestRuntimeHTTPServiceAdapterRootsAWDDefenseContainerPath(t *testing.T) {
	workbench := &recordingRuntimeHTTPDefenseWorkbench{
		fileContent: []byte("print('vuln')"),
	}
	adapter := newRuntimeHTTPServiceAdapter(
		nil,
		nil,
		nil,
		stubRuntimeHTTPProxyTicketReader{scope: testAWDDefenseSSHScopeWithEditable("docker/src/app.py")},
		workbench,
		0,
		false,
		"",
		0,
		true,
		"/home/student",
	)

	if _, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/src/app.py"); err != nil {
		t.Fatalf("ReadAWDDefenseFile() error = %v", err)
	}
	if workbench.readPath != "/home/student/src/app.py" {
		t.Fatalf("expected rooted read path, got %q", workbench.readPath)
	}

	if _, err := adapter.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/src"); err != nil {
		t.Fatalf("ListAWDDefenseDirectory() error = %v", err)
	}
	if workbench.listPath != "/home/student/src" {
		t.Fatalf("expected rooted list path, got %q", workbench.listPath)
	}
}

func TestRuntimeHTTPServiceAdapterMapsAWDPackageDockerPathsToContainerRoot(t *testing.T) {
	workbench := &recordingRuntimeHTTPDefenseWorkbench{
		fileContent: []byte("print('vuln')"),
	}
	adapter := newRuntimeHTTPServiceAdapter(
		nil,
		nil,
		nil,
		stubRuntimeHTTPProxyTicketReader{scope: testAWDDefenseSSHScope()},
		workbench,
		0,
		false,
		"",
		0,
		true,
		"/home/student",
	)

	if _, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker/challenge_app.py"); err != nil {
		t.Fatalf("ReadAWDDefenseFile() error = %v", err)
	}
	if workbench.readPath != "/home/student/challenge_app.py" {
		t.Fatalf("expected mapped read path, got %q", workbench.readPath)
	}

	if _, err := adapter.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "docker"); err != nil {
		t.Fatalf("ListAWDDefenseDirectory() error = %v", err)
	}
	if workbench.listPath != "/home/student" {
		t.Fatalf("expected mapped docker directory path, got %q", workbench.listPath)
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

type stubRuntimeHTTPDefenseWorkbench struct {
	files   map[string][]byte
	entries []runtimeports.ContainerDirectoryEntry
}

func (s stubRuntimeHTTPDefenseWorkbench) ReadFileFromContainer(_ context.Context, _ string, filePath string, _ int64) ([]byte, error) {
	return append([]byte(nil), s.files[filePath]...), nil
}

func (s stubRuntimeHTTPDefenseWorkbench) ListDirectoryFromContainer(context.Context, string, string, int) ([]runtimeports.ContainerDirectoryEntry, error) {
	return append([]runtimeports.ContainerDirectoryEntry(nil), s.entries...), nil
}

func (s stubRuntimeHTTPDefenseWorkbench) WriteFileToContainer(context.Context, string, string, []byte) error {
	return nil
}

func (s stubRuntimeHTTPDefenseWorkbench) ExecContainerCommand(context.Context, string, []string, []byte, int64) ([]byte, error) {
	return nil, nil
}

type recordingRuntimeHTTPDefenseWorkbench struct {
	readPath    string
	listPath    string
	fileContent []byte
}

func (s *recordingRuntimeHTTPDefenseWorkbench) ReadFileFromContainer(_ context.Context, _ string, filePath string, _ int64) ([]byte, error) {
	s.readPath = filePath
	return append([]byte(nil), s.fileContent...), nil
}

func (s *recordingRuntimeHTTPDefenseWorkbench) ListDirectoryFromContainer(_ context.Context, _ string, dirPath string, _ int) ([]runtimeports.ContainerDirectoryEntry, error) {
	s.listPath = dirPath
	return []runtimeports.ContainerDirectoryEntry{
		{Name: "challenge_app.py", Type: "file", Size: 13},
		{Name: "templates", Type: "dir"},
	}, nil
}

func (s *recordingRuntimeHTTPDefenseWorkbench) WriteFileToContainer(context.Context, string, string, []byte) error {
	return nil
}

func (s *recordingRuntimeHTTPDefenseWorkbench) ExecContainerCommand(context.Context, string, []string, []byte, int64) ([]byte, error) {
	return nil, nil
}

func testAWDDefenseSSHScope() *runtimeports.AWDDefenseSSHScope {
	return testAWDDefenseSSHScopeWithEditable("docker/challenge_app.py", "docker/templates/mail.html")
}

func testAWDDefenseSSHScopeWithEditable(editablePaths ...string) *runtimeports.AWDDefenseSSHScope {
	return &runtimeports.AWDDefenseSSHScope{
		ContainerID:    "container-12",
		EditablePaths:  append([]string(nil), editablePaths...),
		ProtectedPaths: []string{"docker/app.py", "docker/ctf_runtime.py", "docker/check/check.py", "challenge.yml"},
	}
}
