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
	forbiddenInterfaces := []string{
		"type ChallengeCommandRepository interface",
		"type ChallengeQueryRepository interface",
		"type ChallengeWriteupRepository interface",
		"type ChallengeTopologyRepository interface",
		"type ImageRepository interface",
		"type EnvironmentTemplateRepository interface",
		"type TagRepository interface",
	}
	for _, forbidden := range forbiddenInterfaces {
		if strings.Contains(string(content), forbidden) {
			t.Fatalf("challenge ports must not declare wide interface %s", forbidden)
		}
	}
}

func TestContractsDoNotReExportSharedTaxonomy(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("contracts", "*.go"))
	if err != nil {
		t.Fatalf("glob challenge contracts files: %v", err)
	}

	blockedMarkers := []string{
		"DimensionWeb",
		"DimensionPwn",
		"DimensionReverse",
		"DimensionCrypto",
		"DimensionMisc",
		"DimensionForensics",
		"AllDimensions",
		"IsValidDimension(",
		"ChallengeDifficultyBeginner",
		"ChallengeDifficultyEasy",
		"ChallengeDifficultyMedium",
		"ChallengeDifficultyHard",
		"ChallengeDifficultyInsane",
		"internal/shared/taxonomy",
	}

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read challenge contracts file %s: %v", file, err)
		}
		source := string(content)
		for _, marker := range blockedMarkers {
			if strings.Contains(source, marker) {
				t.Fatalf("challenge contracts must not re-export shared taxonomy marker %s in %s", marker, file)
			}
		}
	}
}

func TestRuntimeOwnsChallengeWiring(t *testing.T) {
	t.Parallel()

	assertFileImports(t, filepath.Join("runtime", "module.go"), "ctf-platform/internal/module/challenge/infrastructure")
	assertFileImports(t, filepath.Join("runtime", "module.go"), "ctf-platform/internal/module/challenge/api/http")
	assertFileImports(t, filepath.Join("runtime", "wiring.go"), "ctf-platform/internal/module/challenge/application/commands")
	assertFileImports(t, filepath.Join("runtime", "wiring.go"), "ctf-platform/internal/module/challenge/application/queries")
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

func TestChallengeCommandServicesAreSeparated(t *testing.T) {
	t.Parallel()

	expectedServiceFiles := map[string]string{
		filepath.Join("application", "challengecore", "service.go"):          "type ChallengeService struct",
		filepath.Join("application", "challengeimport", "service.go"):        "type ChallengeImportService struct",
		filepath.Join("application", "challengeselfcheck", "service.go"):     "type ChallengeSelfCheckService struct",
		filepath.Join("application", "challengepublishcheck", "service.go"):  "type ChallengePublishCheckService struct",
		filepath.Join("application", "challengepackageexport", "service.go"): "type ChallengePackageExportService struct",
	}
	for file, marker := range expectedServiceFiles {
		source := readFileSource(t, file)
		if !strings.Contains(source, marker) {
			t.Fatalf("%s must contain %s", file, marker)
		}
	}

	challengeImportService := readFileSource(t, filepath.Join("application", "challengeimport", "service.go"))
	if !strings.Contains(challengeImportService, "type ChallengeImportService struct") {
		t.Fatalf("jeopardy import flow must be owned by ChallengeImportService")
	}
	if strings.Contains(challengeImportService, "type ChallengeService struct") {
		t.Fatalf("challenge_import_service.go must not define the core ChallengeService")
	}

	coreService := readFileSource(t, filepath.Join("application", "challengecore", "service.go"))
	blockedCoreMethods := []string{
		"PreviewChallengeImport",
		"ListChallengeImports",
		"GetChallengeImport",
		"CommitChallengeImport",
		"SelfCheckChallenge",
		"RequestPublishCheck",
		"GetLatestPublishCheck",
		"RunPublishCheckLoop",
		"ExportChallengePackage",
		"GetChallengePackageExport",
	}
	for _, method := range blockedCoreMethods {
		marker := "func (s *ChallengeService) " + method + "("
		if strings.Contains(coreService, marker) {
			t.Fatalf("core ChallengeService must not own %s", method)
		}
	}

	commandFiles, err := filepath.Glob(filepath.Join("application", "commands", "*.go"))
	if err != nil {
		t.Fatalf("glob commands files: %v", err)
	}
	blockedCommandMarkers := []string{
		"type ChallengeService struct",
		"type ChallengeImportService struct",
		"type ChallengeSelfCheckService struct",
		"type ChallengePublishCheckService struct",
		"type ChallengePackageExportService struct",
	}
	blockedUseCaseImports := []string{
		"ctf-platform/internal/module/challenge/application/challengecore",
		"ctf-platform/internal/module/challenge/application/challengeimport",
		"ctf-platform/internal/module/challenge/application/challengeselfcheck",
		"ctf-platform/internal/module/challenge/application/challengepublishcheck",
		"ctf-platform/internal/module/challenge/application/challengepackageexport",
	}
	for _, file := range commandFiles {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		source := readFileSource(t, file)
		for _, marker := range blockedCommandMarkers {
			if strings.Contains(source, marker) {
				t.Fatalf("commands package must not retain split service marker %s in %s", marker, file)
			}
		}
		for _, importPath := range blockedUseCaseImports {
			assertFileDoesNotImport(t, file, importPath)
		}
	}

	runtime := readFileSource(t, filepath.Join("runtime", "wiring.go"))
	for _, importPath := range blockedUseCaseImports {
		if !strings.Contains(runtime, importPath) {
			t.Fatalf("runtime wiring must import split use-case package %s directly", importPath)
		}
	}

	handler := readFileSource(t, filepath.Join("api", "http", "handler.go"))
	if strings.Contains(handler, "NewPackageDeliveryService(commands, nil)") {
		t.Fatalf("handler must receive package delivery wiring instead of constructing it from the wide command service")
	}
}

func TestChallengeCommandFileStorageBoundaryIsInInfrastructure(t *testing.T) {
	t.Parallel()

	blockedMarkers := []string{
		"writeImportUploadArchive",
		"extractChallengeImportArchive",
		"saveChallengeImportPreviewRecord",
		"loadChallengeImportPreviewRecord",
		"saveAWDChallengeImportPreviewRecord",
		"loadAWDChallengeImportPreviewRecord",
		"challengeImportPreviewRoot",
		"challengeImportedAttachmentRoot",
		"challengePackageSourceRoot",
		"challengePackageExportRoot",
		"importedImageBuildSourceRoot",
		"awdChallengeImportPreviewRoot",
		"persistImportedImageBuildSource",
		"copyDirectoryTree",
		"zipDirectory",
		"addZipFile",
		"DefaultArtifactGCConfigFromEnv",
	}

	files := []string{
		filepath.Join("application", "challengeimport", "service.go"),
		filepath.Join("application", "challengeimport", "package_revision.go"),
		filepath.Join("application", "challengepackageexport", "revision_service.go"),
		filepath.Join("application", "commands", "awd_challenge_import_service.go"),
		filepath.Join("application", "commands", "artifact_gc_service.go"),
	}
	for _, file := range files {
		source := readFileSource(t, file)
		for _, marker := range blockedMarkers {
			if strings.Contains(source, marker) {
				t.Fatalf("%s must not own LocalFS/zip helper %s", file, marker)
			}
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
