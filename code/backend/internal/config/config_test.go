package config

import (
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	startuprecovery "ctf-platform/internal/module/instance/startuprecovery"
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

func TestValidateRejectsInvalidSingleContainerSubnetBase(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.Network.SingleContainerSubnetBase = "10.11.0.1/16"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject non-network subnet base, got nil")
	}
	if !strings.Contains(err.Error(), "container.network.single_container_subnet_base must use the network address") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsSingleContainerSubnetMaskNotNarrowerThanBase(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.Network.SingleContainerSubnetMask = 16

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject non-narrower subnet mask, got nil")
	}
	if !strings.Contains(err.Error(), "container.network.single_container_subnet_mask must be greater than the base CIDR prefix") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsOverlappingSubnetPools(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.Network.SingleContainerSubnetBase = "10.10.1.0/24"
	cfg.Container.Network.SingleContainerSubnetMask = 29

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject overlapping subnet pools, got nil")
	}
	if !strings.Contains(err.Error(), "container.network.single_container_subnet_base and container.network.topology_subnet_base must not overlap") {
		t.Fatalf("unexpected error: %v", err)
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

func TestLoadRuntimeAgentDoesNotRequireAPIProductionSecrets(t *testing.T) {
	chdirToBackendRoot(t)
	t.Setenv("CTF_RUNTIME_AGENT_SERVER_ENABLED", "true")
	t.Setenv("CTF_RUNTIME_AGENT_SERVER_CERT_FILE", "/etc/ctf/runtime-agent/server.pem")
	t.Setenv("CTF_RUNTIME_AGENT_SERVER_KEY_FILE", "/etc/ctf/runtime-agent/server-key.pem")
	t.Setenv("CTF_RUNTIME_AGENT_SERVER_CLIENT_CA_FILE", "/etc/ctf/runtime-agent/ca.pem")
	t.Setenv("CTF_RUNTIME_AGENT_SERVER_KEEPALIVE_MIN_TIME", "45s")

	cfg, err := LoadRuntimeAgent("prod")
	if err != nil {
		t.Fatalf("LoadRuntimeAgent() error = %v", err)
	}
	if !cfg.RuntimeAgent.Server.Enabled {
		t.Fatal("runtime agent server should be enabled from env")
	}
	if cfg.RuntimeAgent.Server.Port != 9443 {
		t.Fatalf("runtime agent server port = %d, want 9443", cfg.RuntimeAgent.Server.Port)
	}
	if cfg.RuntimeAgent.Server.KeepaliveMinTime != 45*time.Second {
		t.Fatalf("runtime agent server keepalive min time = %s, want 45s", cfg.RuntimeAgent.Server.KeepaliveMinTime)
	}
}

func TestLoadReadsRuntimeAgentClientKeepaliveFromEnv(t *testing.T) {
	chdirToBackendRoot(t)
	setContainerFlagSecretEnv(t, "integration-secret-123456789012345")
	t.Setenv("CTF_RUNTIME_AGENT_KEEPALIVE_TIME", "45s")
	t.Setenv("CTF_RUNTIME_AGENT_KEEPALIVE_TIMEOUT", "7s")

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.RuntimeAgent.KeepaliveTime != 45*time.Second {
		t.Fatalf("runtime agent keepalive time = %s, want 45s", cfg.RuntimeAgent.KeepaliveTime)
	}
	if cfg.RuntimeAgent.KeepaliveTimeout != 7*time.Second {
		t.Fatalf("runtime agent keepalive timeout = %s, want 7s", cfg.RuntimeAgent.KeepaliveTimeout)
	}
}

func TestLoadReadsRuntimeAgentNodeNameFromEnv(t *testing.T) {
	chdirToBackendRoot(t)
	setContainerFlagSecretEnv(t, "integration-secret-123456789012345")
	t.Setenv("CTF_RUNTIME_AGENT_NODE_NAME", "runtime-node-a")

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.RuntimeAgent.NodeName != "runtime-node-a" {
		t.Fatalf("runtime agent node name = %q, want runtime-node-a", cfg.RuntimeAgent.NodeName)
	}
}

func TestLoadRuntimeAgentReadsServerNodeNameFromEnv(t *testing.T) {
	chdirToBackendRoot(t)
	t.Setenv("CTF_RUNTIME_AGENT_SERVER_ENABLED", "true")
	t.Setenv("CTF_RUNTIME_AGENT_SERVER_CERT_FILE", "/etc/ctf/runtime-agent/server.pem")
	t.Setenv("CTF_RUNTIME_AGENT_SERVER_KEY_FILE", "/etc/ctf/runtime-agent/server-key.pem")
	t.Setenv("CTF_RUNTIME_AGENT_SERVER_CLIENT_CA_FILE", "/etc/ctf/runtime-agent/ca.pem")
	t.Setenv("CTF_RUNTIME_AGENT_SERVER_NODE_NAME", "runtime-node-a")

	cfg, err := LoadRuntimeAgent("dev")
	if err != nil {
		t.Fatalf("LoadRuntimeAgent() error = %v", err)
	}

	if cfg.RuntimeAgent.Server.NodeName != "runtime-node-a" {
		t.Fatalf("runtime agent server node name = %q, want runtime-node-a", cfg.RuntimeAgent.Server.NodeName)
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

func TestLoadDevConfigOverridesPracticeSchedulerThroughput(t *testing.T) {
	chdirToBackendRoot(t)
	setContainerFlagSecretEnv(t, "integration-secret-123456789012345")

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Container.Scheduler.BatchSize != 12 {
		t.Fatalf("expected dev scheduler batch_size=12, got %d", cfg.Container.Scheduler.BatchSize)
	}
	if cfg.Container.Scheduler.MaxConcurrentStarts != 12 {
		t.Fatalf("expected dev scheduler max_concurrent_starts=12, got %d", cfg.Container.Scheduler.MaxConcurrentStarts)
	}
	if cfg.Container.Scheduler.MaxActiveInstances != 120 {
		t.Fatalf("expected dev scheduler max_active_instances=120, got %d", cfg.Container.Scheduler.MaxActiveInstances)
	}
}

func TestLoadAppliesRuntimeNodeHealthDefaults(t *testing.T) {
	chdirToBackendRoot(t)
	setContainerFlagSecretEnv(t, "integration-secret-123456789012345")

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if !cfg.Container.RuntimeNodeHealth.Enabled {
		t.Fatal("expected runtime node health to be enabled by default")
	}
	if cfg.Container.RuntimeNodeHealth.PollInterval != 10*time.Second {
		t.Fatalf("poll interval = %s, want 10s", cfg.Container.RuntimeNodeHealth.PollInterval)
	}
	if cfg.Container.RuntimeNodeHealth.ProbeTimeout != 2*time.Second {
		t.Fatalf("probe timeout = %s, want 2s", cfg.Container.RuntimeNodeHealth.ProbeTimeout)
	}
	if cfg.Container.RuntimeNodeHealth.StaleAfter != 30*time.Second {
		t.Fatalf("stale after = %s, want 30s", cfg.Container.RuntimeNodeHealth.StaleAfter)
	}
	if cfg.Container.RuntimeNodeHealth.FailureThreshold != 3 {
		t.Fatalf("failure threshold = %d, want 3", cfg.Container.RuntimeNodeHealth.FailureThreshold)
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

func TestResolveContainerFlagSecretRejectsProductionAutoGeneration(t *testing.T) {
	t.Parallel()

	secretFile := filepath.Join(t.TempDir(), "flag-global-secret")

	_, err := resolveContainerFlagGlobalSecret("", secretFile, false)
	if err == nil {
		t.Fatal("expected production secret resolution to reject missing secret auto-generation")
	}
	if !strings.Contains(err.Error(), "must be explicitly configured") {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(secretFile); !os.IsNotExist(statErr) {
		t.Fatalf("expected secret file not to be generated, stat error = %v", statErr)
	}
}

func TestContainerFlagSecretKeyringIncludesActiveAndPreviousKeys(t *testing.T) {
	t.Parallel()

	cfg := &Config{
		Container: ContainerConfig{
			FlagGlobalSecret:      "active-secret-12345678901234567890",
			FlagGlobalSecretKeyID: "active",
			FlagGlobalSecretKeyring: []ContainerFlagSecretKeyConfig{
				{KeyID: "previous", Secret: "previous-secret-123456789012345678"},
			},
		},
	}

	if err := cfg.resolveContainerFlagSecretKeyring(); err != nil {
		t.Fatalf("resolveContainerFlagSecretKeyring() error = %v", err)
	}
	if cfg.Container.ResolvedFlagSecretKeyID != "active" {
		t.Fatalf("active key id = %q, want active", cfg.Container.ResolvedFlagSecretKeyID)
	}
	if got := cfg.Container.ResolvedFlagSecrets["active"]; got != "active-secret-12345678901234567890" {
		t.Fatalf("active secret = %q", got)
	}
	if got := cfg.Container.ResolvedFlagSecrets["previous"]; got != "previous-secret-123456789012345678" {
		t.Fatalf("previous secret = %q", got)
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

func TestValidateRejectsUnsupportedRedisMode(t *testing.T) {
	t.Parallel()

	cfg := validConfigForValidationTests()
	cfg.Redis.Mode = "cluster"
	cfg.Redis.Addr = "127.0.0.1:6379"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject unsupported redis mode, got nil")
	}
	if !strings.Contains(err.Error(), "redis.mode") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsSingleRedisModeWithoutAddr(t *testing.T) {
	t.Parallel()

	cfg := validConfigForValidationTests()
	cfg.Redis.Mode = "single"
	cfg.Redis.Addr = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject single redis mode without addr, got nil")
	}
	if !strings.Contains(err.Error(), "redis.addr") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsSentinelRedisModeWithoutMasterName(t *testing.T) {
	t.Parallel()

	cfg := validConfigForValidationTests()
	cfg.Redis.Mode = "sentinel"
	cfg.Redis.MasterName = ""
	cfg.Redis.SentinelAddrs = []string{"127.0.0.1:26379"}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject sentinel redis mode without master name, got nil")
	}
	if !strings.Contains(err.Error(), "redis.master_name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsSentinelRedisModeWithoutSentinelAddrs(t *testing.T) {
	t.Parallel()

	cfg := validConfigForValidationTests()
	cfg.Redis.Mode = "sentinel"
	cfg.Redis.MasterName = "mymaster"
	cfg.Redis.SentinelAddrs = nil

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject sentinel redis mode without sentinel addrs, got nil")
	}
	if !strings.Contains(err.Error(), "redis.sentinel_addrs") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsRedisClusterModeEvenWhenReserveFieldsArePresent(t *testing.T) {
	t.Parallel()

	cfg := validConfigForValidationTests()
	cfg.Redis.Mode = "cluster"
	cfg.Redis.Cluster.Addrs = []string{"10.0.1.10:6379", "10.0.1.11:6379"}
	cfg.Redis.Cluster.RouteByLatency = true
	cfg.Redis.Cluster.RouteRandomly = true

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to keep rejecting redis cluster mode, got nil")
	}
	if !strings.Contains(err.Error(), "redis.mode") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPostgresDSNIncludesUTCAndEscapesKeywordValues(t *testing.T) {
	t.Parallel()

	cfg := PostgresConfig{
		Host:     "db.internal",
		Port:     5432,
		Database: "ctf db",
		Username: "ctf user",
		Password: "p'ass\\word value",
		SSLMode:  "require",
	}

	dsn := cfg.DSN()
	if !strings.Contains(dsn, "TimeZone=UTC") {
		t.Fatalf("DSN() = %q, want TimeZone=UTC", dsn)
	}

	parsed, err := pgconn.ParseConfig(dsn)
	if err != nil {
		t.Fatalf("ParseConfig(DSN()) error = %v, dsn = %q", err, dsn)
	}
	if parsed.Host != cfg.Host {
		t.Fatalf("parsed host = %q, want %q", parsed.Host, cfg.Host)
	}
	if parsed.Database != cfg.Database {
		t.Fatalf("parsed database = %q, want %q", parsed.Database, cfg.Database)
	}
	if parsed.User != cfg.Username {
		t.Fatalf("parsed user = %q, want %q", parsed.User, cfg.Username)
	}
	if parsed.Password != cfg.Password {
		t.Fatalf("parsed password = %q, want %q", parsed.Password, cfg.Password)
	}
	if got := runtimeParamValue(parsed.RuntimeParams, "TimeZone"); got != "UTC" {
		t.Fatalf("runtime TimeZone = %q, want UTC", got)
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

func TestValidateRejectsSharedFSWithoutRoot(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.SharedStorage.Type = "shared_fs"
	cfg.SharedStorage.SharedFS.Root = ""

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject shared_fs without root, got nil")
	}
	if !strings.Contains(err.Error(), "shared_storage.shared_fs.root must not be empty") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestConfigSharedStoragePathResolvesRelativePath(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.SharedStorage.Type = "shared_fs"
	cfg.SharedStorage.SharedFS.Root = filepath.Join(t.TempDir(), "shared")

	resolved := cfg.SharedStoragePath("runtime/flag-global-secret")
	want := filepath.Join(cfg.SharedStorage.SharedFS.Root, "runtime", "flag-global-secret")
	if resolved != want {
		t.Fatalf("resolved = %q, want %q", resolved, want)
	}
}

func TestLoadAppliesSharedStorageDefaults(t *testing.T) {
	chdirToBackendRoot(t)
	setContainerFlagSecretEnv(t, "integration-secret-123456789012345")

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.SharedStorage.Type != "shared_fs" {
		t.Fatalf("shared storage type = %q", cfg.SharedStorage.Type)
	}
	if cfg.SharedStorage.SharedFS.Root != "storage/shared" {
		t.Fatalf("shared storage root = %q", cfg.SharedStorage.SharedFS.Root)
	}
}

func TestLoadResolvesContainerFlagSecretFileRelativeToSharedStorageRoot(t *testing.T) {
	chdirToBackendRoot(t)
	sharedRoot := filepath.Join(t.TempDir(), "shared")
	t.Setenv("CTF_SHARED_STORAGE_SHARED_FS_ROOT", sharedRoot)
	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", "integration-secret-123456789012345")
	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET_FILE", "runtime/flag-global-secret")

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := filepath.Join(sharedRoot, "runtime", "flag-global-secret")
	if cfg.Container.FlagGlobalSecretFile != want {
		t.Fatalf("flag secret file = %q, want %q", cfg.Container.FlagGlobalSecretFile, want)
	}
	persistedSecret, err := os.ReadFile(want)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if strings.TrimSpace(string(persistedSecret)) != cfg.Container.FlagGlobalSecret {
		t.Fatalf("expected persisted secret to match resolved secret, got %q", persistedSecret)
	}
}

func TestLoadResolvesDefenseSSHHostKeyPathRelativeToSharedStorageRoot(t *testing.T) {
	chdirToBackendRoot(t)
	sharedRoot := filepath.Join(t.TempDir(), "shared")
	t.Setenv("CTF_SHARED_STORAGE_SHARED_FS_ROOT", sharedRoot)
	t.Setenv("CTF_CONTAINER_FLAG_GLOBAL_SECRET", "integration-secret-123456789012345")

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	want := filepath.Join(sharedRoot, "runtime", "awd-defense-ssh-host-key.pem")
	if cfg.Container.DefenseSSHHostKeyPath != want {
		t.Fatalf("defense ssh host key path = %q, want %q", cfg.Container.DefenseSSHHostKeyPath, want)
	}
}

func TestValidateRejectsTooLargeStartupRecoveryLockTTL(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.StartupRecoveryLockTTL = containerStartupRecoveryMaxLockTTL + time.Second

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject oversized startup recovery lock ttl")
	}
	if got := err.Error(); got != "container.startup_recovery_lock_ttl must be less than or equal to "+containerStartupRecoveryMaxLockTTL.String() {
		t.Fatalf("unexpected error: %s", got)
	}
}

func TestValidateAllowsEnabledDefenseSSHWithHostKeyPath(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Container.DefenseSSHEnabled = true
	cfg.Container.DefenseSSHHost = "127.0.0.1"
	cfg.Container.DefenseSSHPort = 2222
	cfg.Container.DefenseSSHHostKeyPath = "runtime/awd-defense-ssh-host-key.pem"

	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected Validate() to allow enabled defense ssh with host key path, got %v", err)
	}
}

func TestValidateRejectsInvalidRuntimeNodeHealthConfig(t *testing.T) {
	testCases := []struct {
		name      string
		mutate    func(*Config)
		wantError string
	}{
		{
			name: "poll interval",
			mutate: func(cfg *Config) {
				cfg.Container.RuntimeNodeHealth.PollInterval = 0
			},
			wantError: "container.runtime_node_health.poll_interval must be greater than 0",
		},
		{
			name: "probe timeout",
			mutate: func(cfg *Config) {
				cfg.Container.RuntimeNodeHealth.ProbeTimeout = 0
			},
			wantError: "container.runtime_node_health.probe_timeout must be greater than 0",
		},
		{
			name: "stale after",
			mutate: func(cfg *Config) {
				cfg.Container.RuntimeNodeHealth.StaleAfter = time.Second
				cfg.Container.RuntimeNodeHealth.ProbeTimeout = 2 * time.Second
			},
			wantError: "container.runtime_node_health.stale_after must be greater than probe_timeout",
		},
		{
			name: "failure threshold",
			mutate: func(cfg *Config) {
				cfg.Container.RuntimeNodeHealth.FailureThreshold = 0
			},
			wantError: "container.runtime_node_health.failure_threshold must be greater than 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validConfigForValidationTests()
			tc.mutate(cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatal("expected Validate() to reject invalid runtime node health config, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
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

func TestAuthConfigDoesNotExposeLegacyMCPTokenTTL(t *testing.T) {
	if _, ok := reflect.TypeOf(AuthConfig{}).FieldByName("MCPTokenTTL"); ok {
		t.Fatal("AuthConfig must not expose legacy MCPTokenTTL after OAuth migration")
	}
}

func TestLoadAppliesOAuthDefaults(t *testing.T) {
	chdirToBackendRoot(t)

	cfg, err := Load("dev")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Auth.OAuth.AuthorizationCodeTTL != 5*time.Minute {
		t.Fatalf("authorization code ttl = %s, want 5m", cfg.Auth.OAuth.AuthorizationCodeTTL)
	}
	if cfg.Auth.OAuth.AccessTokenTTL != 15*time.Minute {
		t.Fatalf("access token ttl = %s, want 15m", cfg.Auth.OAuth.AccessTokenTTL)
	}
	if cfg.Auth.OAuth.RefreshTokenTTL != 30*24*time.Hour {
		t.Fatalf("refresh token ttl = %s, want 720h", cfg.Auth.OAuth.RefreshTokenTTL)
	}
	if !cfg.Auth.OAuth.ClientRegistrationEnabled {
		t.Fatal("client registration should be enabled by default")
	}
	if cfg.Auth.OAuth.RedisKeyPrefix != "ctf:auth:oauth" {
		t.Fatalf("oauth redis key prefix = %q", cfg.Auth.OAuth.RedisKeyPrefix)
	}
}

func TestValidateOAuthRejectsInvalidProductionIssuer(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.App.Env = "prod"
	cfg.Auth.OAuth.IssuerURL = "http://ctf.example.edu"

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject non-HTTPS oauth issuer in production, got nil")
	}
	if !strings.Contains(err.Error(), "auth.oauth.issuer_url must be an https origin in prod") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateOAuthRejectsEmptyRedirectPrefix(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.Auth.OAuth.AllowedRedirectURIPrefixes = []string{"https://agent.example.edu/callback", " "}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject empty oauth redirect prefix, got nil")
	}
	if !strings.Contains(err.Error(), "auth.oauth.allowed_redirect_uri_prefixes must not contain empty prefix") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsInvalidEnabledMCPRateLimit(t *testing.T) {
	testCases := []struct {
		name      string
		mutate    func(*Config)
		wantError string
	}{
		{
			name: "limit",
			mutate: func(cfg *Config) {
				cfg.RateLimit.MCP.Enabled = true
				cfg.RateLimit.MCP.Limit = 0
				cfg.RateLimit.MCP.Window = time.Minute
			},
			wantError: "rate_limit.mcp.limit must be greater than 0 when enabled",
		},
		{
			name: "window",
			mutate: func(cfg *Config) {
				cfg.RateLimit.MCP.Enabled = true
				cfg.RateLimit.MCP.Limit = 1
				cfg.RateLimit.MCP.Window = 0
			},
			wantError: "rate_limit.mcp.window must be greater than 0 when enabled",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validConfigForValidationTests()
			tc.mutate(cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatal("expected Validate() to reject invalid MCP rate limit, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateRejectsIncompleteRuntimeAgentClientConfig(t *testing.T) {
	testCases := []struct {
		name      string
		mutate    func(*Config)
		wantError string
	}{
		{
			name: "missing endpoint",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Enabled = true
			},
			wantError: "runtime_agent.endpoint must not be empty",
		},
		{
			name: "missing server name",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Enabled = true
				cfg.RuntimeAgent.Endpoint = "127.0.0.1:9443"
			},
			wantError: "runtime_agent.server_name must not be empty",
		},
		{
			name: "missing ca file",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Enabled = true
				cfg.RuntimeAgent.Endpoint = "127.0.0.1:9443"
				cfg.RuntimeAgent.ServerName = "runtime-agent.local"
			},
			wantError: "runtime_agent.ca_file must not be empty",
		},
		{
			name: "missing client cert file",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Enabled = true
				cfg.RuntimeAgent.Endpoint = "127.0.0.1:9443"
				cfg.RuntimeAgent.ServerName = "runtime-agent.local"
				cfg.RuntimeAgent.CAFile = "/etc/ctf/ca.pem"
			},
			wantError: "runtime_agent.cert_file must not be empty",
		},
		{
			name: "missing client key file",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Enabled = true
				cfg.RuntimeAgent.Endpoint = "127.0.0.1:9443"
				cfg.RuntimeAgent.ServerName = "runtime-agent.local"
				cfg.RuntimeAgent.CAFile = "/etc/ctf/ca.pem"
				cfg.RuntimeAgent.CertFile = "/etc/ctf/client.pem"
			},
			wantError: "runtime_agent.key_file must not be empty",
		},
		{
			name: "missing dial timeout",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Enabled = true
				cfg.RuntimeAgent.Endpoint = "127.0.0.1:9443"
				cfg.RuntimeAgent.ServerName = "runtime-agent.local"
				cfg.RuntimeAgent.CAFile = "/etc/ctf/ca.pem"
				cfg.RuntimeAgent.CertFile = "/etc/ctf/client.pem"
				cfg.RuntimeAgent.KeyFile = "/etc/ctf/client-key.pem"
				cfg.RuntimeAgent.DialTimeout = 0
			},
			wantError: "runtime_agent.dial_timeout must be greater than 0",
		},
		{
			name: "keepalive time too small",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Enabled = true
				cfg.RuntimeAgent.Endpoint = "127.0.0.1:9443"
				cfg.RuntimeAgent.ServerName = "runtime-agent.local"
				cfg.RuntimeAgent.CAFile = "/etc/ctf/ca.pem"
				cfg.RuntimeAgent.CertFile = "/etc/ctf/client.pem"
				cfg.RuntimeAgent.KeyFile = "/etc/ctf/client-key.pem"
				cfg.RuntimeAgent.KeepaliveTime = 5 * time.Second
			},
			wantError: "runtime_agent.keepalive_time must be at least 10s",
		},
		{
			name: "missing keepalive timeout",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Enabled = true
				cfg.RuntimeAgent.Endpoint = "127.0.0.1:9443"
				cfg.RuntimeAgent.ServerName = "runtime-agent.local"
				cfg.RuntimeAgent.CAFile = "/etc/ctf/ca.pem"
				cfg.RuntimeAgent.CertFile = "/etc/ctf/client.pem"
				cfg.RuntimeAgent.KeyFile = "/etc/ctf/client-key.pem"
				cfg.RuntimeAgent.KeepaliveTimeout = 0
			},
			wantError: "runtime_agent.keepalive_timeout must be greater than 0",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validConfigForValidationTests()
			tc.mutate(cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatalf("expected Validate() to reject incomplete runtime agent client config, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRuntimeAgentLocalFallbackDefaultsFalse(t *testing.T) {
	cfg := validConfigForValidationTests()

	if cfg.RuntimeAgent.AllowLocalFallback {
		t.Fatal("expected runtime_agent.allow_local_fallback to default false")
	}
}

func TestValidateRejectsRuntimeAgentLocalFallbackInProduction(t *testing.T) {
	cfg := validConfigForValidationTests()
	cfg.App.Env = "prod"
	cfg.Postgres.Password = "prod-postgres-secret"
	cfg.Redis.Password = "prod-redis-secret"
	cfg.RuntimeAgent.AllowLocalFallback = true

	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected Validate() to reject runtime agent local fallback in production, got nil")
	}
	if !strings.Contains(err.Error(), "runtime_agent.allow_local_fallback must be false in prod") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidateRejectsIncompleteRuntimeAgentServerConfig(t *testing.T) {
	testCases := []struct {
		name      string
		mutate    func(*Config)
		wantError string
	}{
		{
			name: "missing port",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Server.Enabled = true
				cfg.RuntimeAgent.Server.Port = 0
			},
			wantError: "runtime_agent.server.port must be between 1 and 65535",
		},
		{
			name: "missing cert file",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Server.Enabled = true
				cfg.RuntimeAgent.Server.Port = 9443
			},
			wantError: "runtime_agent.server.cert_file must not be empty",
		},
		{
			name: "missing key file",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Server.Enabled = true
				cfg.RuntimeAgent.Server.Port = 9443
				cfg.RuntimeAgent.Server.CertFile = "/etc/ctf/runtime-agent/server.pem"
			},
			wantError: "runtime_agent.server.key_file must not be empty",
		},
		{
			name: "missing client ca file",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Server.Enabled = true
				cfg.RuntimeAgent.Server.Port = 9443
				cfg.RuntimeAgent.Server.CertFile = "/etc/ctf/runtime-agent/server.pem"
				cfg.RuntimeAgent.Server.KeyFile = "/etc/ctf/runtime-agent/server-key.pem"
			},
			wantError: "runtime_agent.server.client_ca_file must not be empty",
		},
		{
			name: "missing shutdown timeout",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Server.Enabled = true
				cfg.RuntimeAgent.Server.Port = 9443
				cfg.RuntimeAgent.Server.CertFile = "/etc/ctf/runtime-agent/server.pem"
				cfg.RuntimeAgent.Server.KeyFile = "/etc/ctf/runtime-agent/server-key.pem"
				cfg.RuntimeAgent.Server.ClientCAFile = "/etc/ctf/runtime-agent/ca.pem"
				cfg.RuntimeAgent.Server.ShutdownTimeout = 0
			},
			wantError: "runtime_agent.server.shutdown_timeout must be greater than 0",
		},
		{
			name: "keepalive min time too small",
			mutate: func(cfg *Config) {
				cfg.RuntimeAgent.Server.Enabled = true
				cfg.RuntimeAgent.Server.Port = 9443
				cfg.RuntimeAgent.Server.CertFile = "/etc/ctf/runtime-agent/server.pem"
				cfg.RuntimeAgent.Server.KeyFile = "/etc/ctf/runtime-agent/server-key.pem"
				cfg.RuntimeAgent.Server.ClientCAFile = "/etc/ctf/runtime-agent/ca.pem"
				cfg.RuntimeAgent.Server.KeepaliveMinTime = 5 * time.Second
			},
			wantError: "runtime_agent.server.keepalive_min_time must be at least 10s",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validConfigForValidationTests()
			tc.mutate(cfg)

			err := cfg.Validate()
			if err == nil {
				t.Fatalf("expected Validate() to reject incomplete runtime agent server config, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func validConfigForValidationTests() *Config {
	return &Config{
		Auth: AuthConfig{
			OAuth: AuthOAuthConfig{
				AuthorizationCodeTTL:      5 * time.Minute,
				AccessTokenTTL:            15 * time.Minute,
				RefreshTokenTTL:           30 * 24 * time.Hour,
				ClientRegistrationEnabled: true,
				RedisKeyPrefix:            "ctf:auth:oauth",
			},
		},
		Redis: RedisConfig{
			Addr: "127.0.0.1:6379",
			Cluster: RedisClusterConfig{
				Addrs:          []string{"10.0.1.10:6379", "10.0.1.11:6379"},
				RouteByLatency: true,
				RouteRandomly:  true,
			},
		},
		CORS: CORSConfig{
			AllowOrigins:     []string{"https://academy.example.com"},
			AllowCredentials: true,
		},
		Container: ContainerConfig{
			DefaultCPUQuota:        1,
			DefaultMemory:          256 * 1024 * 1024,
			DefaultPidsLimit:       128,
			DefaultExposedPort:     8080,
			PortRangeStart:         30000,
			PortRangeEnd:           40000,
			DeletePollInterval:     time.Second,
			DeleteMaxConcurrent:    8,
			OrphanGracePeriod:      time.Minute,
			CleanupLockTTL:         time.Minute,
			StartupRecoveryLockTTL: startuprecovery.DefaultLockTTL,
			ProxyTicketTTL:         time.Minute,
			ProxyBodyPreviewSize:   1024,
			Scheduler: ContainerSchedulerConfig{
				LockTTL: time.Minute,
			},
			RuntimeNodeHealth: ContainerRuntimeNodeHealthConfig{
				Enabled:          true,
				PollInterval:     10 * time.Second,
				ProbeTimeout:     2 * time.Second,
				StaleAfter:       30 * time.Second,
				FailureThreshold: 3,
			},
			Network: ContainerNetworkConfig{
				SingleContainerSubnetBase: "10.11.0.0/16",
				SingleContainerSubnetMask: 29,
				TopologySubnetBase:        "10.10.0.0/16",
				TopologySubnetMask:        24,
			},
		},
		SharedStorage: SharedStorageConfig{
			Type: "shared_fs",
			SharedFS: SharedFSStorageConfig{
				Root: "storage/shared",
			},
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
		RuntimeAgent: RuntimeAgentConfig{
			DialTimeout:      5 * time.Second,
			KeepaliveTime:    30 * time.Second,
			KeepaliveTimeout: 10 * time.Second,
			Server: RuntimeAgentServerConfig{
				Host:             "0.0.0.0",
				KeepaliveMinTime: 30 * time.Second,
				ShutdownTimeout:  10 * time.Second,
			},
		},
	}
}

func runtimeParamValue(params map[string]string, key string) string {
	for candidate, value := range params {
		if strings.EqualFold(candidate, key) {
			return value
		}
	}
	return ""
}
