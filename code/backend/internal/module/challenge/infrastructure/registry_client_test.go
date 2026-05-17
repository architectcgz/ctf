package infrastructure

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

func TestRegistryClientCheckManifestAcceptsOCIIndexResponses(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); !strings.Contains(got, "application/vnd.oci.image.index.v1+json") {
			t.Fatalf("accept header missing oci index: %q", got)
		}
		if got := r.Header.Get("Accept"); !strings.Contains(got, "application/vnd.docker.distribution.manifest.list.v2+json") {
			t.Fatalf("accept header missing docker manifest list: %q", got)
		}
		w.Header().Set("Docker-Content-Digest", "sha256:index")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	host := strings.TrimPrefix(server.URL, "http://")
	client := NewRegistryClient(RegistryClientConfig{
		Scheme: "http",
		Server: host,
	}, server.Client())

	digest, err := client.CheckManifest(context.Background(), host+"/jeopardy/web-demo:v1")
	if err != nil {
		t.Fatalf("CheckManifest() error = %v", err)
	}
	if digest != "sha256:index" {
		t.Fatalf("digest = %q, want sha256:index", digest)
	}
}

func TestRegistryClientCheckManifestUsesAccessServerOverride(t *testing.T) {
	var sawPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawPath = r.URL.Path
		w.Header().Set("Docker-Content-Digest", "sha256:override")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewRegistryClient(RegistryClientConfig{
		Scheme:       "http",
		Server:       "127.0.0.1:5000",
		AccessServer: strings.TrimPrefix(server.URL, "http://"),
	}, server.Client())

	digest, err := client.CheckManifest(context.Background(), "127.0.0.1:5000/jeopardy/web-demo:v1")
	if err != nil {
		t.Fatalf("CheckManifest() error = %v", err)
	}
	if digest != "sha256:override" {
		t.Fatalf("digest = %q, want sha256:override", digest)
	}
	if sawPath != "/v2/jeopardy/web-demo/manifests/v1" {
		t.Fatalf("path = %s", sawPath)
	}
}

func TestRegistryClientCheckManifestRejectsMismatchedRegistry(t *testing.T) {
	client := NewRegistryClient(RegistryClientConfig{Scheme: "http", Server: "registry.example.edu"}, nil)

	if _, err := client.CheckManifest(context.Background(), "docker.io/library/nginx:latest"); err == nil {
		t.Fatal("expected mismatched registry error")
	}
}
