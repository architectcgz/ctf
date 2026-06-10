package infrastructure

import (
	"context"
	"os"
	"testing"
	"time"

	"ctf-platform/internal/config"
)

func TestEnginePullsImageFromPrivateRegistryWithConfiguredAuth(t *testing.T) {
	imageRef := os.Getenv("CTF_TEST_PRIVATE_REGISTRY_IMAGE")
	server := os.Getenv("CTF_TEST_PRIVATE_REGISTRY_SERVER")
	username := os.Getenv("CTF_TEST_PRIVATE_REGISTRY_USERNAME")
	password := os.Getenv("CTF_TEST_PRIVATE_REGISTRY_PASSWORD")
	if imageRef == "" || server == "" || username == "" || password == "" {
		t.Skip("set CTF_TEST_PRIVATE_REGISTRY_IMAGE/SERVER/USERNAME/PASSWORD to run private registry integration test")
	}

	engine, err := NewEngine(&config.ContainerConfig{
		DefaultCPUQuota:  0.5,
		DefaultMemory:    128 * 1024 * 1024,
		DefaultPidsLimit: 128,
		Registry: config.ContainerRegistryConfig{
			Enabled:  true,
			Server:   server,
			Username: username,
			Password: password,
		},
	})
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	port, err := engine.ResolveServicePort(ctx, imageRef, 8080)
	if err != nil {
		t.Fatalf("ResolveServicePort() error = %v", err)
	}
	if port != 8080 {
		t.Fatalf("ResolveServicePort() = %d, want fallback port 8080", port)
	}
}
