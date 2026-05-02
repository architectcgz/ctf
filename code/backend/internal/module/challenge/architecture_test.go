package challenge

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
		t.Fatalf("glob challenge root files: %v", err)
	}

	nonTestFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		nonTestFiles = append(nonTestFiles, file)
	}

	if len(nonTestFiles) != 0 {
		t.Fatalf("challenge root package should keep no non-test go files, got %v", nonTestFiles)
	}
}

func TestAPIHTTPDoesNotDependOnInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("api", "http", "*.go"))
	if err != nil {
		t.Fatalf("glob api/http files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/challenge/infrastructure")
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
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/challenge/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/challenge/infrastructure")
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
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/challenge/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/challenge/infrastructure")
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

func TestPortsDoNotDeclareWideChallengeRepository(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("ports", "ports.go"))
	if err != nil {
		t.Fatalf("read challenge ports file: %v", err)
	}
	if strings.Contains(string(content), "type ChallengeRepository interface") {
		t.Fatalf("challenge ports must not declare the legacy wide ChallengeRepository interface")
	}
}

func TestRuntimeOwnsChallengeWiring(t *testing.T) {
	t.Parallel()

	runtimeFile := filepath.Join("runtime", "module.go")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/challenge/infrastructure")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/challenge/application/commands")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/challenge/application/queries")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/challenge/api/http")
}

func TestRuntimeUsesTypedPortsDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("runtime", "module.go"))
	if err != nil {
		t.Fatalf("read challenge runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"type moduleDeps struct",
		"challengeCommandRepo    challengeports.ChallengeCommandRepository",
		"challengeQueryRepo      challengeports.ChallengeQueryRepository",
		"flagRepo                challengeports.ChallengeFlagRepository",
		"imageUsageRepo          challengeports.ChallengeImageUsageRepository",
		"topologyRepo            challengeports.ChallengeTopologyRepository",
		"writeupRepo             challengeports.ChallengeWriteupRepository",
		"templateRepo            challengeports.EnvironmentTemplateRepository",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("challenge runtime should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"challengeRepo *challengeinfra.Repository",
		"imageRepo *challengeinfra.ImageRepository",
		"templateRepo *challengeinfra.TemplateRepository",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("challenge runtime should not keep concrete repository field %s", marker)
		}
	}
}

func TestRuntimeDelegatesThroughSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("runtime", "module.go"))
	if err != nil {
		t.Fatalf("read challenge runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"buildImageHandler(",
		"buildCoreHandler(",
		"buildFlagHandler(",
		"buildTopologyHandler(",
		"buildWriteupHandler(",
		"buildAWDChallengeHandler(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("challenge runtime should delegate through %s", marker)
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
