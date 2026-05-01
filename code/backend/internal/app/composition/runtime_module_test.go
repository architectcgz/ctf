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
	practiceports "ctf-platform/internal/module/practice/ports"
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

func TestRuntimeHTTPServiceAdapterBuildsVSCodeSSHConfig(t *testing.T) {
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
	)

	resp, err := adapter.IssueAWDDefenseSSHTicket(context.Background(), authctx.CurrentUser{
		UserID:   1001,
		Username: "student",
		Role:     "student",
	}, 5, 12)
	if err != nil {
		t.Fatalf("IssueAWDDefenseSSHTicket() error = %v", err)
	}

	if resp.SSHProfile == nil {
		t.Fatal("expected ssh profile in response")
	}
	if resp.SSHProfile.Alias != "ctf-awd-5-12" ||
		resp.SSHProfile.HostName != "ssh.ctf.local" ||
		resp.SSHProfile.Port != 2222 ||
		resp.SSHProfile.User != "student+5+12" {
		t.Fatalf("unexpected ssh profile: %+v", resp.SSHProfile)
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
