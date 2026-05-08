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

var allowedDomainInternalImports = map[string]struct{}{
	"assessment/domain/profile.go -> ctf-platform/internal/config":                    {},
	"assessment/domain/profile.go -> ctf-platform/internal/dto":                       {},
	"assessment/domain/profile.go -> ctf-platform/internal/model":                     {},
	"assessment/domain/recommendation.go -> ctf-platform/internal/config":             {},
	"assessment/domain/report.go -> ctf-platform/internal/config":                     {},
	"assessment/domain/report.go -> ctf-platform/internal/model":                      {},
	"challenge/domain/awd_package_parser.go -> ctf-platform/internal/model":           {},
	"challenge/domain/image_delivery.go -> ctf-platform/internal/model":               {},
	"challenge/domain/mappers.go -> ctf-platform/internal/dto":                        {},
	"challenge/domain/mappers.go -> ctf-platform/internal/model":                      {},
	"challenge/domain/package_parser.go -> ctf-platform/internal/model":               {},
	"challenge/domain/package_topology_parser.go -> ctf-platform/internal/dto":        {},
	"challenge/domain/package_topology_parser.go -> ctf-platform/internal/model":      {},
	"challenge/domain/response_mapper_goverter.go -> ctf-platform/internal/dto":       {},
	"challenge/domain/response_mapper_goverter.go -> ctf-platform/internal/model":     {},
	"challenge/domain/response_mapper_goverter_gen.go -> ctf-platform/internal/dto":   {},
	"challenge/domain/response_mapper_goverter_gen.go -> ctf-platform/internal/model": {},
	"challenge/domain/topology_codec.go -> ctf-platform/internal/dto":                 {},
	"challenge/domain/topology_codec.go -> ctf-platform/internal/model":               {},
	"contest/domain/awd_checker_validation_support.go -> ctf-platform/internal/model": {},
	"contest/domain/awd_service_config.go -> ctf-platform/internal/model":             {},
	"contest/domain/awd_source_support.go -> ctf-platform/internal/model":             {},
	"contest/domain/contest.go -> ctf-platform/internal/model":                        {},
	"contest/domain/registration.go -> ctf-platform/internal/model":                   {},
	"practice/domain/mappers.go -> ctf-platform/internal/dto":                         {},
	"practice/domain/mappers.go -> ctf-platform/internal/model":                       {},
	"practice/domain/response_mapper_goverter.go -> ctf-platform/internal/dto":        {},
	"practice/domain/response_mapper_goverter.go -> ctf-platform/internal/model":      {},
	"practice/domain/response_mapper_goverter_gen.go -> ctf-platform/internal/dto":    {},
	"practice/domain/response_mapper_goverter_gen.go -> ctf-platform/internal/model":  {},
	"practice/domain/score.go -> ctf-platform/internal/model":                         {},
	"practice/domain/topology_runtime.go -> ctf-platform/internal/model":              {},
	"runtime/domain/resources.go -> ctf-platform/internal/model":                      {},
	"runtime/domain/topology_acl.go -> ctf-platform/internal/model":                   {},
}

var allowedModuleDependencies = map[string]struct{}{
	"assessment -> contest":            {},
	"assessment -> practice":           {},
	"auth -> identity":                 {},
	"contest -> auth":                  {},
	"contest -> challenge":             {},
	"contest -> runtime":               {},
	"identity -> auth":                 {},
	"ops -> auth":                      {},
	"ops -> practice":                  {},
	"practice -> assessment":           {},
	"practice -> challenge":            {},
	"practice -> contest":              {},
	"practice -> runtime":              {},
	"runtime -> challenge":             {},
	"runtime -> contest":               {},
	"runtime -> ops":                   {},
	"runtime -> practice":              {},
	"teaching_readmodel -> assessment": {},
}

var allowedTransactionFiles = map[string]struct{}{
	"challenge/application/commands/awd_challenge_import_service.go":       {},
	"challenge/application/commands/challenge_import_service.go":           {},
	"challenge/application/commands/challenge_package_revision_service.go": {},
	"challenge/infrastructure/repository.go":                               {},
	"challenge/infrastructure/tag_repository.go":                           {},
	"contest/infrastructure/awd_repository.go":                             {},
	"contest/infrastructure/contest_status_update_repository.go":           {},
	"contest/infrastructure/submission_repository.go":                      {},
	"contest/infrastructure/team_membership_lifecycle_repository.go":       {},
	"contest/infrastructure/team_membership_repository.go":                 {},
	"identity/infrastructure/repository.go":                                {},
	"ops/infrastructure/notification_repository.go":                        {},
	"practice/infrastructure/repository.go":                                {},
	"runtime/infrastructure/repository.go":                                 {},
}

var allowedOversizedRuntimeModules = map[string]struct{}{
	"challenge/runtime/module.go": {},
	"runtime/runtime/module.go":   {},
}

var allowedTimeNowFiles = map[string]struct{}{
	"assessment/application/commands/profile_service.go":                   {},
	"assessment/application/commands/report_service.go":                    {},
	"assessment/application/queries/teacher_awd_review_service.go":         {},
	"assessment/infrastructure/report_repository.go":                       {},
	"assessment/infrastructure/repository.go":                              {},
	"auth/application/commands/cas_service.go":                             {},
	"auth/application/commands/service.go":                                 {},
	"auth/infrastructure/token_service.go":                                 {},
	"challenge/application/commands/awd_challenge_import_service.go":       {},
	"challenge/application/commands/challenge_import_service.go":           {},
	"challenge/application/commands/challenge_package_revision_service.go": {},
	"challenge/application/commands/challenge_service.go":                  {},
	"challenge/application/commands/image_build_service.go":                {},
	"challenge/application/commands/topology_service.go":                   {},
	"challenge/application/commands/writeup_service.go":                    {},
	"challenge/application/queries/writeup_service.go":                     {},
	"contest/application/commands/awd_attack_log_transaction.go":           {},
	"contest/application/commands/awd_attack_submit_support.go":            {},
	"contest/application/commands/awd_checker_preview_token_support.go":    {},
	"contest/application/commands/awd_current_round_support.go":            {},
	"contest/application/commands/awd_round_window_support.go":             {},
	"contest/application/commands/awd_service_run_commands.go":             {},
	"contest/application/commands/awd_service_upsert_commands.go":          {},
	"contest/application/commands/contest_awd_service_service.go":          {},
	"contest/application/commands/contest_update_commands.go":              {},
	"contest/application/commands/participation_announcement_commands.go":  {},
	"contest/application/commands/participation_register_commands.go":      {},
	"contest/application/commands/participation_review_commands.go":        {},
	"contest/application/commands/realtime_broadcast.go":                   {},
	"contest/application/commands/scoreboard_admin_freeze_commands.go":     {},
	"contest/application/commands/submission_submit_validation.go":         {},
	"contest/application/jobs/awd_check_run.go":                            {},
	"contest/application/jobs/awd_checker_preview.go":                      {},
	"contest/application/jobs/awd_http_checker_runner.go":                  {},
	"contest/application/jobs/awd_probe_runtime.go":                        {},
	"contest/application/jobs/awd_round_updater.go":                        {},
	"contest/application/jobs/awd_script_checker_runner.go":                {},
	"contest/application/jobs/awd_service_check_result.go":                 {},
	"contest/application/jobs/awd_tcp_checker_runner.go":                   {},
	"contest/application/jobs/status_transition_service.go":                {},
	"contest/application/jobs/status_update_runner.go":                     {},
	"contest/application/queries/scoreboard_list_query.go":                 {},
	"contest/domain/awd_check_result_support.go":                           {},
	"contest/infrastructure/awd_round_repository.go":                       {},
	"contest/infrastructure/contest_repository.go":                         {},
	"contest/infrastructure/docker_checker_runner.go":                      {},
	"contest/infrastructure/team_membership_lifecycle_repository.go":       {},
	"contest/infrastructure/team_membership_repository.go":                 {},
	"contest/infrastructure/team_registration_binding.go":                  {},
	"identity/infrastructure/repository.go":                                {},
	"ops/application/commands/notification_service.go":                     {},
	"ops/application/queries/risk_service.go":                              {},
	"practice/application/commands/contest_awd_operations.go":              {},
	"practice/application/commands/instance_provisioning.go":               {},
	"practice/application/commands/instance_start_service.go":              {},
	"practice/application/commands/manual_review_service.go":               {},
	"practice/application/commands/score_service.go":                       {},
	"practice/application/commands/submission_service.go":                  {},
	"practice/infrastructure/repository.go":                                {},
	"runtime/application/commands/instance_service.go":                     {},
	"runtime/application/commands/provisioning_service.go":                 {},
	"runtime/application/commands/runtime_maintenance_service.go":          {},
	"runtime/application/queries/instance_service.go":                      {},
	"runtime/application/queries/proxy_ticket_service.go":                  {},
	"runtime/infrastructure/repository.go":                                 {},
	"runtime/runtime/adapters.go":                                          {},
	"teaching_readmodel/application/queries/service.go":                    {},
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
			assertDomainInternalImportsAreAllowlisted(t, file, imports)
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

func TestModuleDependencyAllowlistIsCurrent(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		for _, importPath := range parseImports(t, file) {
			if key, ok := moduleDependencyKey(file, importPath); ok {
				actual[key] = struct{}{}
				if _, allowed := allowedModuleDependencies[key]; !allowed {
					t.Fatalf("module dependency is not allowlisted: %s via %s", key, file)
				}
			}
		}
	}

	for allowed := range allowedModuleDependencies {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("module dependency allowlist entry is stale: %s", allowed)
		}
	}
}

func TestDomainInternalImportAllowlistIsCurrent(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		if moduleLayer(file) != "domain" {
			continue
		}
		for _, importPath := range parseImports(t, file) {
			if isDomainInternalImport(importPath) {
				actual[domainInternalImportKey(file, importPath)] = struct{}{}
			}
		}
	}

	for allowed := range allowedDomainInternalImports {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("domain internal import allowlist entry is stale: %s", allowed)
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

func TestModuleRuntimeCodeDoesNotCreateRootContext(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	for _, file := range files {
		content := readFile(t, file)
		if strings.Contains(content, "context.Background()") || strings.Contains(content, "context.TODO()") {
			t.Fatalf("%s must receive context from its caller instead of creating a root context", file)
		}
	}
}

func TestTimeNowUsageAllowlistIsCurrent(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		if strings.Contains(readFile(t, file), "time.Now(") {
			actual[moduleFileKey(file)] = struct{}{}
		}
	}

	for file := range actual {
		if _, allowed := allowedTimeNowFiles[file]; !allowed {
			t.Fatalf("%s uses time.Now; use UTC business time or add a reviewed allowlist entry", file)
		}
	}
	for allowed := range allowedTimeNowFiles {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("time.Now allowlist entry is stale: %s", allowed)
		}
	}
}

func TestTransactionBoundaryAllowlistIsCurrent(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	actual := make(map[string]struct{})
	for _, file := range files {
		if strings.Contains(readFile(t, file), ".Transaction(") {
			actual[file] = struct{}{}
		}
	}

	for file := range actual {
		if _, allowed := allowedTransactionFiles[file]; !allowed {
			t.Fatalf("%s opens a transaction outside the reviewed boundary allowlist", file)
		}
	}
	for allowed := range allowedTransactionFiles {
		if _, exists := actual[allowed]; !exists {
			t.Fatalf("transaction allowlist entry is stale: %s", allowed)
		}
	}
}

func TestRuntimeModulesStaySmallAndWiringOnly(t *testing.T) {
	t.Parallel()

	files := collectGoRuntimeFiles(t, ".")
	for _, file := range files {
		if !strings.HasSuffix(filepath.ToSlash(file), "/runtime/module.go") {
			continue
		}
		lineCount := len(strings.Split(readFile(t, file), "\n"))
		if lineCount <= 250 {
			continue
		}
		if _, allowed := allowedOversizedRuntimeModules[file]; !allowed {
			t.Fatalf("%s has %d lines; runtime module files should stay wiring-only", file, lineCount)
		}
	}
	for allowed := range allowedOversizedRuntimeModules {
		content := readFile(t, allowed)
		if len(strings.Split(content, "\n")) <= 250 {
			t.Fatalf("runtime module size allowlist entry is stale: %s", allowed)
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

func readFile(t *testing.T, filePath string) string {
	t.Helper()

	content, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("read %s: %v", filePath, err)
	}
	return string(content)
}

func moduleFileKey(filePath string) string {
	return filepath.ToSlash(filePath)
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

func assertDomainInternalImportsAreAllowlisted(t *testing.T, filePath string, imports []string) {
	t.Helper()

	for _, importPath := range imports {
		if !isDomainInternalImport(importPath) {
			continue
		}
		key := domainInternalImportKey(filePath, importPath)
		if _, allowed := allowedDomainInternalImports[key]; !allowed {
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
