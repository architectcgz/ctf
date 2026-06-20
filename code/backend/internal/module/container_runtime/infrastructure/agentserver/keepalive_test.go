package agentserver

import (
	"testing"
	"time"

	"ctf-platform/internal/config"
)

func TestRuntimeAgentServerAllowsIdleClientKeepalive(t *testing.T) {
	policy := runtimeAgentKeepaliveEnforcementPolicy(config.RuntimeAgentServerConfig{
		KeepaliveMinTime: 45 * time.Second,
	})

	if policy.MinTime != 45*time.Second {
		t.Fatalf("keepalive min time = %s, want 45s", policy.MinTime)
	}
	if !policy.PermitWithoutStream {
		t.Fatal("expected runtime-agent server to allow idle client keepalive pings")
	}
}
