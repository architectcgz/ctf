package ops

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestRootPackageOnlyKeepsContractsAndModule(t *testing.T) {
	t.Parallel()

	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir(.) error = %v", err)
	}

	allowed := map[string]struct{}{
		"architecture_test.go": {},
		"contracts.go":         {},
		"module.go":            {},
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if _, ok := allowed[entry.Name()]; ok {
			continue
		}
		t.Fatalf("unexpected root file %s, ops should be physically layered", entry.Name())
	}
}

func TestLayerPackagesDoNotCrossImportConcreteImplementations(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..", "..", "..")
	fset := token.NewFileSet()
	packages := []struct {
		dir              string
		forbiddenImports []string
	}{
		{
			dir: "api/http",
			forbiddenImports: []string{
				"ctf-platform/internal/module/ops/infrastructure",
			},
		},
		{
			dir: "application",
			forbiddenImports: []string{
				"ctf-platform/internal/module/ops/api/http",
				"ctf-platform/internal/module/ops/infrastructure",
				"github.com/gin-gonic/gin",
			},
		},
	}

	for _, pkg := range packages {
		files, err := filepath.Glob(filepath.Join(pkg.dir, "*.go"))
		if err != nil {
			t.Fatalf("Glob(%s) error = %v", pkg.dir, err)
		}
		for _, filePath := range files {
			if strings.HasSuffix(filePath, "_test.go") {
				continue
			}

			node, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
			if err != nil {
				t.Fatalf("ParseFile(%s) error = %v", filePath, err)
			}

			for _, importSpec := range node.Imports {
				importPath, err := strconv.Unquote(importSpec.Path.Value)
				if err != nil {
					t.Fatalf("Unquote(%s) error = %v", importSpec.Path.Value, err)
				}
				for _, forbidden := range pkg.forbiddenImports {
					if importPath == forbidden {
						t.Fatalf("%s must not import %s", filepath.Join(repoRoot, "internal", "module", "ops", filePath), forbidden)
					}
				}
			}
		}
	}
}
