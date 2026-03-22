package runtime

import (
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"testing"
)

func TestApplicationDoesNotDependOnHTTPOrRuntimeInfra(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("application", "*.go"))
	if err != nil {
		t.Fatalf("glob application files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "github.com/redis/go-redis/v9")
		assertFileDoesNotImport(t, file, "ctf-platform/pkg/response")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtimeinfra")
	}
}

func TestDomainDoesNotDependOnGinGORMOrRuntimeInfra(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("domain", "*.go"))
	if err != nil {
		t.Fatalf("glob domain files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtimeinfra")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
	}
}

func TestAPIHTTPDoesNotDependOnGORMOrRuntimeInfra(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("api", "http", "*.go"))
	if err != nil {
		t.Fatalf("glob api/http files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtimeinfra")
	}
}

func TestInfrastructureDoesNotDependOnDTOOrGin(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("infrastructure", "*.go"))
	if err != nil {
		t.Fatalf("glob infrastructure files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
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
