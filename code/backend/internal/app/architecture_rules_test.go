package app

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

type importViolation struct {
	filePath    string
	importPath  string
	moduleName  string
	targetLayer string
}

func TestArchitectureRulesRejectConcreteCrossModuleImports(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..")
	violations, err := scanConcreteCrossModuleImports(repoRoot)
	if err != nil {
		t.Fatalf("scanConcreteCrossModuleImports() error = %v", err)
	}
	if len(violations) > 0 {
		t.Fatalf("unexpected concrete cross-module imports: %+v", violations)
	}
}

func scanConcreteCrossModuleImports(repoRoot string) ([]importViolation, error) {
	moduleRoot := filepath.Join(repoRoot, "internal", "module")
	fset := token.NewFileSet()
	violations := make([]importViolation, 0)

	err := filepath.Walk(moduleRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info == nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		moduleName, ok := moduleNameFromFilePath(moduleRoot, path)
		if !ok {
			return nil
		}

		fileNode, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, importSpec := range fileNode.Imports {
			importPath, err := strconv.Unquote(importSpec.Path.Value)
			if err != nil {
				return err
			}
			targetModule, targetLayer, ok := concreteCrossModuleImport(moduleName, importPath)
			if !ok {
				continue
			}
			violations = append(violations, importViolation{
				filePath:    path,
				importPath:  importPath,
				moduleName:  targetModule,
				targetLayer: targetLayer,
			})
		}
		return nil
	})
	return violations, err
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
		case "runtimeinfra", "challenge", "practice", "system", "auth":
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
