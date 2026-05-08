package runtime

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

func TestRuntimeHTTPServiceAdapterReturnsSSHAccessWithoutProfile(t *testing.T) {
	expiresAt := time.Date(2026, 4, 28, 10, 0, 0, 0, time.UTC)
	adapter := newRuntimeHTTPServiceAdapter(
		nil,
		nil,
		stubRuntimeHTTPProxyTickets{ticket: "ticket-secret", expiresAt: expiresAt},
		nil,
		nil,
		0,
		true,
		"ssh.ctf.local",
		2222,
		false,
		"",
	)

	resp, err := adapter.IssueAWDDefenseSSHTicket(context.Background(), authctx.CurrentUser{
		UserID:   1001,
		Username: "student",
		Role:     "student",
	}, 5, 12)
	if err != nil {
		t.Fatalf("IssueAWDDefenseSSHTicket() error = %v", err)
	}

	if resp.Host != "ssh.ctf.local" ||
		resp.Port != 2222 ||
		resp.Username != "student+5+12" ||
		resp.Password != "ticket-secret" ||
		resp.Command != "ssh student+5+12@ssh.ctf.local -p 2222" ||
		resp.ExpiresAt != expiresAt.Format(time.RFC3339) {
		t.Fatalf("unexpected ssh access response: %+v", resp)
	}
	if resp == nil {
		t.Fatal("expected ssh access response")
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

	blocked := []string{".env", ".env.local", ".ssh/id_rsa", "config/.env", "keys/known_hosts", "proc/self/environ", "sys/kernel", "dev/null", "run/secrets", "var/run/docker.sock"}
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

func TestRuntimeHTTPServiceAdapterSavesOnlyEditableDefenseFiles(t *testing.T) {
	workbench := &recordingRuntimeHTTPDefenseWorkbench{
		fileContent: []byte("print('old')"),
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

	resp, err := adapter.SaveAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, dto.AWDDefenseFileSaveReq{
		Path:    "docker/challenge_app.py",
		Content: "print('fixed')",
		Backup:  true,
	})
	if err != nil {
		t.Fatalf("SaveAWDDefenseFile() error = %v", err)
	}
	if workbench.readPath != "/home/student/challenge_app.py" {
		t.Fatalf("expected backup read from mapped editable path, got %q", workbench.readPath)
	}
	if len(workbench.writePaths) != 2 {
		t.Fatalf("expected backup and save writes, got %+v", workbench.writePaths)
	}
	if !strings.HasPrefix(workbench.writePaths[0], "/home/student/challenge_app.py.bak.") {
		t.Fatalf("expected mapped backup path, got %+v", workbench.writePaths)
	}
	if workbench.writePaths[1] != "/home/student/challenge_app.py" {
		t.Fatalf("expected mapped save path, got %+v", workbench.writePaths)
	}
	if resp.Path != "docker/challenge_app.py" || !strings.HasPrefix(resp.BackupPath, "docker/challenge_app.py.bak.") {
		t.Fatalf("unexpected save response: %+v", resp)
	}

	if _, err := adapter.SaveAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, dto.AWDDefenseFileSaveReq{
		Path:    "docker/app.py",
		Content: "print('nope')",
	}); err == nil {
		t.Fatal("expected protected path save to fail")
	}
}

func TestRuntimePracticeTopologyAdapterPreservesAWDNetworkFields(t *testing.T) {
	req := &practiceports.TopologyCreateRequest{
		ContainerName: "ctf-workspace-workspace-c8-t15-s21-r1",
		Networks: []practiceports.TopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-8", Shared: true},
		},
		Nodes: []practiceports.TopologyCreateNode{
			{
				Key:            "web",
				Command:        []string{"tail", "-f", "/dev/null"},
				WorkingDir:     "/workspace",
				IsEntryPoint:   true,
				NetworkKeys:    []string{model.TopologyDefaultNetworkKey},
				NetworkAliases: []string{"awd-c8-t15-s21"},
				Mounts: []model.ContainerMount{
					{Source: "ctf-workspace-root-c8-t15-s21-r1-src", Target: "/workspace/src", ReadOnly: false},
				},
			},
		},
		DisableEntryPortPublishing: true,
	}

	got := toRuntimeTopologyCreateRequest(req)
	if len(got.Networks) != 1 || got.Networks[0].Name != "ctf-awd-contest-8" || !got.Networks[0].Shared {
		t.Fatalf("expected AWD network fields to be preserved, got %+v", got.Networks)
	}
	if len(got.Nodes) != 1 || len(got.Nodes[0].NetworkAliases) != 1 || got.Nodes[0].NetworkAliases[0] != "awd-c8-t15-s21" {
		t.Fatalf("expected AWD network aliases to be preserved, got %+v", got.Nodes)
	}
	if len(got.Nodes[0].Command) != 3 || got.Nodes[0].Command[0] != "tail" || got.Nodes[0].WorkingDir != "/workspace" {
		t.Fatalf("expected runtime command and working dir to be preserved, got %+v", got.Nodes[0])
	}
	if len(got.Nodes[0].Mounts) != 1 || got.Nodes[0].Mounts[0].Target != "/workspace/src" {
		t.Fatalf("expected runtime mounts to be preserved, got %+v", got.Nodes[0].Mounts)
	}
	if got.ContainerName != "ctf-workspace-workspace-c8-t15-s21-r1" {
		t.Fatalf("expected container name to be preserved, got %+v", got)
	}
}

func TestRuntimePracticeTopologyAdapterPreservesWorkspaceShellFields(t *testing.T) {
	req := &practiceports.TopologyCreateRequest{
		ContainerName: "workspace-companion",
		Nodes: []practiceports.TopologyCreateNode{
			{
				Key:             "workspace",
				Image:           "python:3.12-alpine",
				Env:             map[string]string{"LANG": "C.UTF-8"},
				Command:         []string{"/bin/sh", "-lc", "apk add --no-cache git vim nano && exec tail -f /dev/null"},
				WorkingDir:      "/workspace",
				ServicePort:     22,
				ServiceProtocol: model.ChallengeTargetProtocolTCP,
				IsEntryPoint:    true,
				NetworkKeys:     []string{model.TopologyDefaultNetworkKey},
				NetworkAliases:  []string{"awd-c8-t15-s21-workspace"},
				Mounts: []model.ContainerMount{
					{Source: "workspace-src", Target: "/workspace/src"},
					{Source: "workspace-data", Target: "/workspace/data", ReadOnly: true},
				},
			},
		},
	}

	got := toRuntimeTopologyCreateRequest(req)
	if got.ContainerName != "workspace-companion" {
		t.Fatalf("expected container name preserved, got %+v", got)
	}
	if len(got.Nodes) != 1 {
		t.Fatalf("expected one node, got %+v", got.Nodes)
	}
	node := got.Nodes[0]
	if !reflect.DeepEqual(node.Command, req.Nodes[0].Command) {
		t.Fatalf("expected command preserved, got %+v", node.Command)
	}
	if node.WorkingDir != req.Nodes[0].WorkingDir {
		t.Fatalf("expected working dir preserved, got %q", node.WorkingDir)
	}
	if !reflect.DeepEqual(node.Mounts, req.Nodes[0].Mounts) {
		t.Fatalf("expected mounts preserved, got %+v", node.Mounts)
	}
	if !reflect.DeepEqual(node.Env, req.Nodes[0].Env) {
		t.Fatalf("expected env preserved, got %+v", node.Env)
	}
}

func TestRuntimePracticeServiceAdapterInspectManagedContainerDelegatesToEngine(t *testing.T) {
	adapter := newRuntimePracticeServiceAdapter(nil, nil, &stubRuntimePracticeEngine{
		inspectFn: func(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
			if containerID != "workspace-ctr" {
				t.Fatalf("unexpected container inspect target: %s", containerID)
			}
			return &runtimeports.ManagedContainerState{
				ID:      containerID,
				Exists:  true,
				Running: true,
				Status:  "running",
			}, nil
		},
	})

	got, err := adapter.InspectManagedContainer(context.Background(), "workspace-ctr")
	if err != nil {
		t.Fatalf("InspectManagedContainer() error = %v", err)
	}
	if got == nil {
		t.Fatal("expected managed container state")
	}
	if got.ID != "workspace-ctr" || !got.Exists || !got.Running || got.Status != "running" {
		t.Fatalf("unexpected managed container state: %+v", got)
	}
}

func TestRuntimePracticeServiceAdapterInspectManagedContainerPropagatesErrors(t *testing.T) {
	wantErr := errors.New("inspect failed")
	adapter := newRuntimePracticeServiceAdapter(nil, nil, &stubRuntimePracticeEngine{
		inspectFn: func(context.Context, string) (*runtimeports.ManagedContainerState, error) {
			return nil, wantErr
		},
	})

	got, err := adapter.InspectManagedContainer(context.Background(), "workspace-ctr")
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected inspect error %v, got %v", wantErr, err)
	}
	if got != nil {
		t.Fatalf("expected nil managed container state on error, got %+v", got)
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
	return s.files[filePath], nil
}

func (s stubRuntimeHTTPDefenseWorkbench) ListDirectoryFromContainer(context.Context, string, string, int) ([]runtimeports.ContainerDirectoryEntry, error) {
	return s.entries, nil
}

func (s stubRuntimeHTTPDefenseWorkbench) WriteFileToContainer(context.Context, string, string, []byte) error {
	return nil
}

func (s stubRuntimeHTTPDefenseWorkbench) ExecContainerCommand(context.Context, string, []string, []byte, int64) ([]byte, error) {
	return nil, nil
}

type recordingRuntimeHTTPDefenseWorkbench struct {
	fileContent []byte
	readPath    string
	listPath    string
	writePaths  []string
}

func (s *recordingRuntimeHTTPDefenseWorkbench) ReadFileFromContainer(_ context.Context, _ string, filePath string, _ int64) ([]byte, error) {
	s.readPath = filePath
	return s.fileContent, nil
}

func (s *recordingRuntimeHTTPDefenseWorkbench) ListDirectoryFromContainer(_ context.Context, _ string, dirPath string, _ int) ([]runtimeports.ContainerDirectoryEntry, error) {
	s.listPath = dirPath
	return nil, nil
}

func (s *recordingRuntimeHTTPDefenseWorkbench) WriteFileToContainer(_ context.Context, _ string, filePath string, _ []byte) error {
	s.writePaths = append(s.writePaths, filePath)
	return nil
}

func (s *recordingRuntimeHTTPDefenseWorkbench) ExecContainerCommand(context.Context, string, []string, []byte, int64) ([]byte, error) {
	return nil, nil
}

type stubRuntimePracticeEngine struct {
	Engine
	inspectFn func(context.Context, string) (*runtimeports.ManagedContainerState, error)
}

func (s *stubRuntimePracticeEngine) InspectManagedContainer(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
	if s.inspectFn == nil {
		if s.Engine == nil {
			return nil, nil
		}
		return s.Engine.InspectManagedContainer(ctx, containerID)
	}
	return s.inspectFn(ctx, containerID)
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
