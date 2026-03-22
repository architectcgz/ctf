package composition

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCompositionAndRuntimeDoNotImportLegacyContainerModule(t *testing.T) {
	t.Parallel()

	assertDirDoesNotImport(t, ".", "ctf-platform/internal/module/container")
	assertDirDoesNotImport(t, filepath.Join("..", "..", "module", "runtime"), "ctf-platform/internal/module/container")
}

func assertDirDoesNotImport(t *testing.T, dirPath string, blockedImport string) {
	t.Helper()

	files, err := filepath.Glob(filepath.Join(dirPath, "*.go"))
	if err != nil {
		t.Fatalf("glob files in %s: %v", dirPath, err)
	}

	for _, filePath := range files {
		if strings.HasSuffix(filePath, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, filePath, blockedImport)
	}
}

func assertFileDoesNotImport(t *testing.T, filePath string, blockedImport string) {
	t.Helper()

	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse file %s: %v", filePath, err)
	}

	for _, importSpec := range fileNode.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import %s: %v", importSpec.Path.Value, err)
		}
		if importPath == blockedImport {
			t.Fatalf("%s must not import %s", filePath, blockedImport)
		}
	}
}
