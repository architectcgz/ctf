package architecture

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const defaultInternalAppSystemShellMaxLines = 250
const backendRoot = "../.."

var internalAppSystemShellOwnerMarkers = map[string]string{
	"internal/app/practice_flow_scenario_test.go":                    "tests/system/http/practiceflow",
	"internal/app/practice_flow_lifecycle_integration_test.go":       "tests/system/http/practiceflow",
	"internal/app/practice_flow_observability_integration_test.go":   "tests/system/http/practiceflow",
	"internal/app/full_router_access_integration_test.go":            "tests/system/http/fullrouteraccess",
	"internal/app/full_router_admin_integration_test.go":             "tests/system/http/fullrouteradmin",
	"internal/app/full_router_admin_state_matrix_test.go":            "tests/system/http/fullrouteradmin",
	"internal/app/full_router_awd_state_matrix_test.go":              "tests/system/http/fullrouterawdstate",
	"internal/app/full_router_contest_state_matrix_test.go":          "tests/system/http/fullrouterconteststate",
	"internal/app/full_router_state_matrix_integration_test.go":      "tests/system/http/fullrouterreportstate",
	"internal/app/full_router_teacher_authoring_integration_test.go": "tests/system/http/fullrouterteacherauthoring",
	"internal/app/full_router_teacher_state_matrix_test.go":          "tests/system/http/fullrouterteacherstate",
}

var allowedOversizedInternalAppSystemShells = map[string]int{
	"internal/app/full_router_teacher_state_matrix_test.go":     380,
	"internal/app/full_router_state_matrix_integration_test.go": 560,
}

var forbiddenSystemHTTPSnippets = []string{
	"newFullRouterTestEnv(",
	"AutoMigrate(",
	"gorm.Open(",
	"newTestSQLiteDB(",
	"sql.Open(",
}

var forbiddenSystemHTTPImports = []string{
	"gorm.io/gorm",
	"gorm.io/driver/sqlite",
	"gorm.io/driver/postgres",
	"database/sql",
}

var setupDBPattern = regexp.MustCompile(`\bsetup[A-Za-z0-9_]*DB\(`)
var systemHTTPReadmeDirPattern = regexp.MustCompile("`tests/system/http/([a-z0-9]+)/?`")

func TestInternalAppSystemTestShellsStayThinAndReferenceScenarioOwners(t *testing.T) {
	t.Parallel()

	violations := make([]string, 0)
	for file, ownerMarker := range internalAppSystemShellOwnerMarkers {
		content := readBackendTestFile(t, file)
		lineCount := countLines(content)
		maxLines := defaultInternalAppSystemShellMaxLines
		if allowlistedMax, ok := allowedOversizedInternalAppSystemShells[file]; ok {
			maxLines = allowlistedMax
			if lineCount <= defaultInternalAppSystemShellMaxLines {
				violations = append(violations, file+" no longer needs oversized allowlist")
			}
		}
		if lineCount > maxLines {
			violations = append(violations, file+" has "+itoa(lineCount)+" lines; limit is "+itoa(maxLines))
		}
		if !strings.Contains(content, ownerMarker) {
			violations = append(violations, file+" must reference scenario owner "+ownerMarker)
		}
	}

	for file := range allowedOversizedInternalAppSystemShells {
		if _, ok := internalAppSystemShellOwnerMarkers[file]; !ok {
			violations = append(violations, "oversized allowlist is stale: "+file)
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("internal/app system-test shells must stay thin and point at scenario owners:\n%s", strings.Join(violations, "\n"))
	}
}

func TestSystemHTTPScenarioPackagesDoNotOwnEnvOrPersistenceSetup(t *testing.T) {
	t.Parallel()

	files := collectGoFiles(t, "tests/system/http")
	violations := make([]string, 0)
	for _, file := range files {
		content := readBackendTestFile(t, file)
		for _, snippet := range forbiddenSystemHTTPSnippets {
			if strings.Contains(content, snippet) {
				violations = append(violations, file+" must not contain "+snippet)
			}
		}
		if setupDBPattern.FindStringIndex(content) != nil {
			violations = append(violations, file+" must not own setup*DB helpers")
		}
		for _, importPath := range forbiddenSystemHTTPImports {
			if strings.Contains(content, `"`+importPath+`"`) {
				violations = append(violations, file+" must not import "+importPath)
			}
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("tests/system/http packages should stay scenario-only:\n%s", strings.Join(violations, "\n"))
	}
}

func TestBackendTestsReadmeListsCurrentSystemHTTPScenarioDirectories(t *testing.T) {
	t.Parallel()

	content := readBackendTestFile(t, "tests/README.md")
	matches := systemHTTPReadmeDirPattern.FindAllStringSubmatch(content, -1)
	documented := make(map[string]struct{}, len(matches))
	for _, match := range matches {
		documented[match[1]] = struct{}{}
	}

	actualEntries, err := os.ReadDir(filepath.Join(backendRoot, "tests", "system", "http"))
	if err != nil {
		t.Fatalf("read tests/system/http: %v", err)
	}

	actual := make(map[string]struct{}, len(actualEntries))
	for _, entry := range actualEntries {
		if entry.IsDir() {
			actual[entry.Name()] = struct{}{}
		}
	}

	violations := make([]string, 0)
	for dir := range actual {
		if _, ok := documented[dir]; !ok {
			violations = append(violations, "tests/README.md is missing tests/system/http/"+dir)
		}
	}
	for dir := range documented {
		if _, ok := actual[dir]; !ok {
			violations = append(violations, "tests/README.md references stale tests/system/http/"+dir)
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("backend tests README must track system/http scenario directories:\n%s", strings.Join(violations, "\n"))
	}
}

func collectGoFiles(t *testing.T, root string) []string {
	t.Helper()

	files := make([]string, 0)
	absRoot := filepath.Join(backendRoot, filepath.FromSlash(root))
	if err := filepath.Walk(absRoot, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if info == nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		rel, err := filepath.Rel(backendRoot, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	}); err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}

	sort.Strings(files)
	return files
}

func readBackendTestFile(t *testing.T, rel string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(backendRoot, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(content)
}

func countLines(content string) int {
	return len(strings.Split(content, "\n"))
}

func itoa(value int) string {
	return strconv.Itoa(value)
}
