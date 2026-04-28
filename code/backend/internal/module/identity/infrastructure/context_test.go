package infrastructure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInfrastructureDoesNotCreateBackgroundContext(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob go files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		source, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		if strings.Contains(string(source), "context.Background()") {
			t.Fatalf("%s should not create context.Background()", file)
		}
	}
}
