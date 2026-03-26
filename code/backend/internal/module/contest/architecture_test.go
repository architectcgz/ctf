package contest

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
		t.Fatalf("glob contest root files: %v", err)
	}

	nonTestFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		nonTestFiles = append(nonTestFiles, file)
	}

	if len(nonTestFiles) != 0 {
		t.Fatalf("contest root package should keep no non-test go files, got %v", nonTestFiles)
	}
}

func TestAPIHTTPDoesNotDependOnInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("api", "http", "*.go"))
	if err != nil {
		t.Fatalf("glob api/http files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/infrastructure")
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
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/queries")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/jobs")
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
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/commands")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/jobs")
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
	}
}

func TestJobsDoNotDependOnAPIHTTPOrInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("application", "jobs", "*.go"))
	if err != nil {
		t.Fatalf("glob jobs files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/commands")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/queries")
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
	}
}

func TestPortsDoNotDependOnGinOrGORM(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("ports", "*.go"))
	if err != nil {
		t.Fatalf("glob ports files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/commands")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/queries")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/jobs")
	}
}

func TestPortsDoNotDeclareWideRepository(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("ports", "contest.go"))
	if err != nil {
		t.Fatalf("read contest ports file: %v", err)
	}
	if strings.Contains(string(content), "type Repository interface") {
		t.Fatalf("contest ports must not declare the legacy wide Repository interface")
	}
}

func TestDomainDoesNotDependOnGinGORMOrRedis(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("domain", "*.go"))
	if err != nil {
		t.Fatalf("glob domain files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
		assertFileDoesNotImport(t, file, "github.com/redis/go-redis/v9")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/commands")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/queries")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/jobs")
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
