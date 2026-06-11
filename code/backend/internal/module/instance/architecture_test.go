package instance

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestCommandsDoNotDependOnHTTPOrRuntimeInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("application", "commands", "*.go"))
	if err != nil {
		t.Fatalf("glob application command files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/infrastructure")
	}
}

func TestStartupRuntimeRecoveryInfrastructureBoundary(t *testing.T) {
	t.Parallel()

	file := filepath.Join("application", "commands", "startup_runtime_recovery_service.go")
	assertFileDoesNotImport(t, file, "os")
	assertFileDoesNotImport(t, file, "ctf-platform/internal/infrastructure/redislock")

	source := readFileSource(t, file)
	blockedMarkers := []string{
		"os.ReadFile",
		"bootIDPath",
		"*redislock.Lock",
		"defaultBootIDPath",
	}
	for _, marker := range blockedMarkers {
		if strings.Contains(source, marker) {
			t.Fatalf("startup recovery commands must not own infrastructure marker %s", marker)
		}
	}
}

func TestQueriesDoNotDependOnHTTPOrRuntimeInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("application", "queries", "*.go"))
	if err != nil {
		t.Fatalf("glob application query files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/infrastructure")
	}
}

func TestDomainDoesNotDependOnGinGORMOrRuntimeInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("domain", "*.go"))
	if err != nil {
		t.Fatalf("glob domain files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
	}
}

func TestRootPackageKeepsOnlyDocFile(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob instance root files: %v", err)
	}

	nonTestFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		nonTestFiles = append(nonTestFiles, file)
	}

	if len(nonTestFiles) != 1 || nonTestFiles[0] != "doc.go" {
		t.Fatalf("instance root package should keep only doc.go, got %v", nonTestFiles)
	}
}

func TestProductionInstanceDoesNotImportContestModule(t *testing.T) {
	t.Parallel()

	files := instanceProductionGoFiles(t)
	for _, file := range files {
		assertFileDoesNotImportMatching(t, file, "ctf-platform/internal/module/contest")
	}
}

func instanceProductionGoFiles(t *testing.T) []string {
	t.Helper()

	files := make([]string, 0)
	err := filepath.WalkDir(".", func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case "data", "testdata", "testsupport":
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, ".go") && !strings.HasSuffix(path, "_test.go") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk instance production go files: %v", err)
	}
	return files
}

func readFileSource(t *testing.T, filePath string) string {
	t.Helper()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read file %s: %v", filePath, err)
	}
	return string(content)
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

func assertFileDoesNotImportMatching(t *testing.T, filePath string, blockedImportPrefix string) {
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
		if importPath == blockedImportPrefix || strings.HasPrefix(importPath, blockedImportPrefix+"/") {
			t.Fatalf("%s must not import %s", filePath, blockedImportPrefix)
		}
	}
}
