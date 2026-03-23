package app

import (
	"context"
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
	"ctf-platform/internal/config"
	assessmenthttp "ctf-platform/internal/module/assessment/api/http"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	"ctf-platform/internal/module/identity"
	identityhttp "ctf-platform/internal/module/identity/api/http"
	"ctf-platform/internal/module/ops"
	practicereadmodel "ctf-platform/internal/module/practice_readmodel"
	practicereadmodelapp "ctf-platform/internal/module/practice_readmodel/application"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	teachingreadmodel "ctf-platform/internal/module/teaching_readmodel"
	teachingreadmodelapp "ctf-platform/internal/module/teaching_readmodel/application"
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
	assertRouteHandlerContains(t, router, "GET", "/api/v1/users/me/skill-profile", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/users/me/recommendations", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/users/:id/skill-profile", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/reports/personal", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/reports/:id", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/reports/:id/download", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/reports/class", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/teacher/students/:id/skill-profile", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "POST", "/api/v1/teacher/reports/class", "internal/module/assessment/api/http")
	assertRouteHandlerContains(t, router, "GET", "/api/v1/notifications", "internal/module/ops")
	assertRouteHandlerContains(t, router, "PUT", "/api/v1/notifications/:id/read", "internal/module/ops")
	assertRouteHandlerContains(t, router, "GET", "/ws/notifications", "internal/module/ops")
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
	var _ identity.Authenticator = (*identity.Module)(nil)
}

func TestOpsModuleContractsCompile(t *testing.T) {
	var _ ops.AuditRecorder = (*ops.Module)(nil)
}

func TestTeachingReadmodelModuleContractsCompile(t *testing.T) {
	var _ teachingreadmodel.TeachingQuery = (*teachingreadmodelapp.QueryService)(nil)
}

func TestPracticeReadmodelModuleContractsCompile(t *testing.T) {
	var _ practicereadmodel.PracticeQuery = (*practicereadmodelapp.QueryService)(nil)
}

func TestCompositionModulesExposeContracts(t *testing.T) {
	t.Parallel()

	assertFieldType(t, reflect.TypeOf(composition.IdentityModule{}), "AdminHandler", reflect.TypeOf(&identityhttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.IdentityModule{}), "ProfileService", reflect.TypeOf((*identity.ProfileService)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.IdentityModule{}), "TokenService", reflect.TypeOf((*identity.Authenticator)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.PracticeReadmodelModule{}), "Query", reflect.TypeOf((*practicereadmodel.PracticeQuery)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.RuntimeModule{}), "Handler", reflect.TypeOf(&runtimehttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "AuditService", reflect.TypeOf((*ops.AuditRecorder)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "AuditHandler", reflect.TypeOf((*ops.AuditLogHandler)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "DashboardHandler", reflect.TypeOf((*ops.DashboardHandler)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "NotificationHandler", reflect.TypeOf((*ops.NotificationHandler)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.OpsModule{}), "RiskHandler", reflect.TypeOf((*ops.RiskHandler)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.TeachingReadmodelModule{}), "Query", reflect.TypeOf((*teachingreadmodel.TeachingQuery)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "Catalog", reflect.TypeOf((*challengecontracts.ChallengeContract)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "FlagValidator", reflect.TypeOf((*challengecontracts.FlagValidator)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ChallengeModule{}), "ImageStore", reflect.TypeOf((*challengecontracts.ImageStore)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "Handler", reflect.TypeOf(&assessmenthttp.Handler{}))
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "ProfileService", reflect.TypeOf((*assessmentcontracts.ProfileService)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "Recommendations", reflect.TypeOf((*assessmentcontracts.RecommendationProvider)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.AssessmentModule{}), "ReportHandler", reflect.TypeOf(&assessmenthttp.ReportHandler{}))
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
}

func TestCompositionBuildersUseRuntimeModuleForRuntimeDependencies(t *testing.T) {
	t.Parallel()

	assertFunctionParamType(t, reflect.TypeOf(composition.BuildChallengeModule), 1, reflect.TypeOf(&composition.RuntimeModule{}))
	assertFunctionParamType(t, reflect.TypeOf(composition.BuildContestModule), 2, reflect.TypeOf(&composition.RuntimeModule{}))
	assertFunctionParamType(t, reflect.TypeOf(composition.BuildPracticeModule), 2, reflect.TypeOf(&composition.RuntimeModule{}))
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
