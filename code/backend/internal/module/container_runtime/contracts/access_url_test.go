package contracts

import "testing"

func TestResolveRuntimeNodeAccessHostPrefersNodeHostFallbacks(t *testing.T) {
	t.Parallel()

	if got := ResolveRuntimeNodeAccessHost("node-public.local", "node-access.internal", "global-public.local", "global-access.internal"); got != "node-access.internal" {
		t.Fatalf("node access host = %q, want node access", got)
	}
	if got := ResolveRuntimeNodeAccessHost("node-public.local", "", "global-public.local", "global-access.internal"); got != "node-public.local" {
		t.Fatalf("node public host fallback = %q, want node public", got)
	}
	if got := ResolveRuntimeNodeAccessHost("", "", "global-public.local", "global-access.internal"); got != "global-access.internal" {
		t.Fatalf("global access host fallback = %q, want global access", got)
	}
}

func TestResolveRuntimeNodePublicHostFallsBackToGlobalPublicHost(t *testing.T) {
	t.Parallel()

	if got := ResolveRuntimeNodePublicHost("node-public.local", "global-public.local"); got != "node-public.local" {
		t.Fatalf("node public host = %q, want node public", got)
	}
	if got := ResolveRuntimeNodePublicHost("", "global-public.local"); got != "global-public.local" {
		t.Fatalf("global public fallback = %q, want global public", got)
	}
}

func TestResolveRuntimePublicAccessURLUsesNodePublicHost(t *testing.T) {
	t.Parallel()

	nodePublic := ResolveRuntimeNodePublicHost("node-b.ctf.local", "global-public.local")
	nodeAccess := ResolveRuntimeNodeAccessHost("node-b.ctf.local", "node-b.internal", "global-public.local", "global-access.internal")
	got := ResolveRuntimePublicAccessURL("http://node-b.internal:30000", nodePublic, nodeAccess)
	if got != "http://node-b.ctf.local:30000" {
		t.Fatalf("node public access url = %q, want node public URL", got)
	}
}
