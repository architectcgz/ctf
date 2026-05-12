package composition

import (
	"go/parser"
	"go/token"
	"os"
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

func TestCompositionDoesNotReintroduceRuntimeCompatFacade(t *testing.T) {
	t.Parallel()

	if _, err := os.Stat("runtime_adapter_compat.go"); err == nil {
		t.Fatal("composition must not reintroduce runtime_adapter_compat.go")
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat runtime_adapter_compat.go: %v", err)
	}
}

func TestInstanceModuleDoesNotInjectRetiredDefenseWorkbenchService(t *testing.T) {
	t.Parallel()

	assertFileDoesNotContain(t, "instance_module.go", "AWDDefenseWorkbenchService")
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

func assertFileDoesNotContain(t *testing.T, filePath string, blockedSnippet string) {
	t.Helper()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read file %s: %v", filePath, err)
	}
	if strings.Contains(string(content), blockedSnippet) {
		t.Fatalf("%s must not contain %q", filePath, blockedSnippet)
	}
}
