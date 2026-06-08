package module

import (
	"path/filepath"
	"strings"
	"testing"

	"ctf-platform/internal/testutil/archtest"
)

const moduleImportPrefix = "ctf-platform/internal/module/"

func TestModuleArchitectureBoundaries(t *testing.T) {
	t.Parallel()

	files := archtest.RuntimeGoFiles(t, ".")
	for _, file := range files {
		layer := moduleLayer(file)
		imports := archtest.Imports(t, file)
		assertNoCrossModulePrivateImports(t, file, imports)

		switch layer {
		case "domain":
			assertDomainInternalImportsAreReviewed(t, file, imports)
			assertNoForbiddenImports(t, file, imports, []string{
				"github.com/gin-gonic/gin",
				"github.com/redis/go-redis",
				"github.com/docker/docker",
				"gorm.io/gorm",
				"database/sql",
				"net/http",
			})
			assertNoModuleOuterLayerImports(t, file, imports, []string{
				"api",
				"application",
				"infrastructure",
				"runtime",
			})
		case "application":
			assertNoModuleOuterLayerImports(t, file, imports, []string{
				"api",
				"infrastructure",
				"runtime",
			})
			assertNoForbiddenImports(t, file, imports, []string{
				"github.com/gin-gonic/gin",
			})
			assertApplicationConcreteImportsAreReviewed(t, file, imports)
		case "ports":
			assertNoForbiddenImports(t, file, imports, []string{
				"github.com/gin-gonic/gin",
				"github.com/redis/go-redis",
				"github.com/docker/docker",
				"gorm.io/gorm",
				"database/sql",
				"net/http",
			})
		case "api":
			assertNoModuleOuterLayerImports(t, file, imports, []string{
				"infrastructure",
				"runtime",
			})
		}
	}
}

func TestModuleDependencyBaselineIsCurrent(t *testing.T) {
	t.Parallel()

	files := archtest.RuntimeGoFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		for _, importPath := range archtest.Imports(t, file) {
			if key, ok := moduleDependencyKey(file, importPath); ok {
				actual[key] = struct{}{}
				if _, known := moduleDependencyBaseline[key]; !known {
					t.Fatalf("module dependency is outside the reviewed baseline: %s via %s", key, file)
				}
			}
		}
	}

	for baseline := range moduleDependencyBaseline {
		if _, exists := actual[baseline]; !exists {
			t.Fatalf("module dependency baseline entry is stale: %s", baseline)
		}
	}
}

func TestDomainInternalImportExceptionsAreCurrent(t *testing.T) {
	t.Parallel()

	files := archtest.RuntimeGoFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		if moduleLayer(file) != "domain" {
			continue
		}
		for _, importPath := range archtest.Imports(t, file) {
			if isDomainInternalImport(importPath) {
				actual[domainInternalImportKey(file, importPath)] = struct{}{}
			}
		}
	}

	for exception := range reviewedDomainInternalImportExceptions {
		if _, exists := actual[exception]; !exists {
			t.Fatalf("domain internal import exception is stale: %s", exception)
		}
	}
}

func TestApplicationConcreteDependencyExceptionsAreCurrent(t *testing.T) {
	t.Parallel()

	files := archtest.RuntimeGoFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		if moduleLayer(file) != "application" {
			continue
		}
		for _, importPath := range archtest.Imports(t, file) {
			if isConcreteApplicationImport(importPath) {
				actual[applicationConcreteImportKey(file, importPath)] = struct{}{}
			}
		}
	}

	for exception := range reviewedApplicationConcreteImportExceptions {
		if _, exists := actual[exception]; !exists {
			t.Fatalf("application concrete dependency exception is stale: %s", exception)
		}
	}
}

func TestCrossModulePrivateImportExceptionsAreCurrent(t *testing.T) {
	t.Parallel()

	files := archtest.RuntimeGoFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		for _, importPath := range archtest.Imports(t, file) {
			if isCrossModulePrivateImport(file, importPath) {
				actual[crossModuleImportKey(file, importPath)] = struct{}{}
			}
		}
	}

	for exception := range reviewedCrossModulePrivateImportExceptions {
		if _, exists := actual[exception]; !exists {
			t.Fatalf("cross-module private import exception is stale: %s", exception)
		}
	}
}

func TestModuleRuntimeCodeDoesNotCreateRootContext(t *testing.T) {
	t.Parallel()

	files := archtest.RuntimeGoFiles(t, ".")
	for _, file := range files {
		content := archtest.ReadFile(t, file)
		if strings.Contains(content, "context.Background()") || strings.Contains(content, "context.TODO()") {
			t.Fatalf("%s must receive context from its caller instead of creating a root context", file)
		}
	}
}

func TestBackendBusinessCodeDoesNotCreateRootContext(t *testing.T) {
	t.Parallel()

	files := archtest.RuntimeGoFiles(t, "..")
	allowedRootContextFiles := map[string]struct{}{
		"../app/composition/root.go":              {},
		"../bootstrap/awd_defense_ssh_gateway.go": {},
		"../bootstrap/run.go":                     {},
	}
	for _, file := range files {
		content := archtest.ReadFile(t, file)
		if !strings.Contains(content, "context.Background()") && !strings.Contains(content, "context.TODO()") {
			continue
		}
		if _, allowed := allowedRootContextFiles[filepath.ToSlash(file)]; allowed {
			continue
		}
		t.Fatalf("%s must receive context from its caller instead of creating a root context", file)
	}
	for allowed := range allowedRootContextFiles {
		content := archtest.ReadFile(t, allowed)
		if !strings.Contains(content, "context.Background()") && !strings.Contains(content, "context.TODO()") {
			t.Fatalf("root context exception is stale: %s", allowed)
		}
	}
}

func TestTimeNowUsageExceptionsAreCurrent(t *testing.T) {
	t.Parallel()

	files := archtest.RuntimeGoFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		if strings.Contains(archtest.ReadFile(t, file), "time.Now(") {
			actual[moduleFileKey(file)] = struct{}{}
		}
	}

	for file := range actual {
		if _, allowed := reviewedTimeNowFiles[file]; !allowed {
			t.Fatalf("%s uses time.Now; use UTC business time or add a reviewed exception", file)
		}
	}
	for allowed := range reviewedTimeNowFiles {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("time.Now exception is stale: %s", allowed)
		}
	}
}

func TestTransactionBoundaryExceptionsAreCurrent(t *testing.T) {
	t.Parallel()

	files := archtest.RuntimeGoFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		if strings.Contains(archtest.ReadFile(t, file), ".Transaction(") {
			actual[file] = struct{}{}
		}
	}

	for file := range actual {
		if _, allowed := reviewedTransactionBoundaryFiles[file]; !allowed {
			t.Fatalf("%s opens a transaction outside the reviewed boundary exceptions", file)
		}
	}
	for allowed := range reviewedTransactionBoundaryFiles {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("transaction exception is stale: %s", allowed)
		}
	}
}

func TestRuntimeModulesStaySmallAndWiringOnly(t *testing.T) {
	t.Parallel()

	files := archtest.RuntimeGoFiles(t, ".")
	for _, file := range files {
		if !strings.HasSuffix(filepath.ToSlash(file), "/runtime/module.go") {
			continue
		}
		lineCount := len(strings.Split(archtest.ReadFile(t, file), "\n"))
		if lineCount <= 250 {
			continue
		}
		if _, allowed := reviewedOversizedRuntimeModuleFiles[file]; !allowed {
			t.Fatalf("%s has %d lines; runtime module files should stay wiring-only", file, lineCount)
		}
	}
	for allowed := range reviewedOversizedRuntimeModuleFiles {
		content := archtest.ReadFile(t, allowed)
		if len(strings.Split(content, "\n")) <= 250 {
			t.Fatalf("runtime module size exception is stale: %s", allowed)
		}
	}
}

func moduleLayer(filePath string) string {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	if len(parts) < 2 {
		return ""
	}
	for i := 1; i < len(parts); i++ {
		switch parts[i] {
		case "api", "application", "domain", "infrastructure", "ports", "runtime":
			return parts[i]
		}
	}
	return ""
}

func moduleOwner(filePath string) string {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func moduleFileKey(filePath string) string {
	return filepath.ToSlash(filePath)
}

func assertNoForbiddenImports(t *testing.T, filePath string, imports []string, forbidden []string) {
	t.Helper()

	for _, importPath := range imports {
		for _, blocked := range forbidden {
			if archtest.ImportPathMatches(importPath, blocked) {
				t.Fatalf("%s must not import %s", filePath, importPath)
			}
		}
	}
}

func assertNoModuleOuterLayerImports(t *testing.T, filePath string, imports []string, forbiddenLayers []string) {
	t.Helper()

	for _, importPath := range imports {
		if !strings.HasPrefix(importPath, moduleImportPrefix) {
			continue
		}
		modulePath := strings.TrimPrefix(importPath, moduleImportPrefix)
		parts := strings.Split(modulePath, "/")
		if len(parts) < 2 {
			continue
		}
		importedLayer := parts[1]
		for _, forbiddenLayer := range forbiddenLayers {
			if importedLayer == forbiddenLayer {
				t.Fatalf("%s must not import outer module layer %s", filePath, importPath)
			}
		}
	}
}

func assertNoCrossModulePrivateImports(t *testing.T, filePath string, imports []string) {
	t.Helper()

	for _, importPath := range imports {
		if !isCrossModulePrivateImport(filePath, importPath) {
			continue
		}
		key := crossModuleImportKey(filePath, importPath)
		if _, allowed := reviewedCrossModulePrivateImportExceptions[key]; !allowed {
			t.Fatalf("%s must not import private layer from another module: %s", filePath, importPath)
		}
	}
}

func assertDomainInternalImportsAreReviewed(t *testing.T, filePath string, imports []string) {
	t.Helper()

	for _, importPath := range imports {
		if !isDomainInternalImport(importPath) {
			continue
		}
		key := domainInternalImportKey(filePath, importPath)
		if _, allowed := reviewedDomainInternalImportExceptions[key]; !allowed {
			t.Fatalf("%s imports %s from domain; move through a domain-owned type or update the reviewed baseline", filePath, importPath)
		}
	}
}

func isDomainInternalImport(importPath string) bool {
	return importPath == "ctf-platform/internal/model" ||
		importPath == "ctf-platform/internal/dto" ||
		importPath == "ctf-platform/internal/config" ||
		strings.HasPrefix(importPath, "ctf-platform/internal/model/") ||
		strings.HasPrefix(importPath, "ctf-platform/internal/dto/") ||
		strings.HasPrefix(importPath, "ctf-platform/internal/config/")
}

func domainInternalImportKey(filePath string, importPath string) string {
	return filepath.ToSlash(filePath) + " -> " + importPath
}

func moduleDependencyKey(filePath string, importPath string) (string, bool) {
	if !strings.HasPrefix(importPath, moduleImportPrefix) {
		return "", false
	}
	currentModule := moduleOwner(filePath)
	modulePath := strings.TrimPrefix(importPath, moduleImportPrefix)
	parts := strings.Split(modulePath, "/")
	if len(parts) == 0 || parts[0] == currentModule {
		return "", false
	}
	return currentModule + " -> " + parts[0], true
}

func isCrossModulePrivateImport(filePath string, importPath string) bool {
	if !strings.HasPrefix(importPath, moduleImportPrefix) {
		return false
	}
	modulePath := strings.TrimPrefix(importPath, moduleImportPrefix)
	parts := strings.Split(modulePath, "/")
	if len(parts) < 2 || parts[0] == moduleOwner(filePath) {
		return false
	}
	return parts[1] != "contracts" && parts[1] != "ports"
}

func crossModuleImportKey(filePath string, importPath string) string {
	return filepath.ToSlash(filePath) + " -> " + importPath
}

func assertApplicationConcreteImportsAreReviewed(t *testing.T, filePath string, imports []string) {
	t.Helper()

	for _, importPath := range imports {
		if !isConcreteApplicationImport(importPath) {
			continue
		}
		key := applicationConcreteImportKey(filePath, importPath)
		if _, allowed := reviewedApplicationConcreteImportExceptions[key]; !allowed {
			t.Fatalf("%s imports concrete dependency %s; add a port/infrastructure adapter instead of growing reviewed exceptions", filePath, importPath)
		}
	}
}

func isConcreteApplicationImport(importPath string) bool {
	concretePrefixes := []string{
		"gorm.io/gorm",
		"github.com/redis/go-redis",
		"github.com/docker/docker",
		"database/sql",
		"net/http",
	}
	for _, prefix := range concretePrefixes {
		if importPath == prefix || strings.HasPrefix(importPath, prefix+"/") {
			return true
		}
	}
	return false
}

func applicationConcreteImportKey(filePath string, importPath string) string {
	return filepath.ToSlash(filePath) + " -> " + importPath
}
