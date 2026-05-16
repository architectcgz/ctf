package config

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func chdirToBackendRoot(t *testing.T) {
	t.Helper()

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
}

func setContainerFlagSecretEnv(t *testing.T, secret string) string {
	t.Helper()

	secretFile := filepath.Join(t.TempDir(), "flag-global-secret")
	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", secret)
	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET_FILE", secretFile)
	return secretFile
}

func TestLoadReadsContainerFlagSecretFromEnv(t *testing.T) {
	chdirToBackendRoot(t)
	secretFile := setContainerFlagSecretEnv(t, "integration-secret-123456789012345")

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Container.FlagGlobalSecret != "integration-secret-123456789012345" {
		t.Fatalf("expected container flag secret from env, got %q", cfg.Container.FlagGlobalSecret)
	}
	persistedSecret, err := os.ReadFile(secretFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.TrimSpace(string(persistedSecret)) != cfg.Container.FlagGlobalSecret {
		t.Fatalf("expected persisted secret to match env secret, got %q", persistedSecret)
	}
}

func TestLoadReadsContainerRegistryCredentialsFromEnv(t *testing.T) {
	chdirToBackendRoot(t)
	setContainerFlagSecretEnv(t, "integration-secret-123456789012345")
	t.Setenv("CTF_CONTAINER_REGISTRY_ENABLED", "true")
	t.Setenv("CTF_CONTAINER_REGISTRY_SERVER", "registry.example.edu")
	t.Setenv("CTF_CONTAINER_REGISTRY_ACCESS_SERVER", "registry-internal:5000")
	t.Setenv("CTF_CONTAINER_ACCESS_HOST", "host-gateway.internal")
	t.Setenv("CTF_CONTAINER_REGISTRY_USERNAME", "ctf")
	t.Setenv("CTF_CONTAINER_REGISTRY_PASSWORD", "registry-token")

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Container.Registry.Enabled {
		t.Fatal("expected container registry to be enabled from env")
	}
	if cfg.Container.Registry.Server != "registry.example.edu" {
		t.Fatalf("registry server = %q, want registry.example.edu", cfg.Container.Registry.Server)
	}
	if cfg.Container.Registry.AccessServer != "registry-internal:5000" {
		t.Fatalf("registry access server = %q, want registry-internal:5000", cfg.Container.Registry.AccessServer)
	}
	if cfg.Container.AccessHost != "host-gateway.internal" {
		t.Fatalf("container access host = %q, want host-gateway.internal", cfg.Container.AccessHost)
	}
	if cfg.Container.Registry.Username != "ctf" {
		t.Fatalf("registry username = %q, want ctf", cfg.Container.Registry.Username)
	}
	if cfg.Container.Registry.Password != "registry-token" {
		t.Fatalf("registry password = %q, want registry-token", cfg.Container.Registry.Password)
	}
}

func TestValidateRejectsEnabledRegistryWithoutServer(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.Registry.Enabled = true
	cfg.Container.Registry.Username = "ctf"
	cfg.Container.Registry.Password = "registry-token"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject enabled registry without server, got nil")
	}
	if !strings.Contains(err.Error(), "container.registry.server must not be empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsEnabledRegistryWithoutCredentials(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.Registry.Enabled = true
	cfg.Container.Registry.Server = "registry.example.edu"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject enabled registry without credentials, got nil")
	}
	if !strings.Contains(err.Error(), "container.registry requires username/password or identity_token") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsLocalRegistryServerInContainerWithoutAccessServer(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.Registry.Enabled = true
	cfg.Container.Registry.Server = "127.0.0.1:5000"
	cfg.Container.Registry.Username = "ctf"
	cfg.Container.Registry.Password = "registry-token"

	previous := runningInContainer
	runningInContainer = func() bool { return true }
	t.Cleanup(func() {
		runningInContainer = previous
	})

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject localhost registry server in container without access_server, got nil")
	}
	if !strings.Contains(err.Error(), "container.registry.access_server must not be empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAllowsLocalRegistryServerOutsideContainerWithoutAccessServer(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.Registry.Enabled = true
	cfg.Container.Registry.Server = "127.0.0.1:5000"
	cfg.Container.Registry.Username = "ctf"
	cfg.Container.Registry.Password = "registry-token"

	previous := runningInContainer
	runningInContainer = func() bool { return false }
	t.Cleanup(func() {
		runningInContainer = previous
	})

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected Validate() to allow localhost registry server outside container, got %v", err)
	}
}

func TestLoadRejectsTooShortContainerFlagSecret(t *testing.T) {
	chdirToBackendRoot(t)
	setContainerFlagSecretEnv(t, "too-short-secret")

	_, err := Load("dev")
	if err == nil {
		t.Fatal("expected Load() to fail for short CTF_CONTAINER_FLAG_GLOBAL_SECRET, got nil")
	}
	if !strings.Contains(err.Error(), "at least 32 bytes") {
		t.Fatalf("expected short-secret validation error, got %v", err)
	}
}

func TestLoadRejectsCredentialedCORSWithoutAllowOrigins(t *testing.T) {
	chdirToBackendRoot(t)
	setContainerFlagSecretEnv(t, "integration-secret-123456789012345")

	_, err := Load("prod")
	if err == nil {
		t.Fatal("expected Load() to fail for credentialed CORS without allow_origins, got nil")
	}
	if !strings.Contains(err.Error(), "cors.allow_origins must not be empty") {
		t.Fatalf("expected credentialed CORS validation error, got %v", err)
	}
}

func TestLoadDevConfigDoesNotShipDefaultPasswords(t *testing.T) {
	chdirToBackendRoot(t)
	setContainerFlagSecretEnv(t, "integration-secret-123456789012345")

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
	if !cfg.Container.DefenseWorkbenchReadOnlyEnabled {
		t.Fatal("expected defense workbench readonly mode enabled in dev baseline config")
	}
	if cfg.Container.DefenseWorkbenchRoot != "/app" {
		t.Fatalf("expected defense workbench root /app in dev baseline config, got %q", cfg.Container.DefenseWorkbenchRoot)
	}
}

func TestLoadRestoresContainerFlagSecretFromPersistedFile(t *testing.T) {
	chdirToBackendRoot(t)

	secretFile := filepath.Join(t.TempDir(), "flag-global-secret")
	expectedSecret := "persisted-secret-12345678901234567890"
	if err := os.WriteFile(secretFile, []byte(expectedSecret+"\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", "")
	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET_FILE", secretFile)

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Container.FlagGlobalSecret != expectedSecret {
		t.Fatalf("expected persisted secret %q, got %q", expectedSecret, cfg.Container.FlagGlobalSecret)
	}
}

func TestLoadGeneratesAndPersistsContainerFlagSecretWhenMissing(t *testing.T) {
	chdirToBackendRoot(t)

	secretFile := filepath.Join(t.TempDir(), "flag-global-secret")
	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", "")
	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET_FILE", secretFile)

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if len(cfg.Container.FlagGlobalSecret) < 32 {
		t.Fatalf("expected generated secret length >= 32, got %d", len(cfg.Container.FlagGlobalSecret))
	}
	persistedSecret, err := os.ReadFile(secretFile)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.TrimSpace(string(persistedSecret)) != cfg.Container.FlagGlobalSecret {
		t.Fatalf("expected generated secret to be persisted, got %q", persistedSecret)
	}
}

func TestLoadRejectsMismatchedPersistedContainerFlagSecret(t *testing.T) {
	chdirToBackendRoot(t)

	secretFile := filepath.Join(t.TempDir(), "flag-global-secret")
	if err := os.WriteFile(secretFile, []byte("persisted-secret-12345678901234567890\n"), 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", "different-secret-12345678901234567890")
	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET_FILE", secretFile)

	_, err := Load("dev")
	if err == nil {
		t.Fatal("expected Load() to reject mismatched persisted container flag secret, got nil")
	}
	if !strings.Contains(err.Error(), "does not match persisted secret file") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsProductionPlaceholderSecrets(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.App.Env = "prod"
	cfg.Postgres.Password = "change_me"
	cfg.Redis.Password = "change_me"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject production placeholder secrets, got nil")
	}
	if !strings.Contains(err.Error(), "postgres.password must be provided from a non-placeholder secret in prod") {
		t.Fatalf("unexpected error: %v", err)
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

func TestValidateRejectsEnabledDefenseSSHWithoutHostKeyPath(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.DefenseSSHEnabled = true
	cfg.Container.DefenseSSHHost = "127.0.0.1"
	cfg.Container.DefenseSSHPort = 2222
	cfg.Container.DefenseSSHHostKeyPath = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject enabled defense ssh without host key path, got nil")
	}
	if !strings.Contains(err.Error(), "container.defense_ssh_host_key_path must not be empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateAllowsEnabledDefenseSSHWithHostKeyPath(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.DefenseSSHEnabled = true
	cfg.Container.DefenseSSHHost = "127.0.0.1"
	cfg.Container.DefenseSSHPort = 2222
	cfg.Container.DefenseSSHHostKeyPath = "storage/runtime/awd-defense-ssh-host-key.pem"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected Validate() to allow enabled defense ssh with host key path, got %v", err)
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
				CheckerSandbox: CheckerSandboxConfig{
					Image:            "python:3.12-alpine",
					User:             "65532:65532",
					WorkDir:          "/checker",
					Timeout:          10 * time.Second,
					CPUQuota:         0.5,
					MemoryBytes:      128 * 1024 * 1024,
					PidsLimit:        64,
					NofileLimit:      128,
					OutputLimitBytes: 32768,
				},
			},
		},
	}
}
