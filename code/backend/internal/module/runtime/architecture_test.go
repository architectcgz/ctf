package runtime

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestApplicationDoesNotDependOnHTTPOrRuntimeInfra(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("application", "*.go"))
	if err != nil {
		t.Fatalf("glob application files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "github.com/redis/go-redis/v9")
		assertFileDoesNotImport(t, file, "ctf-platform/pkg/response")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtimeinfra")
	}
}

func TestCommandsDoNotDependOnAPIHTTPOrInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("application", "commands", "*.go"))
	if err != nil {
		t.Fatalf("glob application command files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/application/queries")
	}
}

func TestQueriesDoNotDependOnAPIHTTPOrInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("application", "queries", "*.go"))
	if err != nil {
		t.Fatalf("glob application query files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/application/commands")
	}
}

func TestDomainDoesNotDependOnGinGORMOrRuntimeInfra(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("domain", "*.go"))
	if err != nil {
		t.Fatalf("glob domain files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtimeinfra")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
	}
}

func TestAPIHTTPDoesNotDependOnGORMOrRuntimeInfra(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("api", "http", "*.go"))
	if err != nil {
		t.Fatalf("glob api/http files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtimeinfra")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/application")
	}
}

func TestInfrastructureDoesNotDependOnDTOOrGin(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("infrastructure", "*.go"))
	if err != nil {
		t.Fatalf("glob infrastructure files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/runtime/application")
	}
}

func TestPortsDoNotDeclareWideInstanceRepository(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("ports", "http.go"))
	if err != nil {
		t.Fatalf("read runtime ports file: %v", err)
	}
	if strings.Contains(string(content), "type InstanceRepository interface") {
		t.Fatalf("runtime ports must not declare the legacy wide InstanceRepository interface")
	}
}

func TestRuntimeModuleDoesNotExposeLegacyEngineSurface(t *testing.T) {
	t.Parallel()

	fileNode := parseGoFile(t, filepath.Join("runtime", "module.go"))
	assertTypeDoesNotExist(t, fileNode, "Engine")
	assertStructDoesNotDeclareField(t, fileNode, "Module", "Engine")
	assertStructDoesNotDeclareField(t, fileNode, "Deps", "Engine")
}

func TestAPIHTTPDoesNotDeclareRetiredDefenseWorkbenchMethods(t *testing.T) {
	t.Parallel()

	fileNode := parseGoFile(t, filepath.Join("api", "http", "handler.go"))
	assertInterfaceDoesNotDeclareMethods(t, fileNode, "runtimeService",
		"ReadAWDDefenseFile",
		"ListAWDDefenseDirectory",
		"SaveAWDDefenseFile",
		"RunAWDDefenseCommand",
	)
}

func TestRootPackageKeepsOnlyDocFile(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob runtime root files: %v", err)
	}

	nonTestFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		nonTestFiles = append(nonTestFiles, file)
	}

	if len(nonTestFiles) != 1 || nonTestFiles[0] != "doc.go" {
		t.Fatalf("runtime root package should keep only doc.go, got %v", nonTestFiles)
	}
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

func parseGoFile(t *testing.T, filePath string) *ast.File {
	t.Helper()

	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, filePath, nil, 0)
	if err != nil {
		t.Fatalf("parse file %s: %v", filePath, err)
	}
	return fileNode
}

func assertTypeDoesNotExist(t *testing.T, fileNode *ast.File, typeName string) {
	t.Helper()

	if _, ok := lookupTypeSpec(fileNode, typeName); ok {
		t.Fatalf("%s must not be declared in this file", typeName)
	}
}

func assertStructDoesNotDeclareField(t *testing.T, fileNode *ast.File, typeName string, fieldName string) {
	t.Helper()

	typeSpec := findTypeSpec(t, fileNode, typeName)
	structType, ok := typeSpec.Type.(*ast.StructType)
	if !ok {
		t.Fatalf("%s must stay a struct", typeName)
	}

	for _, field := range structType.Fields.List {
		for _, name := range field.Names {
			if name.Name == fieldName {
				t.Fatalf("%s must not declare field %s", typeName, fieldName)
			}
		}
	}
}

func assertInterfaceDoesNotDeclareMethods(t *testing.T, fileNode *ast.File, typeName string, methodNames ...string) {
	t.Helper()

	typeSpec := findTypeSpec(t, fileNode, typeName)
	interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
	if !ok {
		t.Fatalf("%s must stay an interface", typeName)
	}

	blockedMethods := make(map[string]struct{}, len(methodNames))
	for _, methodName := range methodNames {
		blockedMethods[methodName] = struct{}{}
	}

	for _, field := range interfaceType.Methods.List {
		for _, name := range field.Names {
			if _, blocked := blockedMethods[name.Name]; blocked {
				t.Fatalf("%s must not declare method %s", typeName, name.Name)
			}
		}
	}
}

func findTypeSpec(t *testing.T, fileNode *ast.File, typeName string) *ast.TypeSpec {
	t.Helper()

	typeSpec, ok := lookupTypeSpec(fileNode, typeName)
	if ok {
		return typeSpec
	}

	t.Fatalf("type %s not found", typeName)
	return nil
}

func lookupTypeSpec(fileNode *ast.File, typeName string) (*ast.TypeSpec, bool) {
	for _, decl := range fileNode.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if typeSpec.Name.Name == typeName {
				return typeSpec, true
			}
		}
	}

	return nil, false
}
