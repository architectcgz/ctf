package commands

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegistryClientCheckManifestReturnsDigest(t *testing.T) {
	var sawAuth bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Fatalf("method = %s, want HEAD", r.Method)
		}
		if r.URL.Path != "/v2/jeopardy/web-demo/manifests/v1" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		username, password, ok := r.BasicAuth()
		if ok && username == "ctf" && password == "secret" {
			sawAuth = true
		}
		w.Header().Set("Docker-Content-Digest", "sha256:demo")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	client := NewRegistryClient(RegistryClientConfig{
		Scheme:   "http",
		Server:   host,
		Username: "ctf",
		Password: "secret",
	}, server.Client())

	digest, err := client.CheckManifest(context.Background(), host+"/jeopardy/web-demo:v1")
	if err != nil {
		t.Fatalf("CheckManifest() error = %v", err)
	}
	if digest != "sha256:demo" {
		t.Fatalf("digest = %q, want sha256:demo", digest)
	}
	if !sawAuth {
		t.Fatal("expected basic auth")
	}
}

func TestRegistryClientCheckManifestRejectsMismatchedRegistry(t *testing.T) {
	client := NewRegistryClient(RegistryClientConfig{Scheme: "http", Server: "registry.example.edu"}, nil)

	if _, err := client.CheckManifest(context.Background(), "docker.io/library/nginx:latest"); err == nil {
		t.Fatal("expected mismatched registry error")
	}
}
