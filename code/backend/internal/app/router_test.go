package app

import (
	"context"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/app/composition"
	"ctf-platform/internal/config"
	"ctf-platform/internal/module/identity"
	"ctf-platform/internal/module/ops"
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
}

func TestOpsModuleContractsCompile(t *testing.T) {
	var _ ops.AuditRecorder = (*ops.Module)(nil)
}

func TestTeachingReadmodelModuleContractsCompile(t *testing.T) {
	var _ teachingreadmodel.TeachingQuery = (*teachingreadmodel.Module)(nil)
}

func TestCompositionModulesExposeContracts(t *testing.T) {
	t.Parallel()

	assertFieldType(t, reflect.TypeOf(composition.AuthModule{}), "TokenService", reflect.TypeOf((*identity.Authenticator)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.ContainerModule{}), "Service", reflect.TypeOf((*runtime.RuntimeFacade)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.SystemModule{}), "AuditService", reflect.TypeOf((*ops.AuditRecorder)(nil)).Elem())
	assertFieldType(t, reflect.TypeOf(composition.TeacherModule{}), "Query", reflect.TypeOf((*teachingreadmodel.TeachingQuery)(nil)).Elem())
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
