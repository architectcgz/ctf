package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestLoadReadsContainerFlagSecretFromEnv(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}

	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(backendRoot); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(currentDir)
	})

	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", "integration-secret-123456789012345")

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Container.FlagGlobalSecret != "integration-secret-123456789012345" {
		t.Fatalf("expected container flag secret from env, got %q", cfg.Container.FlagGlobalSecret)
	}
}

func TestLoadRejectsTooShortContainerFlagSecret(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}

	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(backendRoot); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(currentDir)
	})

	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", "too-short-secret")

	_, err = Load("dev")
	if err == nil {
		t.Fatal("expected Load() to fail for short CTF_CONTAINER_FLAG_GLOBAL_SECRET, got nil")
	}
	if !strings.Contains(err.Error(), "at least 32 bytes") {
		t.Fatalf("expected short-secret validation error, got %v", err)
	}
}

func TestLoadRejectsCredentialedCORSWithoutAllowOrigins(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}

	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(backendRoot); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(currentDir)
	})

	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", "integration-secret-123456789012345")

	_, err = Load("prod")
	if err == nil {
		t.Fatal("expected Load() to fail for credentialed CORS without allow_origins, got nil")
	}
	if !strings.Contains(err.Error(), "cors.allow_origins must not be empty") {
		t.Fatalf("expected credentialed CORS validation error, got %v", err)
	}
}

func TestLoadDevConfigDoesNotShipDefaultPasswords(t *testing.T) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}

	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(backendRoot); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(currentDir)
	})

	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", "integration-secret-123456789012345")

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Postgres.Password != "" {
		t.Fatalf("expected empty postgres password in dev baseline config, got %q", cfg.Postgres.Password)
	}
	if cfg.Redis.Password != "" {
		t.Fatalf("expected empty redis password in dev baseline config, got %q", cfg.Redis.Password)
	}
}

func TestValidateRejectsInvalidContainerPortRangeOrder(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.PortRangeStart = 40000
	cfg.Container.PortRangeEnd = 40000

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject equal container port range, got nil")
	}
	if !strings.Contains(err.Error(), "container.port_range_start must be less than container.port_range_end") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsInvalidContainerPortRangeBounds(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.PortRangeStart = 0
	cfg.Container.PortRangeEnd = 70000

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject out-of-range container ports, got nil")
	}
	if !strings.Contains(err.Error(), "container.port_range_start and container.port_range_end must be between 1 and 65535") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsNonPositiveContestSubmissionRateLimitTTL(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Contest.SubmissionRateLimitTTL = 0

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject non-positive contest submission rate limit ttl, got nil")
	}
	if !strings.Contains(err.Error(), "contest.submission_rate_limit_ttl must be greater than 0") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func validConfigForValidationTests() *Config {
	return &Config{
		CORS: CORSConfig{
			AllowOrigins:     []string{"https://academy.example.com"},
			AllowCredentials: true,
		},
		Container: ContainerConfig{
			DefaultCPUQuota:      1,
			DefaultMemory:        256 * 1024 * 1024,
			DefaultPidsLimit:     128,
			DefaultExposedPort:   8080,
			PortRangeStart:       30000,
			PortRangeEnd:         40000,
			OrphanGracePeriod:    time.Minute,
			CleanupLockTTL:       time.Minute,
			ProxyTicketTTL:       time.Minute,
			ProxyBodyPreviewSize: 1024,
		},
		Recommendation: RecommendationConfig{
			WeakThreshold: 0.4,
			CacheTTL:      time.Minute,
			DefaultLimit:  6,
			MaxLimit:      20,
		},
		Report: ReportConfig{
			StorageDir:      "storage/exports",
			DefaultFormat:   "pdf",
			PersonalTimeout: time.Minute,
			ClassTimeout:    2 * time.Minute,
			FileTTL:         time.Hour,
			MaxWorkers:      1,
		},
		Dashboard: DashboardConfig{
			CacheTTL:       time.Minute,
			AlertThreshold: 80,
		},
		WebSocket: WebSocketConfig{
			TicketTTL:         time.Minute,
			TicketKeyPrefix:   "ctf:ws:ticket",
			HeartbeatInterval: 30 * time.Second,
			ReadTimeout:       time.Minute,
			RetryInitialDelay: time.Second,
			RetryMaxDelay:     2 * time.Second,
		},
		Contest: ContestConfig{
			StatusUpdateInterval:   time.Minute,
			StatusUpdateBatchSize:  1,
			StatusUpdateLockTTL:    time.Minute,
			SubmissionRateLimitTTL: 5 * time.Second,
			AWD: ContestAWDConfig{
				SchedulerInterval:  time.Minute,
				SchedulerLockTTL:   time.Minute,
				SchedulerBatchSize: 1,
				RoundInterval:      time.Minute,
				RoundLockTTL:       time.Minute,
				PreviousRoundGrace: 0,
				CheckerTimeout:     time.Second,
			},
		},
	}
}
