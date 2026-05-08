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

		switch layer {
		case "domain":
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
