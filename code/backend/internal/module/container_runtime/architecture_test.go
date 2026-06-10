package container_runtime

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRuntimePackageDoesNotDependOnBusinessOwnerPorts(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("runtime", "*.go"))
	if err != nil {
		t.Fatalf("glob container_runtime runtime files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/ports")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/practice/ports")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/instance/ports")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/contracts")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/practice/contracts")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/instance/contracts")
	}
}

func TestRuntimePackageDoesNotImportTransportOrPersistenceConcreteTypes(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("runtime", "*.go"))
	if err != nil {
		t.Fatalf("glob container_runtime runtime files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImportMatching(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImportMatching(t, file, "github.com/redis/go-redis/v9")
		assertFileDoesNotImportMatching(t, file, "github.com/docker/docker")
		assertFileDoesNotImportMatching(t, file, "gorm.io/gorm")
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

func assertFileDoesNotImportMatching(t *testing.T, filePath string, blockedImport string) {
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
		if importPath == blockedImport || strings.HasPrefix(importPath, blockedImport+"/") {
			t.Fatalf("%s must not import %s", filePath, importPath)
		}
	}
}
