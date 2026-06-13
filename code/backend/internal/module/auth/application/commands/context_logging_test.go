package commands

import (
	"context"
	"testing"

	"ctf-platform/internal/config"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	"ctf-platform/internal/platform/requestctx"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func TestServiceLoginFailureLogIncludesRequestID(t *testing.T) {
	t.Parallel()

	core, observed := observer.New(zap.InfoLevel)
	logger := zap.New(core)
	service := NewService(&mockRepository{
		findByUsernameFn: func(context.Context, string) (*identitycontracts.User, error) {
			return nil, identitycontracts.ErrUserNotFound
		},
	}, &mockTokenService{}, config.RateLimitPolicyConfig{}, logger)

	ctx := requestctx.WithRequestID(context.Background(), "req-auth-log-1")
	_, _, err := service.Login(ctx, LoginInput{Username: "alice", Password: "wrong"})
	if err == nil {
		t.Fatal("expected login to fail")
	}

	entries := observed.All()
	if len(entries) == 0 {
		t.Fatal("expected login failure log entry")
	}
	fields := entries[len(entries)-1].ContextMap()
	if got := fields["request_id"]; got != "req-auth-log-1" {
		t.Fatalf("request_id = %v, want req-auth-log-1", got)
	}
}
