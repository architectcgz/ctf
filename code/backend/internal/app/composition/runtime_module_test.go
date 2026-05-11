package composition

import (
	"context"
	"ctf-platform/internal/authctx"
	"net/http"
	"testing"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/config"
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
