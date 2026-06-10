package ops

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

func TestRootPackageKeepsNoConcreteGoFiles(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob ops root files: %v", err)
	}

	nonTestFiles := make([]string, 0, len(files))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		nonTestFiles = append(nonTestFiles, file)
	}

	if len(nonTestFiles) != 0 {
		t.Fatalf("ops root package should keep no non-test go files, got %v", nonTestFiles)
	}
}

func TestAPIHTTPDoesNotDependOnInfrastructure(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("api", "http", "*.go"))
	if err != nil {
		t.Fatalf("glob api/http files: %v", err)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/ops/infrastructure")
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
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/ops/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/ops/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/ops/application/queries")
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
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/ops/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/ops/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/ops/application/commands")
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
	}
}

func TestPortsDoNotDependOnDTOGinOrGORM(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("ports", "*.go"))
	if err != nil {
		t.Fatalf("glob ports files: %v", err)
	}
	for _, file := range files {
		assertFileDoesNotImport(t, file, "ctf-platform/internal/dto")
		assertFileDoesNotImport(t, file, "github.com/gin-gonic/gin")
		assertFileDoesNotImport(t, file, "gorm.io/gorm")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/ops/api/http")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/ops/infrastructure")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/ops/application/commands")
		assertFileDoesNotImport(t, file, "ctf-platform/internal/module/ops/application/queries")
	}
}

func TestRuntimeOwnsOpsWiring(t *testing.T) {
	t.Parallel()

	runtimeFile := filepath.Join("runtime", "module.go")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/ops/infrastructure")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/ops/application/commands")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/ops/application/queries")
	assertFileImports(t, runtimeFile, "ctf-platform/internal/module/ops/api/http")
}

func TestRuntimeUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("runtime", "module.go"))
	if err != nil {
		t.Fatalf("read ops runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"type moduleDeps struct",
		"auditRepo        opsports.AuditRepository",
		"riskRepo         opsports.RiskRepository",
		"notificationRepo opsports.NotificationRepository",
		"runtimeQuery          opsports.RuntimeQuery",
		"runtimeStats          opsports.RuntimeStatsProvider",
		"webSocketManager      *websocketpkg.Manager",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("ops runtime should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"type opsModuleDeps struct",
		"type opsNotificationDeps struct",
		"buildOpsModuleDeps(",
		"buildOpsNotificationDeps(",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("ops runtime should not keep composition glue marker %s", marker)
		}
	}
}

func TestRuntimeDelegatesThroughSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("runtime", "module.go"))
	if err != nil {
		t.Fatalf("read ops runtime module: %v", err)
	}

	source := string(content)
	expected := []string{
		"newModuleDeps(",
		"buildAuditHandler(",
		"buildDashboardHandler(",
		"buildRiskHandler(",
		"buildNotificationHandler(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("ops runtime should delegate through %s", marker)
		}
	}
}

func TestNotificationHTTPResponseMapperUsesGoverterDelegationOnly(t *testing.T) {
	t.Parallel()

	mapperFile := filepath.Join("api", "http", "notification_response_mapper.go")
	assertFileContains(t, mapperFile, "type notificationHTTPResponseMapper interface")
	assertFileContains(t, mapperFile, "var notificationResponseMapper notificationHTTPResponseMapper")

	assertMapperDelegatesOnly(t, mapperFile, "toNotificationInfo", "notificationResponseMapper", "ToNotificationInfo")
	assertMapperDelegatesOnly(t, mapperFile, "toNotificationInfos", "notificationResponseMapper", "ToNotificationInfos")
	assertMapperDelegatesOnly(t, mapperFile, "toAdminNotificationPublishResp", "notificationResponseMapper", "ToAdminNotificationPublishRespPtr")
}

func assertFileContains(t *testing.T, filePath string, marker string) {
	t.Helper()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read file %s: %v", filePath, err)
	}
	if !strings.Contains(string(content), marker) {
		t.Fatalf("%s must contain marker %q", filePath, marker)
	}
}

func assertMapperDelegatesOnly(t *testing.T, filePath string, funcName string, mapperVar string, mapperMethod string) {
	t.Helper()

	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("parse file %s: %v", filePath, err)
	}

	for _, decl := range fileNode.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil || fn.Name.Name != funcName {
			continue
		}

		if fn.Body == nil || len(fn.Body.List) != 1 {
			t.Fatalf("%s in %s must contain exactly one statement", funcName, filePath)
		}
		ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 {
			t.Fatalf("%s in %s must contain exactly one return value", funcName, filePath)
		}

		call, ok := ret.Results[0].(*ast.CallExpr)
		if !ok {
			t.Fatalf("%s in %s must return a mapper call", funcName, filePath)
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			t.Fatalf("%s in %s must call mapper selector", funcName, filePath)
		}
		mapperIdent, ok := sel.X.(*ast.Ident)
		if !ok || mapperIdent.Name != mapperVar {
			t.Fatalf("%s in %s must call %s", funcName, filePath, mapperVar)
		}
		if sel.Sel == nil || sel.Sel.Name != mapperMethod {
			t.Fatalf("%s in %s must call mapper method %s", funcName, filePath, mapperMethod)
		}
		if len(call.Args) != 1 {
			t.Fatalf("%s in %s must pass exactly one argument", funcName, filePath)
		}
		arg, ok := call.Args[0].(*ast.Ident)
		if !ok || arg.Name != "source" {
			t.Fatalf("%s in %s must pass source directly", funcName, filePath)
		}
		return
	}

	t.Fatalf("function %s not found in %s", funcName, filePath)
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
