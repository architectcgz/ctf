package identity

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRootPackageKeepsNoConcreteGoFiles(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob identity root files: %v", err)
	}

	nonTestFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		nonTestFiles = append(nonTestFiles, file)
	}

	if len(nonTestFiles) != 0 {
		t.Fatalf("identity root package should keep no non-test go files, got %v", nonTestFiles)
	}
}

func TestAPIHTTPDoesNotDependOnInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("api", "http", "*.go"))
	if err != nil {
		t.Fatalf("glob api/http files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/identity/infrastructure")
	}
}

func TestCommandsDoNotDependOnAPIHTTPOrInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("application", "commands", "*.go"))
	if err != nil {
		t.Fatalf("glob commands files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/identity/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/identity/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/identity/application/queries")
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
	}
}

func TestQueriesDoNotDependOnAPIHTTPOrInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("application", "queries", "*.go"))
	if err != nil {
		t.Fatalf("glob queries files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/identity/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/identity/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/identity/application/commands")
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
	}
}

func TestContractsDoNotDependOnGinGORMOrConcreteLayers(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("contracts", "*.go"))
	if err != nil {
		t.Fatalf("glob contracts files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/identity/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/identity/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/identity/application/commands")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/identity/application/queries")
	}
}

func TestContractsDoNotDeclareWideUserRepository(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("contracts", "auth.go"))
	if err != nil {
		t.Fatalf("read identity auth contracts file: %v", err)
	}
	if strings.Contains(string(content), "type UserRepository interface") {
		t.Fatalf("identity contracts must not declare the legacy wide UserRepository interface")
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
