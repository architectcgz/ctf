package agentclient

import (
	"testing"
	"time"

	"ctf-platform/internal/config"
)

func TestRuntimeAgentClientKeepalivePingsIdleConnections(t *testing.T) {
	params := runtimeAgentClientKeepaliveParameters(config.RuntimeAgentConfig{
		KeepaliveTime:    45 * time.Second,
		KeepaliveTimeout: 7 * time.Second,
	})

	if params.Time != 45*time.Second {
		t.Fatalf("keepalive time = %s, want 45s", params.Time)
	}
	if params.Timeout != 7*time.Second {
		t.Fatalf("keepalive timeout = %s, want 7s", params.Timeout)
	}
	if !params.PermitWithoutStream {
		t.Fatal("expected runtime-agent client keepalive to ping idle connections")
	}
}
