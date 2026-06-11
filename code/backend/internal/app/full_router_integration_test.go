package app

import (
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"ctf-platform/internal/app/composition"
	authcontracts "ctf-platform/internal/module/auth/contracts"
	authruntime "ctf-platform/internal/module/auth/runtime"
)

func TestRouterBuildUsesCompositionModules(t *testing.T) {
	cfg, db, cache := newAppTestDependencies(t)

	var calls []string

	originalBuildContainerRuntimeModule := buildContainerRuntimeModule
	originalBuildInstanceModule := buildInstanceModule
	originalBuildOpsModule := buildOpsModule
	originalBuildIdentityModule := buildIdentityModule
	originalBuildAuthModule := buildAuthModule
	originalBuildChallengeModule := buildChallengeModule
	originalBuildAssessmentModule := buildAssessmentModule
	originalBuildTeachingAnalysisModule := buildTeachingAnalysisModule
	originalBuildContestModule := buildContestModule
	originalBuildPracticeModule := buildPracticeModule
	defer func() {
		buildContainerRuntimeModule = originalBuildContainerRuntimeModule
		buildInstanceModule = originalBuildInstanceModule
		buildOpsModule = originalBuildOpsModule
		buildIdentityModule = originalBuildIdentityModule
		buildAuthModule = originalBuildAuthModule
		buildChallengeModule = originalBuildChallengeModule
		buildAssessmentModule = originalBuildAssessmentModule
		buildTeachingAnalysisModule = originalBuildTeachingAnalysisModule
		buildContestModule = originalBuildContestModule
		buildPracticeModule = originalBuildPracticeModule
	}()

	buildContainerRuntimeModule = func(root *composition.Root) (*composition.ContainerRuntimeModule, error) {
		if root == nil {
			t.Fatal("expected root for container runtime module builder")
		}
		calls = append(calls, "container_runtime")
		return originalBuildContainerRuntimeModule(root)
	}
	buildInstanceModule = func(root *composition.Root, runtime *composition.ContainerRuntimeModule) *composition.InstanceModule {
		if root == nil || runtime == nil {
			t.Fatal("expected root and container runtime for instance module builder")
		}
		calls = append(calls, "instance")
		return originalBuildInstanceModule(root, runtime)
	}
	buildOpsModule = func(root *composition.Root, runtime *composition.ContainerRuntimeModule) *composition.OpsModule {
		if root == nil || runtime == nil {
			t.Fatal("expected root and container runtime for ops module builder")
		}
		calls = append(calls, "ops")
		return originalBuildOpsModule(root, runtime)
	}
	buildIdentityModule = func(root *composition.Root) (*composition.IdentityModule, error) {
		if root == nil {
			t.Fatal("expected root for identity module builder")
		}
		calls = append(calls, "identity")
		return originalBuildIdentityModule(root)
	}
	buildAuthModule = func(root *composition.Root, ops *composition.OpsModule, identity *composition.IdentityModule, tokenService authcontracts.TokenService) (*authruntime.Module, error) {
		if root == nil || ops == nil || identity == nil || tokenService == nil {
			t.Fatal("expected root, ops, identity and token service for auth module builder")
		}
		calls = append(calls, "auth")
		return originalBuildAuthModule(root, ops, identity, tokenService)
	}
	buildChallengeModule = func(root *composition.Root, runtime *composition.ContainerRuntimeModule) (*composition.ChallengeModule, error) {
		if root == nil || runtime == nil {
			t.Fatal("expected root and container runtime for challenge module builder")
		}
		calls = append(calls, "challenge")
		return originalBuildChallengeModule(root, runtime)
	}
	buildAssessmentModule = func(root *composition.Root, challenge *composition.ChallengeModule) *composition.AssessmentModule {
		if root == nil || challenge == nil {
			t.Fatal("expected root and challenge for assessment module builder")
		}
		calls = append(calls, "assessment")
		return originalBuildAssessmentModule(root, challenge)
	}
	buildTeachingAnalysisModule = func(root *composition.Root, assessment *composition.AssessmentModule, identity *composition.IdentityModule) *composition.TeachingAnalysisModule {
		if root == nil || assessment == nil || identity == nil {
			t.Fatal("expected root, assessment and identity for teaching analysis module builder")
		}
		calls = append(calls, "teaching_analysis")
		return originalBuildTeachingAnalysisModule(root, assessment, identity)
	}
	buildContestModule = func(root *composition.Root, challenge *composition.ChallengeModule, runtime *composition.ContainerRuntimeModule) *composition.ContestModule {
		if root == nil || challenge == nil || runtime == nil {
			t.Fatal("expected root, challenge and container runtime for contest module builder")
		}
		calls = append(calls, "contest")
		return originalBuildContestModule(root, challenge, runtime)
	}
	buildPracticeModule = func(root *composition.Root, challenge *composition.ChallengeModule, instance *composition.InstanceModule) *composition.PracticeModule {
		if root == nil || challenge == nil || instance == nil {
			t.Fatal("expected root, challenge and instance for practice module builder")
		}
		calls = append(calls, "practice")
		return originalBuildPracticeModule(root, challenge, instance)
	}

	router, err := NewRouter(cfg, zap.NewNop(), db, cache)
	if err != nil {
		t.Fatalf("NewRouter() error = %v", err)
	}
	if router == nil {
		t.Fatal("expected router")
	}

	expectedCalls := []string{"container_runtime", "ops", "instance", "identity", "auth", "challenge", "assessment", "teaching_analysis", "contest", "practice"}
	if len(calls) != len(expectedCalls) {
		t.Fatalf("expected %d module builder calls, got %d (%v)", len(expectedCalls), len(calls), calls)
	}
	for i, expected := range expectedCalls {
		if calls[i] != expected {
			t.Fatalf("expected builder call %d to be %q, got %q (%v)", i, expected, calls[i], calls)
		}
	}
}

func isAcceptableSmokeStatus(method, path string, status int) bool {
	if status < http.StatusInternalServerError {
		return true
	}
	if method == http.MethodGet && (path == "/api/v1/auth/cas/login" || path == "/api/v1/auth/cas/callback") && status == http.StatusServiceUnavailable {
		return true
	}
	return false
}

type routeAccessLevel string

const (
	routeAccessPublic    routeAccessLevel = "public"
	routeAccessProtected routeAccessLevel = "protected"
	routeAccessTeacher   routeAccessLevel = "teacher"
	routeAccessAdmin     routeAccessLevel = "admin"
)

func filteredRouterRoutes(routes gin.RoutesInfo) gin.RoutesInfo {
	filtered := make(gin.RoutesInfo, 0, len(routes))
	for _, route := range routes {
		if strings.HasPrefix(route.Path, "/ws/") {
			continue
		}
		if route.Path == "/favicon.ico" {
			continue
		}
		filtered = append(filtered, route)
	}
	return filtered
}

func classifyRouteAccess(method, path string) routeAccessLevel {
	if isPublicRoute(method, path) {
		return routeAccessPublic
	}
	if strings.HasPrefix(path, "/api/v1/admin") {
		if isTeacherAuthoringAdminRoute(path) {
			return routeAccessTeacher
		}
		return routeAccessAdmin
	}
	if strings.HasPrefix(path, "/api/v1/teacher") {
		return routeAccessTeacher
	}
	if path == "/api/v1/users/:id/skill-profile" || path == "/api/v1/reports/class" {
		return routeAccessTeacher
	}
	return routeAccessProtected
}

func isTeacherAuthoringAdminRoute(path string) bool {
	return strings.HasPrefix(path, "/api/v1/authoring/challenges") ||
		strings.HasPrefix(path, "/api/v1/authoring/images") ||
		strings.HasPrefix(path, "/api/v1/authoring/environment-templates")
}

func isPublicRoute(method, path string) bool {
	switch path {
	case "/live", "/ready", "/health", "/health/db", "/health/redis",
		"/api/v1/live", "/api/v1/ready", "/api/v1/health", "/api/v1/health/db", "/api/v1/health/redis",
		"/api/v1/auth/register", "/api/v1/auth/login",
		"/api/v1/auth/cas/status", "/api/v1/auth/cas/login", "/api/v1/auth/cas/callback",
		"/ws/notifications",
		"/ws/contests/:id/announcements", "/ws/contests/:id/scoreboard",
		"/api/v1/contests", "/api/v1/contests/:id", "/api/v1/contests/:id/scoreboard", "/api/v1/contests/:id/announcements",
		"/api/v1/instances/:id/proxy", "/api/v1/instances/:id/proxy/*proxyPath",
		"/api/v1/contests/:id/awd/services/:sid/targets/:team_id/proxy",
		"/api/v1/contests/:id/awd/services/:sid/targets/:team_id/proxy/*proxyPath":
		return true
	}
	return false
}

func authorizedHeadersForRoute(t *testing.T, env *fullRouterTestEnv, method, path string) map[string]string {
	t.Helper()

	switch classifyRouteAccess(method, path) {
	case routeAccessAdmin:
		return sessionHeaders(loginForSession(t, env.router, env.admin.Username, env.adminPwd))
	case routeAccessTeacher:
		return sessionHeaders(loginForSession(t, env.router, env.teacher.Username, env.teacherPwd))
	case routeAccessProtected:
		return sessionHeaders(loginForSession(t, env.router, env.student.Username, env.studentPwd))
	default:
		return nil
	}
}

func routePayload(method, path string) any {
	if method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch {
		if strings.HasPrefix(path, "/api/v1/auth/register") {
			return map[string]any{
				"username": "matrix_user",
				"password": "Password123",
			}
		}
		if strings.HasPrefix(path, "/api/v1/auth/login") {
			return map[string]any{
				"username": "matrix_user",
				"password": "Password123",
			}
		}
		return map[string]any{}
	}
	return nil
}

func materializeRoutePath(path string, env *fullRouterTestEnv) string {
	target := path

	switch {
	case strings.Contains(path, "/api/v1/authoring/images/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.image.ID, 10))
	case strings.Contains(path, "/api/v1/authoring/challenges/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.challenge.ID, 10))
	case strings.Contains(path, "/api/v1/authoring/environment-templates/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.template.ID, 10))
	case strings.Contains(path, "/api/v1/admin/users/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.student.ID, 10))
	case strings.Contains(path, "/api/v1/admin/contests/:id/awd/rounds/:rid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.awdContest.ID, 10))
		target = strings.ReplaceAll(target, ":rid", strconv.FormatInt(env.awdRound.ID, 10))
	case strings.Contains(path, "/api/v1/admin/contests/:id/registrations/:rid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
		target = strings.ReplaceAll(target, ":rid", strconv.FormatInt(env.registration.ID, 10))
	case strings.Contains(path, "/api/v1/admin/contests/:id/announcements/:aid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
		target = strings.ReplaceAll(target, ":aid", strconv.FormatInt(env.announcement.ID, 10))
	case strings.Contains(path, "/api/v1/admin/contests/:id/challenges/:cid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
		target = strings.ReplaceAll(target, ":cid", strconv.FormatInt(env.challenge.ID, 10))
	case strings.Contains(path, "/api/v1/admin/contests/:id/scoreboard/live"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.awdContest.ID, 10))
	case strings.Contains(path, "/api/v1/admin/contests/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
	case strings.Contains(path, "/api/v1/teacher/instances/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.instance.ID, 10))
	case strings.Contains(path, "/api/v1/teacher/students/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.student.ID, 10))
	case strings.Contains(path, "/api/v1/teacher/classes/:name"):
		target = strings.ReplaceAll(target, ":name", env.className)
	case strings.Contains(path, "/api/v1/notifications/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.notification.ID, 10))
	case strings.Contains(path, "/api/v1/reports/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.report.ID, 10))
	case strings.Contains(path, "/api/v1/users/:id/skill-profile"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.student.ID, 10))
	case strings.Contains(path, "/api/v1/contests/:id/awd/challenges/:cid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.awdContest.ID, 10))
		target = strings.ReplaceAll(target, ":cid", strconv.FormatInt(env.challenge.ID, 10))
	case strings.Contains(path, "/api/v1/contests/:id/teams/:tid/members/:uid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
		target = strings.ReplaceAll(target, ":tid", strconv.FormatInt(env.team.ID, 10))
		target = strings.ReplaceAll(target, ":uid", strconv.FormatInt(env.student.ID, 10))
	case strings.Contains(path, "/api/v1/contests/:id/teams/:tid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
		target = strings.ReplaceAll(target, ":tid", strconv.FormatInt(env.team.ID, 10))
	case strings.Contains(path, "/api/v1/contests/:id/challenges/:cid"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
		target = strings.ReplaceAll(target, ":cid", strconv.FormatInt(env.challenge.ID, 10))
	case strings.Contains(path, "/api/v1/contests/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.contest.ID, 10))
	case strings.Contains(path, "/api/v1/challenges/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.challenge.ID, 10))
	case strings.Contains(path, "/api/v1/instances/:id"):
		target = strings.ReplaceAll(target, ":id", strconv.FormatInt(env.instance.ID, 10))
	}

	target = strings.ReplaceAll(target, ":level", "1")
	target = strings.ReplaceAll(target, "*proxyPath", "sample")
	return target
}
