package teaching_readmodel

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
		t.Fatalf("glob teaching_readmodel root files: %v", err)
	}

	nonTestFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		nonTestFiles = append(nonTestFiles, file)
	}

	if len(nonTestFiles) != 0 {
		t.Fatalf("teaching_readmodel root package should keep no non-test go files, got %v", nonTestFiles)
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
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/teaching_readmodel/infrastructure")
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
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/teaching_readmodel/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/teaching_readmodel/infrastructure")
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
	}
}

func TestPortsDoNotDependOnDTOGinOrGORM(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("ports", "*.go"))
	if err != nil {
		t.Fatalf("glob ports files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/teaching_readmodel/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/teaching_readmodel/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/teaching_readmodel/application/queries")
	}
}

func TestPortsDoNotDeclareWideRepository(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("ports", "query.go"))
	if err != nil {
		t.Fatalf("read teaching_readmodel ports file: %v", err)
	}
	if strings.Contains(string(content), "type Repository interface") {
		t.Fatalf("teaching_readmodel ports must not declare the legacy wide Repository interface")
	}
}

func TestInfrastructureDoesNotDependOnDTO(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("infrastructure", "*.go"))
	if err != nil {
		t.Fatalf("glob infrastructure files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
	}
}

func TestRuntimeOwnsTeachingReadmodelWiring(t *testing.T) {
	t.Parallel()

	runtimeFile := filepath.Join("runtime", "module.go")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/teaching_readmodel/infrastructure")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/teaching_readmodel/application/queries")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/teaching_readmodel/api/http")
}

func TestRuntimeUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("runtime", "module.go"))
	if err != nil {
		t.Fatalf("read teaching_readmodel runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"type moduleDeps struct",
		"readmodelports.TeachingClassInsightRepository",
		"recommendations assessmentcontracts.RecommendationProvider",
		"buildQueryService(",
		"buildOverviewService(",
		"buildClassInsightService(",
		"buildStudentReviewService(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("teaching_readmodel runtime should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"Query   readmodelqueries.Service",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("teaching_readmodel runtime should not keep wide query export marker %s", marker)
		}
	}
}

func TestContractsSplitStudentReviewOwner(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("application", "queries", "contracts.go"))
	if err != nil {
		t.Fatalf("read teaching_readmodel contracts: %v", err)
	}

	source := string(content)
	studentReviewBlock := extractInterfaceBlock(t, source, "StudentReviewService")
	expectedStudentReviewMethods := []string{
		"GetStudentProgress(",
		"GetStudentRecommendations(",
		"GetStudentTimeline(",
		"GetStudentEvidence(",
		"GetStudentAttackSessions(",
	}
	for _, marker := range expectedStudentReviewMethods {
		if !strings.Contains(studentReviewBlock, marker) {
			t.Fatalf("StudentReviewService must declare %s", marker)
		}
	}

	serviceBlock := extractInterfaceBlock(t, source, "Service")
	for _, marker := range expectedStudentReviewMethods {
		if strings.Contains(serviceBlock, marker) {
			t.Fatalf("directory Service must not keep student review method %s", marker)
		}
	}
}

func TestDirectoryQueryServiceDoesNotOwnStudentReviewMethods(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("application", "queries", "service.go"))
	if err != nil {
		t.Fatalf("read teaching_readmodel query service: %v", err)
	}

	source := string(content)
	blocked := []string{
		"func (s *QueryService) GetStudentProgress(",
		"func (s *QueryService) GetStudentRecommendations(",
		"func (s *QueryService) GetStudentTimeline(",
		"func (s *QueryService) GetStudentEvidence(",
		"func (s *QueryService) GetStudentAttackSessions(",
		"func (s *QueryService) getAccessibleStudent(",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("directory query service must not keep student review marker %s", marker)
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

func extractInterfaceBlock(t *testing.T, source, name string) string {
	t.Helper()

	marker := "type " + name + " interface {"
	start := strings.Index(source, marker)
	if start == -1 {
		t.Fatalf("interface %s not found", name)
	}

	rest := source[start:]
	end := strings.Index(rest, "\n}")
	if end == -1 {
		t.Fatalf("interface %s closing brace not found", name)
	}

	return rest[:end+2]
}
