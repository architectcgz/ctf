package ports

import (
	"os"
	"strings"
	"testing"
)

func TestPortsDoNotExposeWithContextNames(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("ports.go")
	if err != nil {
		t.Fatalf("read ports.go: %v", err)
	}
	if strings.Contains(string(source), "WithContext") {
		t.Fatal("assessment ports should use Foo(ctx, ...) names, not FooWithContext(ctx, ...)")
	}
}
