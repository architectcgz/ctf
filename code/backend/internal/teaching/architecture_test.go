package teaching

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
	"testing"
)

func TestSharedTeachingPackagesDoNotDependOnModuleContracts(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("*", "*.go"))
	if err != nil {
		t.Fatalf("glob teaching files: %v", err)
	}

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/")
	}
}

func assertFileDoesNotImport(t *testing.T, filePath, blockedPrefix string) {
	t.Helper()

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse %s: %v", filePath, err)
	}

	for _, spec := range node.Imports {
		path := strings.Trim(spec.Path.Value, "\"")
		if strings.HasPrefix(path, blockedPrefix) {
			t.Fatalf("%s must not import %s, got %s", filePath, blockedPrefix, path)
		}
	}
}
