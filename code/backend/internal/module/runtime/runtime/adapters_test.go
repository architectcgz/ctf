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
