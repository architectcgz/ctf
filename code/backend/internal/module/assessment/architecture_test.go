package assessment

import (
	"go/parser"
	"go/token"
	"io/fs"
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
		t.Fatalf("glob assessment root files: %v", err)
	}

	nonTestFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		nonTestFiles = append(nonTestFiles, file)
	}

	if len(nonTestFiles) != 0 {
		t.Fatalf("assessment root package should keep no non-test go files, got %v", nonTestFiles)
	}
}

func TestAPIHTTPDoesNotDependOnInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("api", "http", "*.go"))
	if err != nil {
		t.Fatalf("glob api/http files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/assessment/infrastructure")
	}
}

func TestSkillProfileQueryAndHTTPDoNotDependOnGlobalDTO(t *testing.T) {
	t.Parallel()

	assertFileDoesNotImport(t, filepath.Join("api", "http", "handler.go"), "ctf-platform/internal/dto")
	assertFileDoesNotImport(t, filepath.Join("application", "queries", "profile_service.go"), "ctf-platform/internal/dto")
}

func TestReportFlowDoesNotDependOnGlobalDTO(t *testing.T) {
	t.Parallel()

	files := []string{
		filepath.Join("api", "http", "report_handler.go"),
		filepath.Join("api", "http", "teacher_awd_review_handler.go"),
		filepath.Join("application", "commands", "report_service.go"),
		filepath.Join("application", "commands", "response_mapper_goverter.go"),
		filepath.Join("runtime", "module.go"),
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
	}
}

func TestAssessmentRuntimeCodeDoesNotDependOnTeachingQuery(t *testing.T) {
	t.Parallel()

	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case "data", "testdata", "testsupport":
				return filepath.SkipDir
			default:
				return nil
			}
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		assertFileDoesNotImportPrefix(t, path, "ctf-platform/internal/module/teaching_query")
		return nil
	})
	if err != nil {
		t.Fatalf("walk assessment runtime files: %v", err)
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
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/assessment/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/assessment/infrastructure")
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
	}
}

func TestReportOutputStorageBoundaryIsInInfrastructure(t *testing.T) {
	t.Parallel()

	commandFiles, err := filepath.Glob(filepath.Join("application", "commands", "*.go"))
	if err != nil {
		t.Fatalf("glob commands files: %v", err)
	}
	for _, file := range commandFiles {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "os")
		assertFileDoesNotImport(t, file, "path/filepath")
	}

	blockedMarkers := []string{
		"os.MkdirAll",
		"os.Stat",
		"os.IsNotExist",
		"filepath.Abs",
		"filepath.Clean",
		"safeReportPath",
	}
	files := []string{
		filepath.Join("application", "commands", "report_service.go"),
		filepath.Join("application", "commands", "report_file_output.go"),
		filepath.Join("application", "commands", "report_generation.go"),
	}
	for _, file := range files {
		source := readFileSource(t, file)
		for _, marker := range blockedMarkers {
			if strings.Contains(source, marker) {
				t.Fatalf("%s must not own report output storage helper %s", file, marker)
			}
		}
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
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/assessment/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/assessment/infrastructure")
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

func TestContractsDoNotDependOnDTOGinOrGORM(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("contracts", "*.go"))
	if err != nil {
		t.Fatalf("glob contracts files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
	}
}

func TestRuntimeOwnsAssessmentWiring(t *testing.T) {
	t.Parallel()

	runtimeFile := filepath.Join("runtime", "module.go")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/assessment/infrastructure")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/assessment/application/commands")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/assessment/application/queries")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/assessment/api/http")
}

func TestRuntimeUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("runtime", "module.go"))
	if err != nil {
		t.Fatalf("read assessment runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"type moduleDeps struct",
		"profileRepo",
		"assessmentports.ProfileRepository",
		"recommendationRepo",
		"assessmentports.RecommendationRepository",
		"reportRepo",
		"assessmentports.ReportRepository",
		"classInsightRepo",
		"assessmentports.AssessmentClassInsightRepository",
		"awdReviewRepo",
		"assessmentports.TeacherAWDReviewRepository",
		"challengeRepo",
		"assessmentports.ChallengeRepository",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("assessment runtime should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"type assessmentModuleDeps struct",
		"type assessmentModuleExternalDeps struct",
		"repo := assessmentinfra.NewRepository(root.DB())",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("assessment runtime should not keep composition glue marker %s", marker)
		}
	}
}

func TestRuntimeDelegatesThroughSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("runtime", "module.go"))
	if err != nil {
		t.Fatalf("read assessment runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"newModuleDeps(",
		"buildProfileHandler(",
		"buildRecommendationHandler(",
		"buildReportHandler(",
		"buildTeacherAWDReviewHandler(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("assessment runtime should delegate through %s", marker)
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

func TestAssessmentTestsDoNotDependOnLegacyModelOrChallengeEntity(t *testing.T) {
	t.Parallel()

	err := filepath.WalkDir(".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		assertFileDoesNotImport(t, path, "ctf-platform/internal/model")
		assertFileDoesNotImport(t, path, "ctf-platform/internal/module/challenge/entity")
		return nil
	})
	if err != nil {
		t.Fatalf("walk assessment test files: %v", err)
	}
}

func readFileSource(t *testing.T, filePath string) string {
	t.Helper()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read file %s: %v", filePath, err)
	}
	return string(content)
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

func assertFileDoesNotImportPrefix(t *testing.T, filePath string, blockedPrefix string) {
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
		if importPath == blockedPrefix || strings.HasPrefix(importPath, blockedPrefix+"/") {
			t.Fatalf("%s must not import %s", filePath, importPath)
		}
	}
}
