package app

import (
	"context"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/config"
	assessmenthttp "ctf-platform/internal/module/assessment/api/http"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	challengehttp "ctf-platform/internal/module/challenge/api/http"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contesthttp "ctf-platform/internal/module/contest/api/http"
	contestports "ctf-platform/internal/module/contest/ports"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	identitycmd "ctf-platform/internal/module/identity/application/commands"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	opshttp "ctf-platform/internal/module/ops/api/http"
	opscmd "ctf-platform/internal/module/ops/application/commands"
	opsports "ctf-platform/internal/module/ops/ports"
	practicehttp "ctf-platform/internal/module/practice/api/http"
	practiceports "ctf-platform/internal/module/practice/ports"
	practicereadmodelqueries "ctf-platform/internal/module/practice_readmodel/application/queries"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	teachingreadmodelqueries "ctf-platform/internal/module/teaching_readmodel/application/queries"
)

func TestNewRouterRegistersStudentChallengeRoutes(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	assertHasRoute(t, router, "GET", "/api/v1/challenges")
	assertHasRoute(t, router, "GET", "/api/v1/challenges/:id")
	assertHasRoute(t, router, "POST", "/api/v1/contests/:id/challenges/:cid/instances")
	assertHasRoute(t, router, "GET", "/api/v1/teacher/instances")
	assertHasRoute(t, router, "DELETE", "/api/v1/teacher/instances/:id")
	assertHasRoute(t, router, "GET", "/api/v1/users/me/progress")
	assertHasRoute(t, router, "GET", "/api/v1/users/me/timeline")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/audit-logs", "internal/module/ops")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/dashboard", "internal/module/ops")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/cheat-detection", "internal/module/ops")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/admin/notifications", "internal/module/ops")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/users/me/skill-profile", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/users/me/recommendations", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/users/:id/skill-profile", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/reports/personal", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/reports/:id", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/reports/:id/download", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/reports/class", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/students/:id/skill-profile", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/reports/class", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/awd/reviews", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/awd/reviews/:id", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/awd/reviews/:id/export/archive", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/awd/reviews/:id/export/report", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/students/:id/evidence", "internal/module/teaching_readmodel/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/challenges/:id/writeup-submissions", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/challenges/:id/writeup-submissions/me", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/challenges/:id/writeup/recommend", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "DELETE", "/api/v1/authoring/challenges/:id/writeup/recommend", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/community-writeups/:id/recommend", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "DELETE", "/api/v1/teacher/community-writeups/:id/recommend", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/community-writeups/:id/hide", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/community-writeups/:id/restore", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/writeup-submissions", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/writeup-submissions/:id", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/challenges", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/challenges/:id", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/challenges", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/challenge-imports", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/authoring/challenge-imports", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/authoring/challenge-imports/:id", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/challenge-imports/:id/commit", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/challenges/:id/self-check", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/challenges/:id/publish-requests", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/authoring/challenges/:id/publish-requests/latest", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "PUT", "/api/v1/authoring/challenges/:id/flag", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/authoring/images", "internal/module/challenge/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/admin/contests", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/admin/contests/:id/awd/checker-preview", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/contests/:id/awd/readiness", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/contests/:id/awd/rounds/:rid/traffic/summary", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/admin/contests/:id/awd/rounds/:rid/traffic/events", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/contests/:id/scoreboard", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/contests/:id/challenges/:cid/submissions", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/contests/:id/teams", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/challenges/:id/instances", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/contests/:id/challenges/:cid/instances", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/challenges/:id/submit", "internal/module/practice/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/notifications", "internal/module/ops")
	assertRouteHandlerContains(t, router, "PUT", "/api/v1/notifications/:id/read", "internal/module/ops")
	assertRouteHandlerContains(t, router, "GET", "/ws/notifications", "internal/module/ops")
	assertRouteHandlerContains(t, router, "GET", "/ws/contests/:id/announcements", "internal/module/contest/api/http")
	assertRouteHandlerContains(t, router, "GET", "/ws/contests/:id/scoreboard", "internal/module/contest/api/http")
}

func TestNewRouterUsesRuntimeHandlersForInstanceRoutes(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	assertRouteHandlerContains(t, router, "GET", "/api/v1/instances", "internal/module/runtime")
	assertRouteHandlerContains(t, router, "DELETE", "/api/v1/instances/:id", "internal/module/runtime")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/instances/:id/extend", "internal/module/runtime")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/instances/:id/access", "internal/module/runtime")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/instances/:id/proxy", "internal/module/runtime")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/instances/:id/proxy/*proxyPath", "internal/module/runtime")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/instances/:id/proxy/*proxyPath", "internal/module/runtime")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/instances", "internal/module/runtime")
	assertRouteHandlerContains(t, router, "DELETE", "/api/v1/teacher/instances/:id", "internal/module/runtime")
}

func TestBuildRoot(t *testing.T) {
	t.Parallel()

	cfg, db, cache := newAppTestDependencies(t)

	root, err := composition.BuildRoot(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("BuildRoot() error = %v", err)
	}
	if root == nil {
		t.Fatal("expected root")
	}
	if root.Events == nil {
		t.Fatal("expected events bus")
	}
}

func TestIdentityModuleContractsCompile(t *testing.T) {
	var _ identitycontracts.Authenticator = (*identitycmd.AuthenticatorService)(nil)
}

func TestOpsModuleContractsCompile(t *testing.T) {
	var _ auditlog.Recorder = (*opscmd.AuditService)(nil)
}

func TestTeachingReadmodelModuleContractsCompile(t *testing.T) {
	var _ teachingreadmodelqueries.Service = (*teachingreadmodelqueries.QueryService)(nil)
}

func TestPracticeReadmodelModuleContractsCompile(t *testing.T) {
	var _ practicereadmodelqueries.Service = (*practicereadmodelqueries.QueryService)(nil)
}

func TestCompositionModulesExposeContracts(t *testing.T) {
	t.Parallel()

	assertFieldType(t, reflect.TypeOf(composition.IdentityModule{}), "AdminHandler", reflect.TypeOf(&identityhttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.IdentityModule{}), "ProfileCommands", reflect.TypeOf((*identitycontracts.ProfileCommandService)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.IdentityModule{}), "ProfileQueries", reflect.TypeOf((*identitycontracts.ProfileQueryService)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.IdentityModule{}), "TokenService", reflect.TypeOf((*identitycontracts.Authenticator)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.IdentityModule{}), "Users", reflect.TypeOf((*identitycontracts.UserRepository)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.PracticeReadmodelModule{}), "Query", reflect.TypeOf((*practicereadmodelqueries.Service)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.RuntimeModule{}), "Handler", reflect.TypeOf(&runtimehttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.RuntimeModule{}), "PracticeInstanceRepository", reflect.TypeOf((*practiceports.InstanceRepository)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.RuntimeModule{}), "PracticeRuntimeService", reflect.TypeOf((*practiceports.RuntimeInstanceService)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.RuntimeModule{}), "ChallengeImageRuntime", reflect.TypeOf((*challengeports.ImageRuntime)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.RuntimeModule{}), "ChallengeRuntimeProbe", reflect.TypeOf((*challengeports.ChallengeRuntimeProbe)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.RuntimeModule{}), "OpsRuntimeQuery", reflect.TypeOf((*opsports.RuntimeQuery)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.RuntimeModule{}), "OpsRuntimeStatsProvider", reflect.TypeOf((*opsports.RuntimeStatsProvider)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.RuntimeModule{}), "ContestContainerFiles", reflect.TypeOf((*contestports.AWDContainerFileWriter)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "AuditService", reflect.TypeOf((*auditlog.Recorder)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "AuditHandler", reflect.TypeOf(&opshttp.AuditHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "DashboardHandler", reflect.TypeOf(&opshttp.DashboardHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "NotificationHandler", reflect.TypeOf(&opshttp.NotificationHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "RiskHandler", reflect.TypeOf(&opshttp.RiskHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.TeachingReadmodelModule{}), "Query", reflect.TypeOf((*teachingreadmodelqueries.Service)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "Catalog", reflect.TypeOf((*challengecontracts.ChallengeContract)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "FlagValidator", reflect.TypeOf((*challengecontracts.FlagValidator)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "FlagHandler", reflect.TypeOf(&challengehttp.FlagHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "Handler", reflect.TypeOf(&challengehttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "ImageHandler", reflect.TypeOf(&challengehttp.ImageHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "ImageStore", reflect.TypeOf((*challengecontracts.ImageStore)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "TopologyHandler", reflect.TypeOf(&challengehttp.TopologyHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "WriteupHandler", reflect.TypeOf(&challengehttp.WriteupHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "Handler", reflect.TypeOf(&assessmenthttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "ProfileService", reflect.TypeOf((*assessmentcontracts.ProfileService)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "Recommendations", reflect.TypeOf((*assessmentcontracts.RecommendationProvider)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "ReportHandler", reflect.TypeOf(&assessmenthttp.ReportHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "TeacherAWDReviewHandler", reflect.TypeOf(&assessmenthttp.TeacherAWDReviewHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContestModule{}), "AWDHandler", reflect.TypeOf(&contesthttp.AWDHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContestModule{}), "ChallengeHandler", reflect.TypeOf(&contesthttp.ChallengeHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContestModule{}), "Handler", reflect.TypeOf(&contesthttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContestModule{}), "ParticipationHandler", reflect.TypeOf(&contesthttp.ParticipationHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContestModule{}), "SubmissionHandler", reflect.TypeOf(&contesthttp.SubmissionHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.ContestModule{}), "TeamHandler", reflect.TypeOf(&contesthttp.TeamHandler{}))
	assertFieldType(t, reflect.TypeOf(composition.PracticeModule{}), "Handler", reflect.TypeOf(&practicehttp.Handler{}))
	assertNoField(t, reflect.TypeOf(composition.AuthModule{}), "TokenService")
	assertNoField(t, reflect.TypeOf(composition.ChallengeModule{}), "FlagService")
	assertNoField(t, reflect.TypeOf(composition.ChallengeModule{}), "ImageRepository")
	assertNoField(t, reflect.TypeOf(composition.ChallengeModule{}), "ImageService")
	assertNoField(t, reflect.TypeOf(composition.ChallengeModule{}), "Repository")
	assertNoField(t, reflect.TypeOf(composition.ContestModule{}), "Repository")
	assertNoField(t, reflect.TypeOf(composition.AssessmentModule{}), "RecommendationService")
	assertNoField(t, reflect.TypeOf(composition.AssessmentModule{}), "ReportService")
	assertNoField(t, reflect.TypeOf(composition.AssessmentModule{}), "Service")
	assertNoField(t, reflect.TypeOf(composition.PracticeModule{}), "Service")
	assertNoField(t, reflect.TypeOf(composition.RuntimeModule{}), "Query")
	assertNoField(t, reflect.TypeOf(composition.RuntimeModule{}), "Repository")
	assertNoField(t, reflect.TypeOf(composition.RuntimeModule{}), "Service")
	assertNoField(t, reflect.TypeOf(composition.IdentityModule{}), "users")
}

func TestCompositionBuildersUseRuntimeModuleForRuntimeDependencies(t *testing.T) {
	t.Parallel()

	assertFunctionParamType(t, reflect.TypeOf(composition.BuildChallengeModule), 1, reflect.TypeOf(&composition.RuntimeModule{}))
	assertFunctionParamType(t, reflect.TypeOf(composition.BuildChallengeModule), 2, reflect.TypeOf(&composition.OpsModule{}))
	assertFunctionParamType(t, reflect.TypeOf(composition.BuildContestModule), 2, reflect.TypeOf(&composition.RuntimeModule{}))
	assertFunctionParamType(t, reflect.TypeOf(composition.BuildOpsModule), 1, reflect.TypeOf(&composition.RuntimeModule{}))
	assertFunctionParamType(t, reflect.TypeOf(composition.BuildPracticeModule), 2, reflect.TypeOf(&composition.RuntimeModule{}))
}

func TestBuildOpsModuleDelegatesToSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "ops_module.go"))
	if err != nil {
		t.Fatalf("read ops_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"buildOpsModuleDeps(",
		"buildOpsAuditHandler(",
		"buildOpsDashboardHandler(",
		"buildOpsRiskHandler(",
		"buildOpsNotificationDeps(",
		"buildOpsNotificationHandler(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("ops module should delegate through %s", marker)
		}
	}
}

func TestBuildRuntimeModuleDelegatesToSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "runtime_module.go"))
	if err != nil {
		t.Fatalf("read runtime_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"buildRuntimeModuleDeps(",
		"registerRuntimeBackgroundJobs(",
		"buildRuntimeHTTPDeps(",
		"buildRuntimePracticeDeps(",
		"buildRuntimeChallengeDeps(",
		"buildRuntimeOpsDeps(",
		"buildRuntimeContestDeps(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime module should delegate through %s", marker)
		}
	}
}

func TestRouterRateLimitStrategyUsesUserAndLoginPrincipalKeys(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile("router.go")
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}

	source := string(content)
	expected := []string{
		`protected.Use(middleware.RateLimitByUser(`,
		`middleware.RateLimitByLoginPrincipalAndIP(`,
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("router should include rate limit marker %s", marker)
		}
	}

	blocked := []string{
		`engine.Use(middleware.RateLimitByIP(rateChecker, "global"`,
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("router should not keep global IP rate limit marker %s", marker)
		}
	}
}

func TestRuntimeModuleUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "runtime_module.go"))
	if err != nil {
		t.Fatalf("read runtime_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type runtimeModuleDeps struct",
		"repo",
		"runtimeports.InstanceRepository",
		"practiceInstanceRepo",
		"practiceports.InstanceRepository",
		"instanceCommands",
		"runtimeHTTPCommandService",
		"instanceQueries",
		"runtimeHTTPQueryService",
		"countRunningQuery",
		"opsports.RuntimeQuery",
		"proxyTicketService",
		"runtimeHTTPProxyTicketService",
		"cleanupService",
		"*runtimecmd.RuntimeCleanupService",
		"maintenanceService",
		"*runtimecmd.RuntimeMaintenanceService",
		"provisioningService",
		"*runtimecmd.ProvisioningService",
		"imageRuntime",
		"challengeports.ImageRuntime",
		"containerFiles",
		"contestports.AWDContainerFileWriter",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime composition should declare typed deps marker %s", marker)
		}
	}
}

func TestRuntimeModuleUsesCommandsQueriesServices(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "runtime_module.go"))
	if err != nil {
		t.Fatalf("read runtime_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"runtimecmd.NewInstanceService(",
		"runtimecmd.NewRuntimeCleanupService(",
		"runtimecmd.NewRuntimeMaintenanceService(",
		"runtimecmd.NewProvisioningService(",
		"runtimeqry.NewInstanceService(",
		"runtimeqry.NewCountRunningService(",
		"runtimeqry.NewProxyTicketService(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime composition should use layered runtime service marker %s", marker)
		}
	}

	blocked := []string{
		"runtimeapp.NewInstanceService(",
		"runtimeapp.NewQueryService(",
		"runtimeapp.NewProxyTicketService(",
		"runtimeapp.NewRuntimeCleanupService(",
		"runtimeapp.NewRuntimeMaintenanceService(",
		"runtimeapp.NewProvisioningService(",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("runtime composition should not keep legacy root service marker %s", marker)
		}
	}
}

func TestAuthModuleUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "auth_module.go"))
	if err != nil {
		t.Fatalf("read auth_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type authModuleDeps struct",
		"users",
		"identitycontracts.UserRepository",
		"tokenService",
		"authcontracts.TokenService",
		"profileCommands",
		"identitycontracts.ProfileCommandService",
		"profileQueries",
		"identitycontracts.ProfileQueryService",
		"auditRecorder",
		"auditlog.Recorder",
		"buildAuthModuleDeps(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("auth composition should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"authcmd.NewService(identity.users, identity.TokenService",
		"authcmd.NewCASService(cfg.Auth.CAS, identity.users, identity.TokenService",
		"users:           identity.users,",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("auth composition should not keep direct module dependency marker %s", marker)
		}
	}
}

func TestOpsModuleUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "ops_module.go"))
	if err != nil {
		t.Fatalf("read ops_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type opsModuleDeps struct",
		"auditRepo",
		"opsports.AuditRepository",
		"riskRepo",
		"opsports.RiskRepository",
		"runtimeQuery",
		"opsports.RuntimeQuery",
		"runtimeStats",
		"opsports.RuntimeStatsProvider",
		"type opsNotificationDeps struct",
		"notificationRepo",
		"opsports.NotificationRepository",
		"broadcaster",
		"opsports.NotificationBroadcaster",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("ops composition should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"auditRepo := opsinfra.NewAuditRepository(db)",
		"riskRepo := opsinfra.NewRiskRepository(db)",
		"notificationRepo := opsinfra.NewNotificationRepository(db)",
		"runtimeapp \"ctf-platform/internal/module/runtime/application\"",
		"runtime.ops.query",
		"runtime.ops.statsProvider",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("ops composition should not keep concrete marker %s", marker)
		}
	}
}

func TestBuildContestModuleDelegatesToSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "contest_module.go"))
	if err != nil {
		t.Fatalf("read contest_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"buildContestCoreHandler(",
		"buildContestAWDHandler(",
		"buildContestChallengeHandler(",
		"buildContestParticipationHandler(",
		"buildContestTeamHandler(",
		"buildContestSubmissionHandler(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("contest module should delegate through %s", marker)
		}
	}
}

func TestContestModuleDepsAvoidConcreteContestRepositories(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "contest_module.go"))
	if err != nil {
		t.Fatalf("read contest_module.go: %v", err)
	}

	source := string(content)
	blocked := []string{
		"*contestinfra.Repository",
		"*contestinfra.AWDRepository",
		"*contestinfra.ChallengeRepository",
		"*contestinfra.TeamRepository",
		"*contestinfra.ParticipationRepository",
		"*contestinfra.SubmissionRepository",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("contest composition deps should not keep concrete repo field %s", marker)
		}
	}
}

func TestContestModuleUsesTypedCrossModuleDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "contest_module.go"))
	if err != nil {
		t.Fatalf("read contest_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type contestModuleDeps struct",
		"challengeCatalog",
		"challengecontracts.ContestChallengeContract",
		"flagValidator",
		"challengecontracts.FlagValidator",
		"containerFiles",
		"contestports.AWDContainerFileWriter",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("contest composition should declare typed cross-module deps marker %s", marker)
		}
	}

	blocked := []string{
		"challenge         *ChallengeModule",
		"runtime           *RuntimeModule",
		"deps.runtime.contest.containerFiles",
		"runtime.contest.containerFiles",
		"deps.challenge.Catalog",
		"deps.challenge.FlagValidator",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("contest composition should not keep direct module dependency marker %s", marker)
		}
	}
}

func TestChallengeModuleUsesTypedPortsDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "challenge_module.go"))
	if err != nil {
		t.Fatalf("read challenge_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type challengeModuleDeps struct",
		"challengeCommandRepo",
		"challengeports.ChallengeCommandRepository",
		"challengeQueryRepo",
		"challengeports.ChallengeQueryRepository",
		"flagRepo",
		"challengeports.ChallengeFlagRepository",
		"imageUsageRepo",
		"challengeports.ChallengeImageUsageRepository",
		"topologyRepo",
		"challengeports.ChallengeTopologyRepository",
		"writeupRepo",
		"challengeports.ChallengeWriteupRepository",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("challenge composition should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"challengeRepo *challengeinfra.Repository",
		"imageRepo *challengeinfra.ImageRepository",
		"templateRepo *challengeinfra.TemplateRepository",
		"runtime.challenge.imageRuntime",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("challenge composition deps should not keep concrete repository field %s", marker)
		}
	}
}

func TestBuildChallengeModuleDelegatesToSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "challenge_module.go"))
	if err != nil {
		t.Fatalf("read challenge_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"buildChallengeImageHandler(",
		"buildChallengeCoreHandler(",
		"buildChallengeFlagHandler(",
		"buildChallengeTopologyHandler(",
		"buildChallengeWriteupHandler(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("challenge module should delegate through %s", marker)
		}
	}
}

func TestPracticeModuleUsesTypedPortsDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "practice_module.go"))
	if err != nil {
		t.Fatalf("read practice_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type practiceModuleDeps struct",
		"commandRepo",
		"practiceports.PracticeCommandRepository",
		"scoreRepo",
		"practiceports.PracticeScoreRepository",
		"rankingRepo",
		"practiceports.PracticeRankingRepository",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("practice composition should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"repo := practiceinfra.NewRepository(db)",
		"*practiceinfra.Repository",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("practice composition deps should not keep concrete practice repository marker %s", marker)
		}
	}
}

func TestPracticeModuleUsesTypedCrossModuleDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "practice_module.go"))
	if err != nil {
		t.Fatalf("read practice_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type practiceModuleExternalDeps struct",
		"instanceRepo",
		"practiceports.InstanceRepository",
		"runtimeService",
		"practiceports.RuntimeInstanceService",
		"runtime.PracticeInstanceRepository",
		"runtime.PracticeRuntimeService",
		"challengeRepo",
		"practiceRuntimeChallengeContract",
		"imageStore",
		"practiceRuntimeImageStore",
		"assessment",
		"practiceRuntimeAssessmentService",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("practice composition should declare typed cross-module deps marker %s", marker)
		}
	}

	blocked := []string{
		"runtime.practice.instanceRepository",
		"runtime.practice.runtimeService",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("practice composition should not keep runtime private marker %s", marker)
		}
	}
}

func TestBuildPracticeModuleDelegatesToSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "practice_module.go"))
	if err != nil {
		t.Fatalf("read practice_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"buildPracticeModuleDeps(",
		"buildPracticeModuleExternalDeps(",
		"buildPracticeHandler(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("practice module should delegate through %s", marker)
		}
	}
}

func TestAssessmentModuleUsesTypedPortsDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "assessment_module.go"))
	if err != nil {
		t.Fatalf("read assessment_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type assessmentModuleDeps struct",
		"profileRepo",
		"assessmentports.ProfileRepository",
		"recommendationRepo",
		"assessmentports.RecommendationRepository",
		"challengeRepo",
		"assessmentports.ChallengeRepository",
		"reportRepo",
		"assessmentports.ReportRepository",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("assessment composition should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"repo := assessmentinfra.NewRepository(db)",
		"reportRepo := assessmentinfra.NewReportRepository(db)",
		"*assessmentinfra.Repository",
		"*assessmentinfra.ReportRepository",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("assessment composition deps should not keep concrete assessment repository marker %s", marker)
		}
	}
}

func TestAssessmentModuleUsesTypedCrossModuleDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "assessment_module.go"))
	if err != nil {
		t.Fatalf("read assessment_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type assessmentModuleExternalDeps struct",
		"challengeRepo",
		"assessmentports.ChallengeRepository",
		"buildAssessmentModuleExternalDeps(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("assessment composition should declare typed cross-module deps marker %s", marker)
		}
	}
}

func TestBuildAssessmentModuleDelegatesToSubBuilders(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "assessment_module.go"))
	if err != nil {
		t.Fatalf("read assessment_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"buildAssessmentModuleDeps(",
		"buildAssessmentModuleExternalDeps(",
		"buildAssessmentProfileHandler(",
		"buildAssessmentRecommendationHandler(",
		"buildAssessmentReportHandler(",
		"buildAssessmentTeacherAWDReviewHandler(",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("assessment module should delegate through %s", marker)
		}
	}
}

func TestPracticeModuleAvoidsRuntimeBridgeGlue(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "practice_module.go"))
	if err != nil {
		t.Fatalf("read practice_module.go: %v", err)
	}

	source := string(content)
	blocked := []string{
		"type practiceRuntimeCleanerBridge interface",
		"type practiceRuntimeRepositoryBridge interface",
		"type practiceRuntimeInstanceService interface",
		"type practiceRuntimeProvisioningBridge interface",
		"type practiceRuntimeInstanceServiceAdapter struct",
		"newPracticeRuntimeInstanceServiceAdapter(",
		"toRuntimeTopologyCreateRequest(",
		"fromRuntimeTopologyCreateResult(",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("practice composition should not keep runtime bridge marker %s", marker)
		}
	}
}

func TestRuntimeModuleUsesExternalPortsForCrossModuleDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "runtime_module.go"))
	if err != nil {
		t.Fatalf("read runtime_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"practiceports.InstanceRepository",
		"practiceports.RuntimeInstanceService",
		"contestports.AWDContainerFileWriter",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("runtime composition should use external ports marker %s", marker)
		}
	}

	blocked := []string{
		"contestinfra.AWDContainerFileWriter",
		"practiceRuntimeRepositoryBridge",
		"practiceRuntimeInstanceService",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("runtime composition should not keep bridge marker %s", marker)
		}
	}
}

func TestCompositionBuildersAvoidPrivateCrossModuleFields(t *testing.T) {
	t.Parallel()

	files, err := filepath.Glob(filepath.Join("composition", "*_module.go"))
	if err != nil {
		t.Fatalf("glob composition modules: %v", err)
	}

	blocked := []string{
		"identity.users",
		"runtime.practice.",
		"runtime.ops.",
		"runtime.challenge.",
		"runtime.contest.",
	}

	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}

		source := string(content)
		for _, marker := range blocked {
			if strings.Contains(source, marker) {
				t.Fatalf("%s should not reference private cross-module field %s", file, marker)
			}
		}
	}
}

func TestIdentityModuleUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "identity_module.go"))
	if err != nil {
		t.Fatalf("read identity_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type identityModuleDeps struct",
		"users",
		"identitycontracts.UserRepository",
		"tokenService",
		"identitycontracts.Authenticator",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("identity composition should declare typed deps marker %s", marker)
		}
	}

	blocked := []string{
		"users := identityinfra.NewRepository(db)",
		"identityinfra.NewRepository(db)",
	}
	for _, marker := range blocked {
		if strings.Contains(source, marker) {
			t.Fatalf("identity composition should not keep concrete marker %s", marker)
		}
	}
}

func TestPracticeReadmodelModuleUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "practice_readmodel_module.go"))
	if err != nil {
		t.Fatalf("read practice_readmodel_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type practiceReadmodelModuleDeps struct",
		"repo",
		"practicereadmodelports.QueryRepository",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("practice readmodel composition should declare typed deps marker %s", marker)
		}
	}

	if strings.Contains(source, "practicereadmodelinfra.NewRepository(db)") {
		t.Fatalf("practice readmodel composition should not instantiate concrete repo inline")
	}
}

func TestTeachingReadmodelModuleUsesTypedDeps(t *testing.T) {
	t.Parallel()

	content, err := os.ReadFile(filepath.Join("composition", "teaching_readmodel_module.go"))
	if err != nil {
		t.Fatalf("read teaching_readmodel_module.go: %v", err)
	}

	source := string(content)
	expected := []string{
		"type teachingReadmodelModuleDeps struct",
		"repo",
		"readmodelports.Repository",
	}
	for _, marker := range expected {
		if !strings.Contains(source, marker) {
			t.Fatalf("teaching readmodel composition should declare typed deps marker %s", marker)
		}
	}

	if strings.Contains(source, "readmodelinfra.NewRepository(db)") {
		t.Fatalf("teaching readmodel composition should not instantiate concrete repo inline")
	}
}

func assertHasRoute(t *testing.T, router *gin.Engine, method, path string) {
	t.Helper()

	for _, route := range router.Routes() {
		if route.Method == method && route.Path == path {
			return
		}
	}

	t.Fatalf("route not found: %s %s", method, path)
}

func assertRouteHandlerContains(t *testing.T, router *gin.Engine, method, path, want string) {
	t.Helper()

	for _, route := range router.Routes() {
		if route.Method == method && route.Path == path {
			if !strings.Contains(route.Handler, want) {
				t.Fatalf("route handler for %s %s = %s, want substring %s", method, path, route.Handler, want)
			}
			return
		}
	}

	t.Fatalf("route not found: %s %s", method, path)
}

func assertFieldType(t *testing.T, structType reflect.Type, fieldName string, want reflect.Type) {
	t.Helper()

	field, ok := structType.FieldByName(fieldName)
	if !ok {
		t.Fatalf("%s missing field %s", structType.Name(), fieldName)
	}
	if field.Type != want {
		t.Fatalf("%s.%s type = %s, want %s", structType.Name(), fieldName, field.Type, want)
	}
}

func assertNoField(t *testing.T, structType reflect.Type, fieldName string) {
	t.Helper()

	if _, ok := structType.FieldByName(fieldName); ok {
		t.Fatalf("%s unexpectedly exposes field %s", structType.Name(), fieldName)
	}
}

func assertFunctionParamType(t *testing.T, fnType reflect.Type, index int, want reflect.Type) {
	t.Helper()

	if fnType.Kind() != reflect.Func {
		t.Fatalf("expected function type, got %s", fnType.Kind())
	}
	if index < 0 || index >= fnType.NumIn() {
		t.Fatalf("function has %d params, index %d out of range", fnType.NumIn(), index)
	}
	if got := fnType.In(index); got != want {
		t.Fatalf("function param %d type = %s, want %s", index, got, want)
	}
}

func newAppTestDependencies(t *testing.T) (*config.Config, *gorm.DB, *redislib.Client) {
	t.Helper()

	mini, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	t.Cleanup(mini.Close)

	cache := redislib.NewClient(&redislib.Options{Addr: mini.Addr()})
	t.Cleanup(func() {
		_ = cache.Close()
	})
	if err := cache.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("ping redis: %v", err)
	}

	dbPath := filepath.Join(t.TempDir(), "router.sqlite")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	return newPracticeFlowTestConfig(t), db, cache
}
