package contracts

import "testing"

func TestResolveInstancePublicAccessURL(t *testing.T) {
	t.Parallel()

	if got := ResolveInstancePublicAccessURL("tcp://internal.local:32001", "public.example", "internal.local"); got != "tcp://public.example:32001" {
		t.Fatalf("expected public host rewrite, got %q", got)
	}
	if got := ResolveInstancePublicAccessURL("tcp://other.local:32001", "public.example", "internal.local"); got != "tcp://other.local:32001" {
		t.Fatalf("expected non-matching host to stay unchanged, got %q", got)
	}
	if got := ResolveInstancePublicAccessURL("tcp://internal.local:32001", "public.example", ""); got != "tcp://internal.local:32001" {
		t.Fatalf("expected empty access host to stay unchanged, got %q", got)
	}
}

func TestResolveInstanceAliasAccessURLUsesEntrypointRuntimeIP(t *testing.T) {
	t.Parallel()

	runtimeDetails := `{
		"networks":[{"key":"default","name":"ctf-awd-contest-8"}],
		"containers":[
			{"container_id":"sidecar","network_keys":["default"],"network_ips":{"ctf-awd-contest-8":"172.30.0.10"}},
			{"container_id":"entry","is_entry_point":true,"network_keys":["default"],"network_ips":{"ctf-awd-contest-8":"172.30.0.20"}}
		]
	}`
	got := ResolveInstanceAliasAccessURL("http://awd-c8-t15-s21:8080", runtimeDetails)
	if got != "http://172.30.0.20:8080" {
		t.Fatalf("expected entrypoint ip rewrite, got %q", got)
	}
}

func TestResolveInstanceAliasAccessURLLeavesNonAliasHost(t *testing.T) {
	t.Parallel()

	runtimeDetails := `{"containers":[{"is_entry_point":true,"network_ips":{"default":"172.30.0.20"}}]}`
	got := ResolveInstanceAliasAccessURL("http://127.0.0.1:8080", runtimeDetails)
	if got != "http://127.0.0.1:8080" {
		t.Fatalf("expected non-alias host to stay unchanged, got %q", got)
	}
}

func TestExtractInstanceRuntimeContainerIDs(t *testing.T) {
	t.Parallel()

	got, err := ExtractInstanceRuntimeContainerIDs(`{
		"containers":[
			{"container_id":"main"},
			{"container_id":"sidecar"},
			{"container_id":"main"},
			{"container_id":""}
		]
	}`)
	if err != nil {
		t.Fatalf("ExtractInstanceRuntimeContainerIDs() error = %v", err)
	}
	want := []string{"main", "sidecar"}
	if len(got) != len(want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
	for idx := range want {
		if got[idx] != want[idx] {
			t.Fatalf("expected %v, got %v", want, got)
		}
	}
}
