package contest

import (
	"go/ast"
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
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
		assertFileDoesNotImport(t, file, "github.com/redis/go-redis/v9")
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

func TestReadinessQueryUsesApplicationResultInsteadOfHTTPDTO(t *testing.T) {
	t.Parallel()

	files := []string{
		filepath.Join("application", "queries", "awd_readiness_query.go"),
		filepath.Join("application", "queries", "awd_readiness_result.go"),
		filepath.Join("domain", "awd_readiness.go"),
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/model")
	}
}

func TestRoundListQueryUsesApplicationResultInsteadOfHTTPDTO(t *testing.T) {
	t.Parallel()

	files := []string{
		filepath.Join("application", "queries", "awd_round_list_query.go"),
		filepath.Join("application", "queries", "awd_round_result.go"),
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/model")
	}
}

func TestAttackLogQueryUsesApplicationResultInsteadOfHTTPDTO(t *testing.T) {
	t.Parallel()

	files := []string{
		filepath.Join("application", "queries", "awd_attack_log_list_query.go"),
		filepath.Join("application", "queries", "awd_attack_log_result.go"),
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
	}
}

func TestTeamServiceQueryUsesApplicationResultInsteadOfHTTPDTO(t *testing.T) {
	t.Parallel()

	files := []string{
		filepath.Join("application", "queries", "awd_service_list_query.go"),
		filepath.Join("application", "queries", "awd_team_service_result.go"),
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/model")
	}
}

func TestTrafficEventQueryUsesApplicationResultInsteadOfHTTPDTO(t *testing.T) {
	t.Parallel()

	files := []string{
		filepath.Join("application", "queries", "awd_traffic_event_list_query.go"),
		filepath.Join("application", "queries", "awd_traffic_event_result.go"),
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
	}
}

func TestContestAWDServiceQueryUsesApplicationResultInsteadOfHTTPDTO(t *testing.T) {
	t.Parallel()

	files := []string{
		filepath.Join("application", "queries", "contest_awd_service_query.go"),
		filepath.Join("application", "queries", "contest_awd_service_result.go"),
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/model")
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

func TestPortsDoNotDependOnFrameworksDTOOrCacheClients(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("ports", "*.go"))
	if err != nil {
		t.Fatalf("glob ports files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
		assertFileDoesNotImport(t, file, "github.com/redis/go-redis/v9")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/commands")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/queries")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/contest/application/jobs")
	}
}

func TestPortsDoNotExposeGORMTags(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("ports", "*.go"))
	if err != nil {
		t.Fatalf("glob ports files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotDeclareGORMTags(t, file)
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

func TestRuntimeOwnsContestWiring(t *testing.T) {
	t.Parallel()

	runtimeFile := filepath.Join("runtime", "module.go")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/contest/infrastructure")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/contest/application/commands")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/contest/application/jobs")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/contest/application/queries")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/contest/api/http")
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

func assertFileDoesNotDeclareGORMTags(t *testing.T, filePath string) {
	t.Helper()

	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		t.Fatalf("parse file %s: %v", filePath, err)
	}

	ast.Inspect(fileNode, func(node ast.Node) bool {
		field, ok := node.(*ast.Field)
		if !ok || field.Tag == nil {
			return true
		}
		tag, err := strconv.Unquote(field.Tag.Value)
		if err != nil {
			t.Fatalf("unquote struct tag %s in %s: %v", field.Tag.Value, filePath, err)
		}
		if strings.Contains(tag, `gorm:"`) {
			t.Fatalf("%s must not expose gorm tags in ports", filePath)
		}
		return true
	})
}
