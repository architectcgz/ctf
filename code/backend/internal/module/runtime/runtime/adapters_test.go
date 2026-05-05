package runtime

import (
	"context"
	"testing"
	"time"

	"ctf-platform/internal/authctx"
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
		stubRuntimeHTTPProxyTicketReader{scope: &runtimeports.AWDDefenseSSHScope{ContainerID: "container-12"}},
		stubRuntimeHTTPDefenseWorkbench{
			files: map[string][]byte{
				"/home/student/app.py": []byte("print('vuln')"),
			},
			entries: []runtimeports.ContainerDirectoryEntry{
				{Name: "app.py", Type: "file", Size: 13},
				{Name: ".env", Type: "file", Size: 10},
				{Name: ".ssh", Type: "dir"},
			},
		},
		0,
		false,
		"",
		0,
		true,
		"/home/student",
	)

	fileResp, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "app.py")
	if err != nil {
		t.Fatalf("ReadAWDDefenseFile() error = %v", err)
	}
	if fileResp.Path != "app.py" || fileResp.Content != "print('vuln')" {
		t.Fatalf("unexpected file response: %+v", fileResp)
	}

	dirResp, err := adapter.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, ".")
	if err != nil {
		t.Fatalf("ListAWDDefenseDirectory() error = %v", err)
	}
	if len(dirResp.Entries) != 1 || dirResp.Entries[0].Path != "app.py" {
		t.Fatalf("unexpected directory response: %+v", dirResp)
	}
}

func TestRuntimeHTTPServiceAdapterRejectsAWDDefenseWorkbenchWhenDisabled(t *testing.T) {
	adapter := newRuntimeHTTPServiceAdapter(
		nil,
		nil,
		nil,
		stubRuntimeHTTPProxyTicketReader{scope: &runtimeports.AWDDefenseSSHScope{ContainerID: "container-12"}},
		stubRuntimeHTTPDefenseWorkbench{},
		0,
		false,
		"",
		0,
		false,
		"/home/student",
	)

	if _, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "app.py"); err == nil {
		t.Fatal("expected disabled workbench read to fail")
	}
	if _, err := adapter.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "."); err == nil {
		t.Fatal("expected disabled workbench list to fail")
	}
}

func TestRuntimeHTTPServiceAdapterRejectsAWDDefenseSensitivePaths(t *testing.T) {
	adapter := newRuntimeHTTPServiceAdapter(
		nil,
		nil,
		nil,
		stubRuntimeHTTPProxyTicketReader{scope: &runtimeports.AWDDefenseSSHScope{ContainerID: "container-12"}},
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
	if _, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "../app.py"); err == nil {
		t.Fatal("expected traversal path to fail")
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
		stubRuntimeHTTPProxyTicketReader{scope: &runtimeports.AWDDefenseSSHScope{ContainerID: "container-12"}},
		workbench,
		0,
		false,
		"",
		0,
		true,
		"/home/student",
	)

	if _, err := adapter.ReadAWDDefenseFile(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "src/app.py"); err != nil {
		t.Fatalf("ReadAWDDefenseFile() error = %v", err)
	}
	if workbench.readPath != "/home/student/src/app.py" {
		t.Fatalf("expected rooted read path, got %q", workbench.readPath)
	}

	if _, err := adapter.ListAWDDefenseDirectory(context.Background(), authctx.CurrentUser{UserID: 1001}, 5, 12, "."); err != nil {
		t.Fatalf("ListAWDDefenseDirectory() error = %v", err)
	}
	if workbench.listPath != "/home/student" {
		t.Fatalf("expected rooted list path, got %q", workbench.listPath)
	}
}

func TestRuntimePracticeTopologyAdapterPreservesAWDNetworkFields(t *testing.T) {
	req := &practiceports.TopologyCreateRequest{
		Networks: []practiceports.TopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-8", Shared: true},
		},
		Nodes: []practiceports.TopologyCreateNode{
			{
				Key:            "web",
				IsEntryPoint:   true,
				NetworkKeys:    []string{model.TopologyDefaultNetworkKey},
				NetworkAliases: []string{"awd-c8-t15-s21"},
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
}

func (s *recordingRuntimeHTTPDefenseWorkbench) ReadFileFromContainer(_ context.Context, _ string, filePath string, _ int64) ([]byte, error) {
	s.readPath = filePath
	return s.fileContent, nil
}

func (s *recordingRuntimeHTTPDefenseWorkbench) ListDirectoryFromContainer(_ context.Context, _ string, dirPath string, _ int) ([]runtimeports.ContainerDirectoryEntry, error) {
	s.listPath = dirPath
	return nil, nil
}

func (s *recordingRuntimeHTTPDefenseWorkbench) WriteFileToContainer(context.Context, string, string, []byte) error {
	return nil
}

func (s *recordingRuntimeHTTPDefenseWorkbench) ExecContainerCommand(context.Context, string, []string, []byte, int64) ([]byte, error) {
	return nil, nil
}
