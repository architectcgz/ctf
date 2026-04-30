package app

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestBackendOperationalBoundariesUseContext(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..")
	files, err := backendGoFiles(repoRoot)
	if err != nil {
		t.Fatalf("list backend files: %v", err)
	}

	violations := make([]string, 0)
	for _, file := range files {
		rel := filepath.ToSlash(file)
		fileSet := token.NewFileSet()
		node, err := parser.ParseFile(fileSet, file, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		if isPortsOrContractsFile(rel) {
			violations = append(violations, portInterfaceContextViolations(fileSet, node)...)
		}
		if isInfrastructureRepositoryFile(rel) {
			violations = append(violations, repositoryMethodContextViolations(fileSet, node)...)
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("backend operational boundaries must accept ctx context.Context:\n%s", strings.Join(violations, "\n"))
	}
}

func TestBackendDoesNotExposeWithContextNames(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..")
	files, err := backendGoFiles(repoRoot)
	if err != nil {
		t.Fatalf("list backend files: %v", err)
	}

	violations := make([]string, 0)
	for _, file := range files {
		fileSet := token.NewFileSet()
		node, err := parser.ParseFile(fileSet, file, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		for _, decl := range node.Decls {
			funcDecl, ok := decl.(*ast.FuncDecl)
			if !ok || funcDecl.Name == nil {
				continue
			}
			if isExportedWithContextName(funcDecl.Name.Name) {
				violations = append(violations, positionString(fileSet, funcDecl.Name.Pos())+" "+funcDecl.Name.Name)
			}
		}
		ast.Inspect(node, func(n ast.Node) bool {
			typeSpec, ok := n.(*ast.TypeSpec)
			if !ok {
				return true
			}
			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				return true
			}
			for _, method := range interfaceType.Methods.List {
				for _, name := range method.Names {
					if isExportedWithContextName(name.Name) {
						violations = append(violations, positionString(fileSet, name.Pos())+" "+name.Name)
					}
				}
			}
			return false
		})
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("backend should use Foo(ctx, ...) names, not FooWithContext(ctx, ...):\n%s", strings.Join(violations, "\n"))
	}
}

func TestContextBackgroundOnlyAtApprovedRoots(t *testing.T) {
	t.Parallel()

	repoRoot := filepath.Join("..", "..")
	files, err := backendGoFiles(repoRoot)
	if err != nil {
		t.Fatalf("list backend files: %v", err)
	}

	violations := make([]string, 0)
	for _, file := range files {
		rel, err := filepath.Rel(repoRoot, file)
		if err != nil {
			t.Fatalf("rel %s: %v", file, err)
		}
		rel = filepath.ToSlash(rel)
		if isApprovedContextRootFile(rel) || strings.Contains(rel, "/testsupport/") {
			continue
		}
		fileSet := token.NewFileSet()
		node, err := parser.ParseFile(fileSet, file, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", file, err)
		}
		ast.Inspect(node, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			ident, ok := selector.X.(*ast.Ident)
			if !ok || ident.Name != "context" {
				return true
			}
			if selector.Sel.Name == "Background" || selector.Sel.Name == "TODO" {
				violations = append(violations, positionString(fileSet, selector.Pos())+" context."+selector.Sel.Name)
			}
			return true
		})
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("context.Background/TODO should only appear at approved roots:\n%s", strings.Join(violations, "\n"))
	}
}

func backendGoFiles(repoRoot string) ([]string, error) {
	root := filepath.Join(repoRoot, "internal")
	files := make([]string, 0)
	err := filepath.Walk(root, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info == nil || info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
}

func isPortsOrContractsFile(path string) bool {
	return strings.Contains(path, "/internal/module/") && (strings.Contains(path, "/ports/") || strings.Contains(path, "/contracts/"))
}

func isInfrastructureRepositoryFile(path string) bool {
	return strings.Contains(path, "/internal/module/") && strings.Contains(path, "/infrastructure/")
}

func portInterfaceContextViolations(fileSet *token.FileSet, node *ast.File) []string {
	violations := make([]string, 0)
	ast.Inspect(node, func(n ast.Node) bool {
		typeSpec, ok := n.(*ast.TypeSpec)
		if !ok {
			return true
		}
		interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
		if !ok {
			return true
		}
		for _, method := range interfaceType.Methods.List {
			funcType, ok := method.Type.(*ast.FuncType)
			if !ok {
				continue
			}
			for _, name := range method.Names {
				if !name.IsExported() || isAllowedNoContextPortMethod(name.Name) {
					continue
				}
				if !firstParamIsContext(funcType.Params) {
					violations = append(violations, positionString(fileSet, name.Pos())+" "+name.Name)
				}
			}
		}
		return false
	})
	return violations
}

func repositoryMethodContextViolations(fileSet *token.FileSet, node *ast.File) []string {
	violations := make([]string, 0)
	for _, decl := range node.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Recv == nil || funcDecl.Name == nil || !funcDecl.Name.IsExported() {
			continue
		}
		if !receiverContainsRepository(funcDecl.Recv) || isAllowedNoContextRepositoryMethod(funcDecl.Name.Name) {
			continue
		}
		if !firstParamIsContext(funcDecl.Type.Params) {
			violations = append(violations, positionString(fileSet, funcDecl.Name.Pos())+" "+funcDecl.Name.Name)
		}
	}
	return violations
}

func firstParamIsContext(fields *ast.FieldList) bool {
	if fields == nil || len(fields.List) == 0 {
		return false
	}
	first := fields.List[0]
	selector, ok := first.Type.(*ast.SelectorExpr)
	if !ok {
		return false
	}
	pkg, ok := selector.X.(*ast.Ident)
	return ok && pkg.Name == "context" && selector.Sel.Name == "Context"
}

func receiverContainsRepository(receiver *ast.FieldList) bool {
	if receiver == nil || len(receiver.List) == 0 {
		return false
	}
	switch expr := receiver.List[0].Type.(type) {
	case *ast.Ident:
		return strings.Contains(expr.Name, "Repository")
	case *ast.StarExpr:
		if ident, ok := expr.X.(*ast.Ident); ok {
			return strings.Contains(ident.Name, "Repository")
		}
	}
	return false
}

func isExportedWithContextName(name string) bool {
	return ast.IsExported(name) && strings.Contains(name, "WithContext")
}

func isAllowedNoContextPortMethod(name string) bool {
	switch name {
	case "Broadcast", "IsUniqueViolation", "SendToChannel", "SendToUser":
		return true
	default:
		return false
	}
}

func isAllowedNoContextRepositoryMethod(name string) bool {
	switch name {
	case "IsUniqueViolation", "WithDB":
		return true
	default:
		return false
	}
}

func isApprovedContextRootFile(path string) bool {
	switch path {
	case "internal/app/composition/root.go",
		"internal/app/http_server.go",
		"internal/bootstrap/run.go",
		"internal/infrastructure/postgres/postgres.go",
		"internal/infrastructure/redis/redis.go":
		return true
	default:
		return false
	}
}

func positionString(fileSet *token.FileSet, pos token.Pos) string {
	position := fileSet.Position(pos)
	return filepath.ToSlash(position.Filename) + ":" + strconv.Itoa(position.Line)
}
