package architecture

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	challengeentity "ctf-platform/internal/module/challenge/entity"
	contestentity "ctf-platform/internal/module/contest/entity"
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

var allowedAutoMigrateOwners = map[string]struct{}{
	"internal/testutil/systemapp/sqlite_helpers.go":        {},
	"internal/module/challenge/testsupport/test_helper.go": {},
	"internal/module/contest/testsupport/db.go":            {},
	"internal/module/practice/testsupport/test_helper.go":  {},
}

var setupDBPattern = regexp.MustCompile(`\bsetup[A-Za-z0-9_]*DB\(`)
var systemHTTPReadmeDirPattern = regexp.MustCompile("`tests/system/http/([a-z0-9]+)/?`")
var rawSensitiveZapFieldPattern = regexp.MustCompile(`(?i)zap\.(?:Any|ByteString|String|Stringer)\("([^"]*(?:password|token|secret)[^"]*)"\s*,`)
var rawSensitiveKeyZapStringPattern = regexp.MustCompile(`zap\.String\("key"\s*,\s*(?:key|keys\[[^\]]+\])\)`)
var rawRequestZapAnyPattern = regexp.MustCompile(`(?i)zap\.Any\("(?:req|request)"\s*,`)

func TestAWDCheckerTypeValuesStayAlignedAcrossChallengeAndContest(t *testing.T) {
	t.Parallel()

	challengeValues := []string{
		string(challengeentity.AWDCheckerTypeLegacyProbe),
		string(challengeentity.AWDCheckerTypeHTTPStandard),
		string(challengeentity.AWDCheckerTypeTCPStandard),
		string(challengeentity.AWDCheckerTypeScript),
	}
	contestValues := []string{
		string(contestentity.AWDCheckerTypeLegacyProbe),
		string(contestentity.AWDCheckerTypeHTTPStandard),
		string(contestentity.AWDCheckerTypeTCPStandard),
		string(contestentity.AWDCheckerTypeScript),
	}

	sort.Strings(challengeValues)
	sort.Strings(contestValues)
	if strings.Join(challengeValues, "\n") != strings.Join(contestValues, "\n") {
		t.Fatalf("AWD checker type values drifted between challenge and contest:\nchallenge=%v\ncontest=%v", challengeValues, contestValues)
	}
}

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

func TestAutoMigrateStaysInTestSupport(t *testing.T) {
	t.Parallel()

	files := collectGoFiles(t, ".")
	violations := make([]string, 0)
	seenAllowed := make(map[string]struct{}, len(allowedAutoMigrateOwners))
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		content := readBackendTestFile(t, file)
		if !strings.Contains(content, "AutoMigrate(") {
			continue
		}
		if _, ok := allowedAutoMigrateOwners[file]; !ok {
			violations = append(violations, file+" must not own AutoMigrate outside test support")
			continue
		}
		seenAllowed[file] = struct{}{}
	}

	for file := range allowedAutoMigrateOwners {
		if _, ok := seenAllowed[file]; !ok {
			violations = append(violations, "AutoMigrate allowlist is stale: "+file)
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("AutoMigrate must stay inside test-support helpers:\n%s", strings.Join(violations, "\n"))
	}
}

func TestNoRawSensitiveZapFields(t *testing.T) {
	t.Parallel()

	files := collectGoFiles(t, "internal")
	violations := make([]string, 0)
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		content := readBackendTestFile(t, file)
		for lineNumber, line := range strings.Split(content, "\n") {
			if rawSensitiveZapFieldPattern.FindStringIndex(line) != nil {
				violations = append(violations, file+":"+itoa(lineNumber+1)+" must not log raw password/token/secret fields")
			}
			if rawSensitiveKeyZapStringPattern.FindStringIndex(line) != nil {
				violations = append(violations, file+":"+itoa(lineNumber+1)+" must sanitize high-risk cache/session key fields")
			}
			if rawRequestZapAnyPattern.FindStringIndex(line) != nil {
				violations = append(violations, file+":"+itoa(lineNumber+1)+" must not log raw request objects")
			}
		}
	}

	if len(violations) > 0 {
		sort.Strings(violations)
		t.Fatalf("sensitive values must be sanitized before logging:\n%s", strings.Join(violations, "\n"))
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
