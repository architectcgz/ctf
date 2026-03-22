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
	containerModule "ctf-platform/internal/module/container"
	"ctf-platform/internal/module/identity"
	"ctf-platform/internal/module/ops"
	practicereadmodel "ctf-platform/internal/module/practice_readmodel"
	"ctf-platform/internal/module/runtime"
	teachingreadmodel "ctf-platform/internal/module/teaching_readmodel"
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

func TestRuntimeModuleContractsCompile(t *testing.T) {
	var _ runtime.RuntimeStatsProvider = (*runtime.Module)(nil)
	var _ runtime.RuntimeFacade = (*runtime.Module)(nil)
	var _ runtime.InstanceRepository = (*containerModule.Repository)(nil)
}

func TestOpsModuleContractsCompile(t *testing.T) {
	var _ ops.AuditRecorder = (*ops.Module)(nil)
}

func TestTeachingReadmodelModuleContractsCompile(t *testing.T) {
	var _ teachingreadmodel.TeachingQuery = (*teachingreadmodel.Module)(nil)
}

func TestPracticeReadmodelModuleContractsCompile(t *testing.T) {
	var _ practicereadmodel.PracticeQuery = (*practicereadmodel.Module)(nil)
}

func TestCompositionModulesExposeContracts(t *testing.T) {
	t.Parallel()

	assertFieldType(t, reflect.TypeOf(composition.AuthModule{}), "TokenService", reflect.TypeOf((*identity.Authenticator)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.PracticeReadmodelModule{}), "Query", reflect.TypeOf((*practicereadmodel.PracticeQuery)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.RuntimeModule{}), "Query", reflect.TypeOf((*runtime.RuntimeQuery)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.RuntimeModule{}), "Repository", reflect.TypeOf((*runtime.InstanceRepository)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.RuntimeModule{}), "Service", reflect.TypeOf((*runtime.RuntimeFacade)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.SystemModule{}), "AuditService", reflect.TypeOf((*ops.AuditRecorder)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.TeacherModule{}), "Query", reflect.TypeOf((*teachingreadmodel.TeachingQuery)(nil)).Elem())
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
