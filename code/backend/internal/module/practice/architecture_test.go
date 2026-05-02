package practice

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
		t.Fatalf("glob practice root files: %v", err)
	}

	nonTestFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		nonTestFiles = append(nonTestFiles, file)
	}

	if len(nonTestFiles) != 0 {
		t.Fatalf("practice root package should keep no non-test go files, got %v", nonTestFiles)
	}
}

func TestAPIHTTPDoesNotDependOnInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("api", "http", "*.go"))
	if err != nil {
		t.Fatalf("glob api/http files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/practice/infrastructure")
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
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/practice/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/practice/infrastructure")
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
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/practice/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/practice/infrastructure")
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
	}
}

func TestPortsDoNotDeclareWidePracticeRepository(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("ports", "ports.go"))
	if err != nil {
		t.Fatalf("read practice ports file: %v", err)
	}
	if strings.Contains(string(content), "type PracticeRepository interface") {
		t.Fatalf("practice ports must not declare the legacy wide PracticeRepository interface")
	}
}

func TestRuntimeOwnsPracticeWiring(t *testing.T) {
	t.Parallel()

	runtimeFile := filepath.Join("runtime", "module.go")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/practice/infrastructure")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/practice/application/commands")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/practice/application/queries")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/practice/api/http")
}

func TestRuntimeUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("runtime", "module.go"))
	if err != nil {
		t.Fatalf("read practice runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"type moduleDeps struct",
		"commandRepo    practiceports.PracticeCommandRepository",
		"scoreRepo      practiceports.PracticeScoreRepository",
		"rankingRepo    practiceports.PracticeRankingRepository",
		"instanceRepo   practiceports.InstanceRepository",
		"runtimeService practiceports.RuntimeInstanceService",
		"challengeRepo  challengecontracts.PracticeChallengeContract",
		"imageStore     challengecontracts.ImageStore",
		"assessment     assessmentcontracts.ProfileService",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("practice runtime should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"repo *practiceinfra.Repository",
		"type practiceModuleDeps struct",
		"type practiceModuleExternalDeps struct",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("practice runtime should not keep composition glue marker %s", marker)
		}
	}
}

func TestRuntimeDelegatesThroughSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("runtime", "module.go"))
	if err != nil {
		t.Fatalf("read practice runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"newModuleDeps(",
		"buildHandler(",
		"practicehttp.NewHandler(",
		"service.StartBackgroundTasks(",
		"service.SetEventBus(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("practice runtime should delegate through %s", marker)
		}
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
	}
}

func assertFileImports(t *testing.T, filePath string, expectedImport string) {
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
		if importPath == expectedImport {
			return
		}
	}

	t.Fatalf("%s must import %s", filePath, expectedImport)
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
