package module

import (
	"go/ast"
	"go/parser"
	"go/token"
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
		if fileNeedsReviewedTimeNowException(t, file) {
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

func TestTimeNowUsageGuardRequiresUTCOnSameCallChain(t *testing.T) {
	t.Parallel()

	const source = `package sample

import "time"

func allowed() time.Time {
	return time.Now().Add(time.Hour).UTC()
}

func blocked(other time.Time) time.Time {
	_ = other.UTC()
	return time.Now()
}
`

	needsException, err := sourceNeedsReviewedTimeNowException(source)
	if err != nil {
		t.Fatalf("parse sample source: %v", err)
	}
	if !needsException {
		t.Fatalf("time.Now usage without UTC on the same call chain must still need a reviewed exception")
	}
}

func TestTimeNowUsageGuardAllowsRuntimeOnlyPatterns(t *testing.T) {
	t.Parallel()

	const source = `package sample

import (
	"fmt"
	"net"
	"time"
)

type result struct {
	StartedAt time.Time
	FinishedAt time.Time
	Duration time.Duration
}

func suffix() string {
	return fmt.Sprintf("tmp-%d", time.Now().UnixNano())
}

func duration() result {
	startedAt := time.Now()
	out := result{StartedAt: startedAt.UTC()}
	finishedAt := time.Now()
	out.FinishedAt = finishedAt.UTC()
	out.Duration = finishedAt.Sub(startedAt)
	return out
}

func probe() time.Duration {
	start := time.Now()
	return time.Since(start)
}

func track() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		return time.Since(start)
	}
}

func deadline(conn net.Conn, timeout time.Duration) bool {
	_ = conn.SetDeadline(time.Now().Add(timeout))
	deadline := time.Now().Add(timeout)
	return time.Now().After(deadline)
}
`

	needsException, err := sourceNeedsReviewedTimeNowException(source)
	if err != nil {
		t.Fatalf("parse sample source: %v", err)
	}
	if needsException {
		t.Fatalf("runtime-only time.Now usage should not need a reviewed exception")
	}
}

func TestTimeNowUsageGuardRejectsRawAssignedBusinessTime(t *testing.T) {
	t.Parallel()

	const source = `package sample

import "time"

type result struct {
	StartedAt time.Time
}

func blocked() time.Time {
	now := time.Now()
	return now
}

func blockedField() result {
	startedAt := time.Now()
	out := result{}
	out.StartedAt = startedAt
	return out
}
`

	needsException, err := sourceNeedsReviewedTimeNowException(source)
	if err != nil {
		t.Fatalf("parse sample source: %v", err)
	}
	if !needsException {
		t.Fatalf("raw assigned time.Now business value must still need a reviewed exception")
	}
}

func TestTimeNowUsageGuardRejectsWrappedAssignedRuntimeTime(t *testing.T) {
	t.Parallel()

	const source = `package sample

import "time"

func wrap(value time.Time) time.Time {
	return value
}

func duration() time.Duration {
	start := wrap(time.Now())
	return time.Since(start)
}

func deadline(timeout time.Duration) bool {
	deadline := wrap(time.Now().Add(timeout))
	return time.Now().After(deadline)
}
`

	needsException, err := sourceNeedsReviewedTimeNowException(source)
	if err != nil {
		t.Fatalf("parse sample source: %v", err)
	}
	if !needsException {
		t.Fatalf("wrapped time.Now values must not be inferred as local runtime-only assignments")
	}
}

func TestTimeNowUsageGuardRejectsClosureBusinessTime(t *testing.T) {
	t.Parallel()

	const source = `package sample

import "time"

func blocked() func() time.Time {
	startedAt := time.Now()
	return func() time.Time {
		return startedAt
	}
}
`

	needsException, err := sourceNeedsReviewedTimeNowException(source)
	if err != nil {
		t.Fatalf("parse sample source: %v", err)
	}
	if !needsException {
		t.Fatalf("closure-captured time.Now business value must still need a reviewed exception")
	}
}

func fileNeedsReviewedTimeNowException(t *testing.T, file string) bool {
	t.Helper()

	needsException, err := sourceNeedsReviewedTimeNowException(archtest.ReadFile(t, file))
	if err != nil {
		t.Fatalf("parse %s for time.Now usage: %v", file, err)
	}
	return needsException
}

func sourceNeedsReviewedTimeNowException(content string) (bool, error) {
	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, "", content, 0)
	if err != nil {
		return false, err
	}

	parents := make(map[ast.Node]ast.Node)
	var stack []ast.Node
	ast.Inspect(fileNode, func(node ast.Node) bool {
		if node == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if len(stack) > 0 {
			parents[node] = stack[len(stack)-1]
		}
		stack = append(stack, node)
		return true
	})

	timeNowAssignments := collectTimeNowAssignments(fileNode)
	identifierUses := collectIdentifierUses(fileNode)

	needsException := false
	ast.Inspect(fileNode, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || !isTimeNowCall(call) {
			return true
		}
		if !timeNowCallIsAllowed(call, parents, timeNowAssignments, identifierUses) {
			needsException = true
			return false
		}
		return true
	})
	return needsException, nil
}

func collectTimeNowAssignments(fileNode *ast.File) map[*ast.CallExpr]*ast.Object {
	assignments := make(map[*ast.CallExpr]*ast.Object)
	ast.Inspect(fileNode, func(node ast.Node) bool {
		assign, ok := node.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for idx, rhs := range assign.Rhs {
			if idx >= len(assign.Lhs) {
				continue
			}
			ident, ok := assign.Lhs[idx].(*ast.Ident)
			if !ok || ident.Obj == nil {
				continue
			}
			if call := directTimeNowCallInAssignedExpr(rhs); call != nil {
				assignments[call] = ident.Obj
			}
		}
		return true
	})
	return assignments
}

func collectIdentifierUses(fileNode *ast.File) map[*ast.Object][]*ast.Ident {
	uses := make(map[*ast.Object][]*ast.Ident)
	ast.Inspect(fileNode, func(node ast.Node) bool {
		ident, ok := node.(*ast.Ident)
		if !ok || ident.Obj == nil {
			return true
		}
		uses[ident.Obj] = append(uses[ident.Obj], ident)
		return true
	})
	return uses
}

func directTimeNowCallInAssignedExpr(expr ast.Expr) *ast.CallExpr {
	switch node := expr.(type) {
	case *ast.CallExpr:
		if isTimeNowCall(node) {
			return node
		}
		selector, ok := node.Fun.(*ast.SelectorExpr)
		if !ok {
			return nil
		}
		return directTimeNowCallInAssignedExpr(selector.X)
	case *ast.SelectorExpr:
		return directTimeNowCallInAssignedExpr(node.X)
	case *ast.ParenExpr:
		return directTimeNowCallInAssignedExpr(node.X)
	default:
		return nil
	}
}

func isTimeNowCall(call *ast.CallExpr) bool {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Now" {
		return false
	}
	ident, ok := selector.X.(*ast.Ident)
	return ok && ident.Name == "time"
}

func timeNowCallIsAllowed(call *ast.CallExpr, parents map[ast.Node]ast.Node, assignments map[*ast.CallExpr]*ast.Object, uses map[*ast.Object][]*ast.Ident) bool {
	if timeNowCallChainHasAllowedTerminal(call, parents) {
		return true
	}
	if timeNowCallFeedsSetDeadline(call, parents) {
		return true
	}
	if timeNowCallComparesDeadline(call, parents) {
		return true
	}
	if obj := assignments[call]; obj != nil && timeNowAssignedObjectIsRuntimeOnly(obj, uses[obj], parents) {
		return true
	}
	return false
}

func timeNowCallChainHasAllowedTerminal(call *ast.CallExpr, parents map[ast.Node]ast.Node) bool {
	var current ast.Node = call
	for {
		parent := parents[current]
		switch node := parent.(type) {
		case *ast.SelectorExpr:
			if node.X != current {
				return false
			}
			if node.Sel.Name == "UTC" || node.Sel.Name == "Unix" || node.Sel.Name == "UnixNano" {
				if utcCall, ok := parents[node].(*ast.CallExpr); ok && utcCall.Fun == node {
					return true
				}
			}
			current = node
		case *ast.CallExpr:
			if node.Fun != current {
				return false
			}
			current = node
		default:
			return false
		}
	}
}

func timeNowCallFeedsSetDeadline(call *ast.CallExpr, parents map[ast.Node]ast.Node) bool {
	terminal := terminalCallInTimeNowChain(call, parents)
	parent, ok := parents[terminal].(*ast.CallExpr)
	if !ok {
		return false
	}
	for _, arg := range parent.Args {
		if arg == terminal {
			selector, ok := parent.Fun.(*ast.SelectorExpr)
			return ok && selector.Sel.Name == "SetDeadline"
		}
	}
	return false
}

func timeNowCallComparesDeadline(call *ast.CallExpr, parents map[ast.Node]ast.Node) bool {
	terminal := terminalCallInTimeNowChain(call, parents)
	selector, ok := terminal.Fun.(*ast.SelectorExpr)
	if !ok || selector.X == nil || (selector.Sel.Name != "After" && selector.Sel.Name != "Before") {
		return false
	}
	if len(terminal.Args) != 1 {
		return false
	}
	ident, ok := terminal.Args[0].(*ast.Ident)
	return ok && strings.Contains(strings.ToLower(ident.Name), "deadline")
}

func terminalCallInTimeNowChain(call *ast.CallExpr, parents map[ast.Node]ast.Node) *ast.CallExpr {
	current := ast.Node(call)
	terminal := call
	for {
		parent := parents[current]
		switch node := parent.(type) {
		case *ast.SelectorExpr:
			if node.X != current {
				return terminal
			}
			current = node
		case *ast.CallExpr:
			if node.Fun != current {
				return terminal
			}
			terminal = node
			current = node
		default:
			return terminal
		}
	}
}

func timeNowAssignedObjectIsRuntimeOnly(obj *ast.Object, uses []*ast.Ident, parents map[ast.Node]ast.Node) bool {
	for _, use := range uses {
		if isDefinitionIdent(use, obj) || isAssignmentLHS(use, parents) {
			continue
		}
		if isRuntimeOnlyTimeIdentUse(use, parents) {
			continue
		}
		return false
	}
	return true
}

func isDefinitionIdent(ident *ast.Ident, obj *ast.Object) bool {
	return obj != nil && obj.Decl == ident
}

func isAssignmentLHS(ident *ast.Ident, parents map[ast.Node]ast.Node) bool {
	assign, ok := parents[ident].(*ast.AssignStmt)
	if !ok {
		return false
	}
	for _, lhs := range assign.Lhs {
		if lhs == ident {
			return true
		}
	}
	return false
}

func isRuntimeOnlyTimeIdentUse(ident *ast.Ident, parents map[ast.Node]ast.Node) bool {
	if identUsedAsUTCReceiver(ident, parents) {
		return true
	}
	if identUsedWithTimeSince(ident, parents) {
		return true
	}
	if identUsedWithSub(ident, parents) {
		return true
	}
	if identUsedAsDeadlineComparisonArg(ident, parents) {
		return true
	}
	return false
}

func identUsedAsUTCReceiver(ident *ast.Ident, parents map[ast.Node]ast.Node) bool {
	selector, ok := parents[ident].(*ast.SelectorExpr)
	if !ok || selector.X != ident || selector.Sel.Name != "UTC" {
		return false
	}
	call, ok := parents[selector].(*ast.CallExpr)
	return ok && call.Fun == selector
}

func identUsedWithTimeSince(ident *ast.Ident, parents map[ast.Node]ast.Node) bool {
	call, ok := parents[ident].(*ast.CallExpr)
	if !ok || len(call.Args) != 1 || call.Args[0] != ident {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Since" {
		return false
	}
	receiver, ok := selector.X.(*ast.Ident)
	return ok && receiver.Name == "time"
}

func identUsedWithSub(ident *ast.Ident, parents map[ast.Node]ast.Node) bool {
	if selector, ok := parents[ident].(*ast.SelectorExpr); ok && selector.X == ident && selector.Sel.Name == "Sub" {
		call, ok := parents[selector].(*ast.CallExpr)
		return ok && call.Fun == selector
	}
	call, ok := parents[ident].(*ast.CallExpr)
	if !ok {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || selector.Sel.Name != "Sub" {
		return false
	}
	for _, arg := range call.Args {
		if arg == ident {
			return true
		}
	}
	return false
}

func identUsedAsDeadlineComparisonArg(ident *ast.Ident, parents map[ast.Node]ast.Node) bool {
	if !strings.Contains(strings.ToLower(ident.Name), "deadline") {
		return false
	}
	call, ok := parents[ident].(*ast.CallExpr)
	if !ok || len(call.Args) != 1 || call.Args[0] != ident {
		return false
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	return ok && (selector.Sel.Name == "After" || selector.Sel.Name == "Before")
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
