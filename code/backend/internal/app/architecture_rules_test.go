package app

import (
	"path/filepath"
	"strings"
	"testing"

	"ctf-platform/internal/testutil/archtest"
)

type importViolation struct {
	filePath    string
	importPath  string
	moduleName  string
	targetLayer string
}

var reviewedConcreteCrossModuleImportExceptions = map[string]struct{}{
	"../../internal/module/container_runtime/runtime/module.go -> ctf-platform/internal/module/runtime/application":          {},
	"../../internal/module/container_runtime/runtime/module.go -> ctf-platform/internal/module/runtime/application/commands": {},
}

func TestArchitectureRulesRejectConcreteCrossModuleImports(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..")
	violations, _ := scanConcreteCrossModuleImports(t, repoRoot)
	if len(violations) > 0 {
		t.Fatalf("unexpected concrete cross-module imports: %+v", violations)
	}
}

func TestArchitectureRulesConcreteCrossModuleImportExceptionsAreCurrent(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..")
	_, actualExceptions := scanConcreteCrossModuleImports(t, repoRoot)
	for exception := range reviewedConcreteCrossModuleImportExceptions {
		if _, exists := actualExceptions[exception]; !exists {
			t.Fatalf("concrete cross-module import exception is stale: %s", exception)
		}
	}
}

func scanConcreteCrossModuleImports(t testing.TB, repoRoot string) ([]importViolation, map[string]struct{}) {
	t.Helper()

	moduleRoot := filepath.Join(repoRoot, "internal", "module")
	violations := make([]importViolation, 0)
	actualExceptions := make(map[string]struct{})

	for _, path := range archtest.RuntimeGoFiles(t, moduleRoot) {
		moduleName, ok := moduleNameFromFilePath(moduleRoot, path)
		if !ok {
			continue
		}

		for _, importPath := range archtest.Imports(t, path) {
			targetModule, targetLayer, ok := concreteCrossModuleImport(moduleName, importPath)
			if !ok {
				continue
			}
			key := concreteCrossModuleImportExceptionKey(path, importPath)
			if _, allowed := reviewedConcreteCrossModuleImportExceptions[key]; allowed {
				actualExceptions[key] = struct{}{}
				continue
			}
			violations = append(violations, importViolation{
				filePath:    path,
				importPath:  importPath,
				moduleName:  targetModule,
				targetLayer: targetLayer,
			})
		}
	}
	return violations, actualExceptions
}

func concreteCrossModuleImportExceptionKey(filePath string, importPath string) string {
	return filepath.ToSlash(filePath) + " -> " + importPath
}

func moduleNameFromFilePath(moduleRoot, filePath string) (string, bool) {
	relPath, err := filepath.Rel(moduleRoot, filePath)
	if err != nil {
		return "", false
	}
	parts := strings.Split(relPath, string(filepath.Separator))
	if len(parts) == 0 {
		return "", false
	}
	return parts[0], true
}

func concreteCrossModuleImport(sourceModule, importPath string) (moduleName string, targetLayer string, ok bool) {
	const prefix = "ctf-platform/internal/module/"
	if !strings.HasPrefix(importPath, prefix) {
		return "", "", false
	}

	parts := strings.Split(strings.TrimPrefix(importPath, prefix), "/")
	if len(parts) == 0 {
		return "", "", false
	}
	if parts[0] == sourceModule {
		return "", "", false
	}
	if len(parts) == 1 {
		switch parts[0] {
		case "auth", "challenge", "practice", "system":
			return parts[0], "root", true
		default:
			return "", "", false
		}
	}
	if len(parts) < 2 {
		return "", "", false
	}

	layer := parts[1]
	switch layer {
	case "application", "infrastructure", "api":
		return parts[0], layer, true
	default:
		return "", "", false
	}
}
