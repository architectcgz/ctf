package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestConfigPackageSplitByResponsibility(t *testing.T) {
	t.Parallel()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}

	configDir := filepath.Dir(file)
	requiredFiles := []string{
		"types.go",
		"load.go",
		"defaults.go",
		"validate.go",
		"container_flag_secret.go",
	}

	for _, name := range requiredFiles {
		path := filepath.Join(configDir, name)
		if _, err := os.Stat(path); err != nil {
			if os.IsNotExist(err) {
				t.Fatalf("expected config package to keep responsibility split file %s", name)
			}
			t.Fatalf("stat %s: %v", name, err)
		}
	}
}
