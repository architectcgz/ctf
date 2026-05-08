package module

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const moduleImportPrefix = "ctf-platform/internal/module/"

func TestModuleArchitectureBoundaries(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	for _, file := range files {
		layer := moduleLayer(file)
		imports := parseImports(t, file)
		assertNoCrossModulePrivateImports(t, file, imports)

		switch layer {
		case "domain":
			assertDomainInternalImportsAreAllowlisted(t, file, imports)
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
			assertApplicationConcreteImportsAreAllowlisted(t, file, imports)
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

func TestModuleDependencyAllowlistIsCurrent(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		for _, importPath := range parseImports(t, file) {
			if key, ok := moduleDependencyKey(file, importPath); ok {
				actual[key] = struct{}{}
				if _, allowed := allowedModuleDependencies[key]; !allowed {
					t.Fatalf("module dependency is not allowlisted: %s via %s", key, file)
				}
			}
		}
	}

	for allowed := range allowedModuleDependencies {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("module dependency allowlist entry is stale: %s", allowed)
		}
	}
}

func TestDomainInternalImportAllowlistIsCurrent(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		if moduleLayer(file) != "domain" {
			continue
		}
		for _, importPath := range parseImports(t, file) {
			if isDomainInternalImport(importPath) {
				actual[domainInternalImportKey(file, importPath)] = struct{}{}
			}
		}
	}

	for allowed := range allowedDomainInternalImports {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("domain internal import allowlist entry is stale: %s", allowed)
		}
	}
}

func TestApplicationConcreteDependencyAllowlistIsCurrent(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		if moduleLayer(file) != "application" {
			continue
		}
		for _, importPath := range parseImports(t, file) {
			if isConcreteApplicationImport(importPath) {
				actual[applicationConcreteImportKey(file, importPath)] = struct{}{}
			}
		}
	}

	for allowed := range allowedApplicationConcreteImports {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("application concrete dependency allowlist entry is stale: %s", allowed)
		}
	}
}

func TestCrossModulePrivateImportAllowlistIsCurrent(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		for _, importPath := range parseImports(t, file) {
			if isCrossModulePrivateImport(file, importPath) {
				actual[crossModuleImportKey(file, importPath)] = struct{}{}
			}
		}
	}

	for allowed := range allowedCrossModulePrivateImports {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("cross-module private import allowlist entry is stale: %s", allowed)
		}
	}
}

func TestModuleRuntimeCodeDoesNotCreateRootContext(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	for _, file := range files {
		content := readFile(t, file)
		if strings.Contains(content, "context.Background()") || strings.Contains(content, "context.TODO()") {
			t.Fatalf("%s must receive context from its caller instead of creating a root context", file)
		}
	}
}

func TestBackendBusinessCodeDoesNotCreateRootContext(t *testing.T) {
	t.Parallel()

	files := collectBackendRuntimeFiles(t, "..")
	allowedRootContextFiles := map[string]struct{}{
		"../app/composition/root.go": {},
		"../bootstrap/run.go":        {},
	}
	for _, file := range files {
		content := readFile(t, file)
		if !strings.Contains(content, "context.Background()") && !strings.Contains(content, "context.TODO()") {
			continue
		}
		if _, allowed := allowedRootContextFiles[filepath.ToSlash(file)]; allowed {
			continue
		}
		t.Fatalf("%s must receive context from its caller instead of creating a root context", file)
	}
	for allowed := range allowedRootContextFiles {
		content := readFile(t, allowed)
		if !strings.Contains(content, "context.Background()") && !strings.Contains(content, "context.TODO()") {
			t.Fatalf("root context allowlist entry is stale: %s", allowed)
		}
	}
}

func TestTimeNowUsageAllowlistIsCurrent(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		if strings.Contains(readFile(t, file), "time.Now(") {
			actual[moduleFileKey(file)] = struct{}{}
		}
	}

	for file := range actual {
		if _, allowed := allowedTimeNowFiles[file]; !allowed {
			t.Fatalf("%s uses time.Now; use UTC business time or add a reviewed allowlist entry", file)
		}
	}
	for allowed := range allowedTimeNowFiles {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("time.Now allowlist entry is stale: %s", allowed)
		}
	}
}

func TestTransactionBoundaryAllowlistIsCurrent(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		if strings.Contains(readFile(t, file), ".Transaction(") {
			actual[file] = struct{}{}
		}
	}

	for file := range actual {
		if _, allowed := allowedTransactionFiles[file]; !allowed {
			t.Fatalf("%s opens a transaction outside the reviewed boundary allowlist", file)
		}
	}
	for allowed := range allowedTransactionFiles {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("transaction allowlist entry is stale: %s", allowed)
		}
	}
}

func TestRuntimeModulesStaySmallAndWiringOnly(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	for _, file := range files {
		if !strings.HasSuffix(filepath.ToSlash(file), "/runtime/module.go") {
			continue
		}
		lineCount := len(strings.Split(readFile(t, file), "\n"))
		if lineCount <= 250 {
			continue
		}
		if _, allowed := allowedOversizedRuntimeModules[file]; !allowed {
			t.Fatalf("%s has %d lines; runtime module files should stay wiring-only", file, lineCount)
		}
	}
	for allowed := range allowedOversizedRuntimeModules {
		content := readFile(t, allowed)
		if len(strings.Split(content, "\n")) <= 250 {
			t.Fatalf("runtime module size allowlist entry is stale: %s", allowed)
		}
	}
}

func collectGoRuntimeFiles(t *testing.T, root string) []string {
	t.Helper()

	files := make([]string, 0)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if entry.Name() == "testsupport" || entry.Name() == "data" {
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
		t.Fatalf("walk module go files: %v", err)
	}

	return files
}

func collectBackendRuntimeFiles(t *testing.T, roots ...string) []string {
	t.Helper()

	files := make([]string, 0)
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				switch entry.Name() {
				case "testsupport", "testdata", "data":
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
			t.Fatalf("walk backend go files under %s: %v", root, err)
		}
	}
	return files
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

func parseImports(t *testing.T, filePath string) []string {
	t.Helper()

	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", filePath, err)
	}

	imports := make([]string, 0, len(fileNode.Imports))
	for _, importSpec := range fileNode.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import in %s: %v", filePath, err)
		}
		imports = append(imports, importPath)
	}
	return imports
}

func readFile(t *testing.T, filePath string) string {
	t.Helper()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read %s: %v", filePath, err)
	}
	return string(content)
}

func moduleFileKey(filePath string) string {
	return filepath.ToSlash(filePath)
}

func assertNoForbiddenImports(t *testing.T, filePath string, imports []string, forbidden []string) {
	t.Helper()

	for _, importPath := range imports {
		for _, blocked := range forbidden {
			if importPath == blocked || strings.HasPrefix(importPath, blocked+"/") {
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
		if _, allowed := allowedCrossModulePrivateImports[key]; !allowed {
			t.Fatalf("%s must not import private layer from another module: %s", filePath, importPath)
		}
	}
}

func assertDomainInternalImportsAreAllowlisted(t *testing.T, filePath string, imports []string) {
	t.Helper()

	for _, importPath := range imports {
		if !isDomainInternalImport(importPath) {
			continue
		}
		key := domainInternalImportKey(filePath, importPath)
		if _, allowed := allowedDomainInternalImports[key]; !allowed {
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

func assertApplicationConcreteImportsAreAllowlisted(t *testing.T, filePath string, imports []string) {
	t.Helper()

	for _, importPath := range imports {
		if !isConcreteApplicationImport(importPath) {
			continue
		}
		key := applicationConcreteImportKey(filePath, importPath)
		if _, allowed := allowedApplicationConcreteImports[key]; !allowed {
			t.Fatalf("%s imports concrete dependency %s; add a port/infrastructure adapter instead of growing the allowlist", filePath, importPath)
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
