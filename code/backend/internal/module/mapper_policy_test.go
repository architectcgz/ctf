package module

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"ctf-platform/internal/testutil/archtest"
)

var allowedManualMapperWrapperFiles = map[string]string{
	"runtime/api/http/teacher_instance_mapper.go":                         "runtime teacher instance http query 需要轻量定制映射",
	"teaching_query/api/http/request_mapper.go":                           "teaching query request 需按 query 参数规则做裁剪与归一",
	"teaching_query/application/queries/class_insight_response_mapper.go": "class insight 聚合响应包含业务整形逻辑",
}

func TestMapperWrappersFollowGlobalDelegationPolicy(t *testing.T) {
	t.Parallel()

	files := archtest.RuntimeGoFiles(t, ".")
	manualWrapperFiles := make(map[string]struct{})

	for _, file := range files {
		normalized := filepath.ToSlash(file)
		if !isMapperSourceFile(normalized) {
			continue
		}

		info := parseMapperPolicyFile(t, normalized)
		if !info.hasToWrapper {
			continue
		}

		if info.hasGoverterConverter {
			assertGoverterWrapperDelegation(t, normalized, info)
			continue
		}

		if _, allowed := allowedManualMapperWrapperFiles[normalized]; !allowed {
			t.Fatalf("%s contains to* mapper wrappers but is not goverter-backed; convert to goverter or add reviewed allowlist entry", normalized)
		}
		manualWrapperFiles[normalized] = struct{}{}
	}

	for allowed := range allowedManualMapperWrapperFiles {
		if _, exists := manualWrapperFiles[allowed]; !exists {
			t.Fatalf("manual mapper wrapper allowlist entry is stale or no longer needed: %s", allowed)
		}
	}
}

type mapperPolicyFileInfo struct {
	hasGoverterConverter bool
	hasToWrapper         bool
	mapperVarName        string
	mapperMethods        map[string]struct{}
	toWrappers           []string
}

func parseMapperPolicyFile(t *testing.T, filePath string) mapperPolicyFileInfo {
	t.Helper()

	content := archtest.ReadFile(t, filePath)
	info := mapperPolicyFileInfo{
		hasGoverterConverter: strings.Contains(content, "goverter:converter"),
		mapperMethods:        make(map[string]struct{}),
	}

	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("parse mapper file %s: %v", filePath, err)
	}

	interfaceNames := make(map[string]struct{})
	for _, decl := range fileNode.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			iface, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				continue
			}
			interfaceNames[typeSpec.Name.Name] = struct{}{}
			for _, field := range iface.Methods.List {
				for _, name := range field.Names {
					info.mapperMethods[name.Name] = struct{}{}
				}
			}
		}
	}

	for _, decl := range fileNode.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if ok && gen.Tok == token.VAR {
			for _, spec := range gen.Specs {
				valueSpec, ok := spec.(*ast.ValueSpec)
				if !ok || len(valueSpec.Names) != 1 {
					continue
				}
				typeIdent, ok := valueSpec.Type.(*ast.Ident)
				if !ok {
					continue
				}
				if _, isInterface := interfaceNames[typeIdent.Name]; !isInterface {
					continue
				}
				if strings.Contains(valueSpec.Names[0].Name, "Mapper") {
					info.mapperVarName = valueSpec.Names[0].Name
				}
			}
		}

		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil {
			continue
		}
		if strings.HasPrefix(fn.Name.Name, "to") && len(fn.Name.Name) > 2 && isUpper(fn.Name.Name[2]) {
			info.hasToWrapper = true
			info.toWrappers = append(info.toWrappers, fn.Name.Name)
		}
	}

	slices.Sort(info.toWrappers)
	return info
}

func assertGoverterWrapperDelegation(t *testing.T, filePath string, info mapperPolicyFileInfo) {
	t.Helper()

	if info.mapperVarName == "" {
		t.Fatalf("%s is goverter-backed but mapper var not found", filePath)
	}
	if len(info.toWrappers) == 0 {
		return
	}

	fset := token.NewFileSet()
	fileNode, err := parser.ParseFile(fset, filePath, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("parse mapper file %s: %v", filePath, err)
	}

	for _, wrapperName := range info.toWrappers {
		fn := findFuncDecl(fileNode, wrapperName)
		if fn == nil || fn.Body == nil {
			t.Fatalf("%s in %s must have function body", wrapperName, filePath)
		}
		if len(fn.Body.List) != 1 {
			t.Fatalf("%s in %s must contain exactly one statement", wrapperName, filePath)
		}

		ret, ok := fn.Body.List[0].(*ast.ReturnStmt)
		if !ok || len(ret.Results) != 1 {
			t.Fatalf("%s in %s must contain a single return expression", wrapperName, filePath)
		}

		call, ok := ret.Results[0].(*ast.CallExpr)
		if !ok {
			t.Fatalf("%s in %s must return mapper delegation call", wrapperName, filePath)
		}
		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			t.Fatalf("%s in %s must call mapper method", wrapperName, filePath)
		}
		mapperIdent, ok := sel.X.(*ast.Ident)
		if !ok || mapperIdent.Name != info.mapperVarName {
			t.Fatalf("%s in %s must delegate via %s", wrapperName, filePath, info.mapperVarName)
		}
		if sel.Sel == nil {
			t.Fatalf("%s in %s must call explicit mapper method", wrapperName, filePath)
		}
		if _, exists := info.mapperMethods[sel.Sel.Name]; !exists {
			t.Fatalf("%s in %s delegates to unknown mapper method %s", wrapperName, filePath, sel.Sel.Name)
		}
		if len(call.Args) != 1 {
			t.Fatalf("%s in %s must pass exactly one source argument", wrapperName, filePath)
		}
		arg, ok := call.Args[0].(*ast.Ident)
		if !ok || arg.Name != "source" {
			t.Fatalf("%s in %s must pass source directly to mapper", wrapperName, filePath)
		}
	}
}

func findFuncDecl(fileNode *ast.File, name string) *ast.FuncDecl {
	for _, decl := range fileNode.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name == nil {
			continue
		}
		if fn.Name.Name == name {
			return fn
		}
	}
	return nil
}

func isMapperSourceFile(path string) bool {
	if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
		return false
	}
	if strings.HasSuffix(path, "_gen.go") || strings.HasSuffix(path, "_assign.go") {
		return false
	}
	name := filepath.Base(path)
	return strings.Contains(name, "mapper")
}

func isUpper(ch byte) bool {
	return ch >= 'A' && ch <= 'Z'
}
