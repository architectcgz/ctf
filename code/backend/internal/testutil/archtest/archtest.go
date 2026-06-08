package archtest

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// RuntimeGoFiles returns non-test Go files under the given roots while skipping
// fixture-style directories that should not affect production architecture gates.
func RuntimeGoFiles(t testing.TB, roots ...string) []string {
	t.Helper()

	files := make([]string, 0)
	for _, root := range roots {
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				switch entry.Name() {
				case "data", "testdata", "testsupport":
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
			t.Fatalf("walk go files under %s: %v", root, err)
		}
	}

	sort.Strings(files)
	return files
}

func Imports(t testing.TB, filePath string) []string {
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

func ReadFile(t testing.TB, filePath string) string {
	t.Helper()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read %s: %v", filePath, err)
	}
	return string(content)
}

func ImportPathMatches(importPath string, blockedPrefix string) bool {
	return importPath == blockedPrefix || strings.HasPrefix(importPath, blockedPrefix+"/")
}

func AssertFileDoesNotImport(t testing.TB, filePath string, blockedImport string) {
	t.Helper()

	for _, importPath := range Imports(t, filePath) {
		if importPath == blockedImport {
			t.Fatalf("%s must not import %s", filePath, blockedImport)
		}
	}
}
