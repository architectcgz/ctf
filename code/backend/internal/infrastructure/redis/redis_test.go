package redis

import (
	"context"
	"strings"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/config"
)

func TestBuildSingleOptionsMapsRedisConfig(t *testing.T) {
	t.Parallel()

	cfg := config.RedisConfig{
		Mode:         "single",
		Addr:         "127.0.0.1:6379",
		Password:     "secret",
		DB:           4,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 2 * time.Second,
	}

	options := buildSingleOptions(cfg)
	if options.Addr != cfg.Addr {
		t.Fatalf("Addr = %q, want %q", options.Addr, cfg.Addr)
	}
	if options.Password != cfg.Password {
		t.Fatalf("Password = %q, want %q", options.Password, cfg.Password)
	}
	if options.DB != cfg.DB {
		t.Fatalf("DB = %d, want %d", options.DB, cfg.DB)
	}
	if options.DialTimeout != cfg.DialTimeout {
		t.Fatalf("DialTimeout = %s, want %s", options.DialTimeout, cfg.DialTimeout)
	}
	if options.ReadTimeout != cfg.ReadTimeout {
		t.Fatalf("ReadTimeout = %s, want %s", options.ReadTimeout, cfg.ReadTimeout)
	}
	if options.WriteTimeout != cfg.WriteTimeout {
		t.Fatalf("WriteTimeout = %s, want %s", options.WriteTimeout, cfg.WriteTimeout)
	}
}

func TestBuildFailoverOptionsMapsRedisConfig(t *testing.T) {
	t.Parallel()

	cfg := config.RedisConfig{
		Mode:             "sentinel",
		MasterName:       "mymaster",
		SentinelAddrs:    []string{"127.0.0.1:26379", "127.0.0.1:26380"},
		SentinelUsername: "sentinel-user",
		SentinelPassword: "sentinel-secret",
		Password:         "redis-secret",
		DB:               2,
		DialTimeout:      5 * time.Second,
		ReadTimeout:      4 * time.Second,
		WriteTimeout:     3 * time.Second,
	}

	options := buildFailoverOptions(cfg)
	if options.MasterName != cfg.MasterName {
		t.Fatalf("MasterName = %q, want %q", options.MasterName, cfg.MasterName)
	}
	if len(options.SentinelAddrs) != len(cfg.SentinelAddrs) {
		t.Fatalf("SentinelAddrs len = %d, want %d", len(options.SentinelAddrs), len(cfg.SentinelAddrs))
	}
	for i := range cfg.SentinelAddrs {
		if options.SentinelAddrs[i] != cfg.SentinelAddrs[i] {
			t.Fatalf("SentinelAddrs[%d] = %q, want %q", i, options.SentinelAddrs[i], cfg.SentinelAddrs[i])
		}
	}
	if options.SentinelUsername != cfg.SentinelUsername {
		t.Fatalf("SentinelUsername = %q, want %q", options.SentinelUsername, cfg.SentinelUsername)
	}
	if options.SentinelPassword != cfg.SentinelPassword {
		t.Fatalf("SentinelPassword = %q, want %q", options.SentinelPassword, cfg.SentinelPassword)
	}
	if options.Password != cfg.Password {
		t.Fatalf("Password = %q, want %q", options.Password, cfg.Password)
	}
	if options.DB != cfg.DB {
		t.Fatalf("DB = %d, want %d", options.DB, cfg.DB)
	}
	if options.DialTimeout != cfg.DialTimeout {
		t.Fatalf("DialTimeout = %s, want %s", options.DialTimeout, cfg.DialTimeout)
	}
	if options.ReadTimeout != cfg.ReadTimeout {
		t.Fatalf("ReadTimeout = %s, want %s", options.ReadTimeout, cfg.ReadTimeout)
	}
	if options.WriteTimeout != cfg.WriteTimeout {
		t.Fatalf("WriteTimeout = %s, want %s", options.WriteTimeout, cfg.WriteTimeout)
	}
}

func TestNewClientRequiresContext(t *testing.T) {
	t.Parallel()

	_, err := NewClient(nil, config.RedisConfig{Mode: "single", Addr: "127.0.0.1:6379"})
	if err == nil {
		t.Fatal("expected NewClient() to reject nil context, got nil")
	}
	if !strings.Contains(err.Error(), "redis client requires context") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNewClientSingleModeCanPingMiniredis(t *testing.T) {
	t.Parallel()

	mini := miniredis.RunT(t)
	client, err := NewClient(context.Background(), config.RedisConfig{
		Mode: "single",
		Addr: mini.Addr(),
	})
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("Ping() error = %v", err)
	}

	options := client.Options()
	if options == nil {
		t.Fatal("Options() returned nil")
	}
	if options.Addr != mini.Addr() {
		t.Fatalf("client addr = %q, want %q", options.Addr, mini.Addr())
	}
	if options.DB != 0 {
		t.Fatalf("client DB = %d, want 0", options.DB)
	}
}

var (
	_ *redislib.Options
	_ *redislib.FailoverOptions
)
