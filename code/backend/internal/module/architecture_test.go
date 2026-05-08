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

var allowedApplicationConcreteImports = map[string]struct{}{
	"assessment/application/commands/profile_service.go -> github.com/redis/go-redis/v9":                 {},
	"assessment/application/commands/report_service.go -> gorm.io/gorm":                                  {},
	"assessment/application/queries/recommendation_service.go -> github.com/redis/go-redis/v9":           {},
	"auth/application/commands/cas_service.go -> net/http":                                               {},
	"challenge/application/commands/awd_challenge_command_facade.go -> gorm.io/gorm":                     {},
	"challenge/application/commands/awd_challenge_import_service.go -> gorm.io/gorm":                     {},
	"challenge/application/commands/awd_challenge_service.go -> gorm.io/gorm":                            {},
	"challenge/application/commands/challenge_import_service.go -> gorm.io/gorm":                         {},
	"challenge/application/commands/challenge_package_revision_service.go -> gorm.io/gorm":               {},
	"challenge/application/commands/challenge_package_revision_service.go -> gorm.io/gorm/clause":        {},
	"challenge/application/commands/challenge_service.go -> gorm.io/gorm":                                {},
	"challenge/application/commands/flag_service.go -> gorm.io/gorm":                                     {},
	"challenge/application/commands/image_build_service.go -> gorm.io/gorm":                              {},
	"challenge/application/commands/image_service.go -> gorm.io/gorm":                                    {},
	"challenge/application/commands/registry_client.go -> net/http":                                      {},
	"challenge/application/commands/topology_service.go -> gorm.io/gorm":                                 {},
	"challenge/application/commands/writeup_service.go -> gorm.io/gorm":                                  {},
	"challenge/application/queries/awd_challenge_service.go -> gorm.io/gorm":                             {},
	"challenge/application/queries/challenge_service.go -> github.com/redis/go-redis/v9":                 {},
	"challenge/application/queries/challenge_service.go -> gorm.io/gorm":                                 {},
	"challenge/application/queries/flag_service.go -> gorm.io/gorm":                                      {},
	"challenge/application/queries/image_service.go -> gorm.io/gorm":                                     {},
	"challenge/application/queries/topology_service.go -> gorm.io/gorm":                                  {},
	"challenge/application/queries/writeup_service.go -> gorm.io/gorm":                                   {},
	"contest/application/commands/awd_attack_log_commands.go -> gorm.io/gorm":                            {},
	"contest/application/commands/awd_checker_preview_token_support.go -> github.com/redis/go-redis/v9":  {},
	"contest/application/commands/awd_current_round_active_support.go -> gorm.io/gorm":                   {},
	"contest/application/commands/awd_current_round_fallback_support.go -> github.com/redis/go-redis/v9": {},
	"contest/application/commands/awd_current_round_fallback_support.go -> gorm.io/gorm":                 {},
	"contest/application/commands/awd_flag_support.go -> github.com/redis/go-redis/v9":                   {},
	"contest/application/commands/awd_flag_support.go -> gorm.io/gorm":                                   {},
	"contest/application/commands/awd_preview_runtime_support.go -> gorm.io/gorm":                        {},
	"contest/application/commands/awd_resource_validation_support.go -> gorm.io/gorm":                    {},
	"contest/application/commands/awd_service.go -> github.com/redis/go-redis/v9":                        {},
	"contest/application/commands/awd_status_cache.go -> github.com/redis/go-redis/v9":                   {},
	"contest/application/commands/awd_team_validation_support.go -> gorm.io/gorm":                        {},
	"contest/application/commands/awd_validation_support.go -> gorm.io/gorm":                             {},
	"contest/application/commands/challenge_add_commands.go -> gorm.io/gorm":                             {},
	"contest/application/commands/challenge_service.go -> github.com/redis/go-redis/v9":                  {},
	"contest/application/commands/contest_awd_service_service.go -> github.com/redis/go-redis/v9":        {},
	"contest/application/commands/contest_awd_service_service.go -> gorm.io/gorm":                        {},
	"contest/application/commands/contest_awd_service_support.go -> github.com/redis/go-redis/v9":        {},
	"contest/application/commands/contest_service.go -> github.com/redis/go-redis/v9":                    {},
	"contest/application/commands/participation_register_commands.go -> gorm.io/gorm":                    {},
	"contest/application/commands/participation_review_commands.go -> gorm.io/gorm":                      {},
	"contest/application/commands/scoreboard_admin_score_commands.go -> github.com/redis/go-redis/v9":    {},
	"contest/application/commands/scoreboard_admin_service.go -> github.com/redis/go-redis/v9":           {},
	"contest/application/commands/submission_service.go -> github.com/redis/go-redis/v9":                 {},
	"contest/application/commands/submission_submit_validation.go -> gorm.io/gorm":                       {},
	"contest/application/commands/submission_validation.go -> gorm.io/gorm":                              {},
	"contest/application/commands/team_captain_manage_commands.go -> gorm.io/gorm":                       {},
	"contest/application/commands/team_create_retry_support.go -> gorm.io/gorm":                          {},
	"contest/application/commands/team_join_commands.go -> gorm.io/gorm":                                 {},
	"contest/application/commands/team_leave_commands.go -> gorm.io/gorm":                                {},
	"contest/application/commands/team_support.go -> gorm.io/gorm":                                       {},
	"contest/application/jobs/awd_check_cache_support.go -> github.com/redis/go-redis/v9":                {},
	"contest/application/jobs/awd_check_cache_support.go -> gorm.io/gorm":                                {},
	"contest/application/jobs/awd_http_checker_request.go -> net/http":                                   {},
	"contest/application/jobs/awd_http_target_client.go -> net/http":                                     {},
	"contest/application/jobs/awd_probe_runtime.go -> net/http":                                          {},
	"contest/application/jobs/awd_round_flag_lookup_support.go -> github.com/redis/go-redis/v9":          {},
	"contest/application/jobs/awd_round_flag_lookup_support.go -> gorm.io/gorm":                          {},
	"contest/application/jobs/awd_round_runtime.go -> gorm.io/gorm":                                      {},
	"contest/application/jobs/awd_round_runtime_bridge.go -> net/http":                                   {},
	"contest/application/jobs/awd_round_updater.go -> github.com/redis/go-redis/v9":                      {},
	"contest/application/jobs/awd_round_updater.go -> net/http":                                          {},
	"contest/application/jobs/status_updater.go -> github.com/redis/go-redis/v9":                         {},
	"contest/application/queries/awd_support.go -> gorm.io/gorm":                                         {},
	"contest/application/queries/awd_workspace_query.go -> gorm.io/gorm":                                 {},
	"contest/application/queries/participation_progress_query.go -> gorm.io/gorm":                        {},
	"contest/application/queries/scoreboard_list_support.go -> github.com/redis/go-redis/v9":             {},
	"contest/application/queries/scoreboard_rank_query.go -> github.com/redis/go-redis/v9":               {},
	"contest/application/queries/scoreboard_service.go -> github.com/redis/go-redis/v9":                  {},
	"contest/application/queries/scoreboard_support.go -> github.com/redis/go-redis/v9":                  {},
	"contest/application/queries/team_info_query.go -> gorm.io/gorm":                                     {},
	"contest/application/queries/team_list_query.go -> gorm.io/gorm":                                     {},
	"contest/application/statusmachine/side_effects.go -> github.com/redis/go-redis/v9":                  {},
	"ops/application/commands/notification_service.go -> gorm.io/gorm":                                   {},
	"ops/application/queries/dashboard_service.go -> github.com/redis/go-redis/v9":                       {},
	"practice/application/commands/contest_awd_operations.go -> gorm.io/gorm":                            {},
	"practice/application/commands/contest_instance_scope.go -> gorm.io/gorm":                            {},
	"practice/application/commands/instance_provisioning.go -> net/http":                                 {},
	"practice/application/commands/manual_review_service.go -> gorm.io/gorm":                             {},
	"practice/application/commands/score_service.go -> github.com/redis/go-redis/v9":                     {},
	"practice/application/commands/service.go -> github.com/redis/go-redis/v9":                           {},
	"practice/application/commands/submission_service.go -> gorm.io/gorm":                                {},
	"practice/application/queries/score_service.go -> github.com/redis/go-redis/v9":                      {},
	"practice/application/queries/score_service.go -> gorm.io/gorm":                                      {},
	"practice_readmodel/application/queries/service.go -> github.com/redis/go-redis/v9":                  {},
}

var allowedCrossModulePrivateImports = map[string]struct{}{
	"contest/infrastructure/docker_checker_runner.go -> ctf-platform/internal/module/runtime/domain":                {},
	"practice/application/commands/awd_defense_workspace_support.go -> ctf-platform/internal/module/contest/domain": {},
}

func TestModuleArchitectureBoundaries(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	for _, file := range files {
		layer := moduleLayer(file)
		imports := parseImports(t, file)
		assertNoCrossModulePrivateImports(t, file, imports)

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
			assertApplicationConcreteImportsAreAllowlisted(t, file, imports)
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

func TestApplicationConcreteDependencyAllowlistIsCurrent(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		if moduleLayer(file) != "application" {
			continue
		}
		for _, importPath := range parseImports(t, file) {
			if isConcreteApplicationImport(importPath) {
				actual[applicationConcreteImportKey(file, importPath)] = struct{}{}
			}
		}
	}

	for allowed := range allowedApplicationConcreteImports {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("application concrete dependency allowlist entry is stale: %s", allowed)
		}
	}
}

func TestCrossModulePrivateImportAllowlistIsCurrent(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		for _, importPath := range parseImports(t, file) {
			if isCrossModulePrivateImport(file, importPath) {
				actual[crossModuleImportKey(file, importPath)] = struct{}{}
			}
		}
	}

	for allowed := range allowedCrossModulePrivateImports {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("cross-module private import allowlist entry is stale: %s", allowed)
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

func moduleOwner(filePath string) string {
	parts := strings.Split(filepath.ToSlash(filePath), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
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

func assertNoCrossModulePrivateImports(t *testing.T, filePath string, imports []string) {
	t.Helper()

	for _, importPath := range imports {
		if !isCrossModulePrivateImport(filePath, importPath) {
			continue
		}
		key := crossModuleImportKey(filePath, importPath)
		if _, allowed := allowedCrossModulePrivateImports[key]; !allowed {
			t.Fatalf("%s must not import private layer from another module: %s", filePath, importPath)
		}
	}
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

func assertApplicationConcreteImportsAreAllowlisted(t *testing.T, filePath string, imports []string) {
	t.Helper()

	for _, importPath := range imports {
		if !isConcreteApplicationImport(importPath) {
			continue
		}
		key := applicationConcreteImportKey(filePath, importPath)
		if _, allowed := allowedApplicationConcreteImports[key]; !allowed {
			t.Fatalf("%s imports concrete dependency %s; add a port/infrastructure adapter instead of growing the allowlist", filePath, importPath)
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
