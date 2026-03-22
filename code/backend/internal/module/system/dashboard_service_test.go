package system

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/runtime"
	rediskeys "ctf-platform/internal/pkg/redis"
)

type stubDashboardRuntimeQuery struct {
	countRunningFn func() (int64, error)
}

func (s *stubDashboardRuntimeQuery) CountRunning() (int64, error) {
	if s.countRunningFn == nil {
		return 0, nil
	}
	return s.countRunningFn()
}

type stubDashboardRuntimeStatsProvider struct {
	listManagedContainerStatsFn func(ctx context.Context) ([]runtime.ManagedContainerStat, error)
}

func (s *stubDashboardRuntimeStatsProvider) ListManagedContainerStats(ctx context.Context) ([]runtime.ManagedContainerStat, error) {
	if s.listManagedContainerStatsFn == nil {
		return nil, nil
	}
	return s.listManagedContainerStatsFn(ctx)
}

func setupDashboardTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Instance{}); err != nil {
		t.Fatalf("migrate instance: %v", err)
	}
	return db
}

func newDashboardTestService(t *testing.T, db *gorm.DB, redis *redislib.Client) *DashboardService {
	t.Helper()

	return NewDashboardService(
		runtime.NewQuery(runtime.NewRepository(db)),
		nil,
		redis,
		&config.Config{
			Dashboard: config.DashboardConfig{
				CacheTTL:       time.Minute,
				AlertThreshold: 80,
				RedisKeyPrefix: "dashboard:test",
			},
		},
		zap.NewNop(),
	)
}

func TestDashboardServiceGetDashboardStatsUsesCache(t *testing.T) {
	mr := miniredis.RunT(t)
	redis := redislib.NewClient(&redislib.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redis.Close() })

	service := newDashboardTestService(t, setupDashboardTestDB(t), redis)
	expected := dto.DashboardStats{
		OnlineUsers:      9,
		ActiveContainers: 4,
		CPUUsage:         71.5,
		MemoryUsage:      58.25,
		ContainerStats: []dto.ContainerStat{
			{ContainerID: "abc123", ContainerName: "practice-1", CPUPercent: 71.5, MemoryPercent: 58.25},
		},
		Alerts: []dto.ResourceAlert{
			{ContainerID: "abc123", Type: "cpu", Value: 71.5, Threshold: 70, Message: "cached"},
		},
	}
	payload, err := json.Marshal(expected)
	if err != nil {
		t.Fatalf("marshal cache payload: %v", err)
	}
	if err := redis.Set(context.Background(), service.getCacheKey(), payload, time.Minute).Err(); err != nil {
		t.Fatalf("seed cache: %v", err)
	}

	got, err := service.GetDashboardStats(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardStats() error = %v", err)
	}
	if got.OnlineUsers != expected.OnlineUsers || got.ActiveContainers != expected.ActiveContainers {
		t.Fatalf("expected cached stats %+v, got %+v", expected, got)
	}
	if len(got.ContainerStats) != 1 || got.ContainerStats[0].ContainerName != "practice-1" {
		t.Fatalf("expected cached container stats, got %+v", got.ContainerStats)
	}
}

func TestDashboardServiceGetDashboardStatsComputesAndCachesSummary(t *testing.T) {
	db := setupDashboardTestDB(t)
	now := time.Now()
	instances := []model.Instance{
		{ID: 1, UserID: 1, ChallengeID: 11, ContainerID: "cont-1", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now},
		{ID: 2, UserID: 2, ChallengeID: 12, ContainerID: "cont-2", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now},
		{ID: 3, UserID: 3, ChallengeID: 13, ContainerID: "cont-3", Status: model.InstanceStatusStopped, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now},
	}
	for _, instance := range instances {
		if err := db.Create(&instance).Error; err != nil {
			t.Fatalf("seed instance: %v", err)
		}
	}

	mr := miniredis.RunT(t)
	redis := redislib.NewClient(&redislib.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redis.Close() })
	if err := redis.Set(context.Background(), rediskeys.TokenKey(101), "token-1", time.Hour).Err(); err != nil {
		t.Fatalf("seed token 1: %v", err)
	}
	if err := redis.Set(context.Background(), rediskeys.TokenKey(102), "token-2", time.Hour).Err(); err != nil {
		t.Fatalf("seed token 2: %v", err)
	}

	service := newDashboardTestService(t, db, redis)

	got, err := service.GetDashboardStats(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardStats() error = %v", err)
	}
	if got.OnlineUsers != 2 {
		t.Fatalf("expected 2 online users, got %+v", got)
	}
	if got.ActiveContainers != 2 {
		t.Fatalf("expected 2 active containers, got %+v", got)
	}
	if len(got.ContainerStats) != 0 || len(got.Alerts) != 0 {
		t.Fatalf("expected empty docker stats when client is nil, got %+v", got)
	}

	cached, err := redis.Get(context.Background(), service.getCacheKey()).Bytes()
	if err != nil {
		t.Fatalf("expected stats cached, get error = %v", err)
	}
	var cachedStats dto.DashboardStats
	if err := json.Unmarshal(cached, &cachedStats); err != nil {
		t.Fatalf("unmarshal cached stats: %v", err)
	}
	if cachedStats.OnlineUsers != 2 || cachedStats.ActiveContainers != 2 {
		t.Fatalf("unexpected cached stats: %+v", cachedStats)
	}
}

func TestDashboardServiceCheckAlertsReturnsCPUAndMemoryAlerts(t *testing.T) {
	service := newDashboardTestService(t, setupDashboardTestDB(t), nil)

	alerts := service.checkAlerts([]dto.ContainerStat{
		{ContainerID: "abc123", ContainerName: "practice-1", CPUPercent: 91, MemoryPercent: 83},
		{ContainerID: "def456", ContainerName: "practice-2", CPUPercent: 45, MemoryPercent: 30},
	})

	if len(alerts) != 2 {
		t.Fatalf("expected 2 alerts, got %+v", alerts)
	}
	if alerts[0].Type != "cpu" || alerts[1].Type != "memory" {
		t.Fatalf("expected cpu then memory alerts, got %+v", alerts)
	}
	if alerts[0].ContainerID != "abc123" || alerts[1].ContainerID != "abc123" {
		t.Fatalf("expected alerts for hot container only, got %+v", alerts)
	}
}

func TestDashboardServiceUsesRuntimeStatsProvider(t *testing.T) {
	var newDashboardService func(runtime.RuntimeQuery, runtime.RuntimeStatsProvider, *redislib.Client, *config.Config, *zap.Logger) *DashboardService = NewDashboardService

	service := newDashboardService(
		&stubDashboardRuntimeQuery{
			countRunningFn: func() (int64, error) {
				return 3, nil
			},
		},
		&stubDashboardRuntimeStatsProvider{
			listManagedContainerStatsFn: func(ctx context.Context) ([]runtime.ManagedContainerStat, error) {
				return []runtime.ManagedContainerStat{
					{
						ContainerID:   "runtime-1",
						ContainerName: "runtime-web",
						CPUPercent:    42.5,
						MemoryPercent: 63.25,
						MemoryUsage:   128,
						MemoryLimit:   256,
					},
				}, nil
			},
		},
		nil,
		&config.Config{
			Dashboard: config.DashboardConfig{
				CacheTTL:       time.Minute,
				AlertThreshold: 80,
				RedisKeyPrefix: "dashboard:test",
			},
		},
		zap.NewNop(),
	)

	got, err := service.GetDashboardStats(context.Background())
	if err != nil {
		t.Fatalf("GetDashboardStats() error = %v", err)
	}
	if got.ActiveContainers != 3 {
		t.Fatalf("expected active containers from runtime query, got %+v", got)
	}
	if len(got.ContainerStats) != 1 || got.ContainerStats[0].ContainerID != "runtime-1" {
		t.Fatalf("expected runtime stats provider output, got %+v", got.ContainerStats)
	}
}
