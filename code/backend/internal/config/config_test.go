package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadReadsContainerFlagSecretFromEnv(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}

	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(backendRoot); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(currentDir)
	})

	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", "integration-secret")

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Container.FlagGlobalSecret != "integration-secret" {
		t.Fatalf("expected container flag secret from env, got %q", cfg.Container.FlagGlobalSecret)
	}
}
