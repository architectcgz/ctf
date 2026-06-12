package config

import (
	"fmt"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	startuprecovery "ctf-platform/internal/module/instance/startuprecovery"
)

var containerStartupRecoveryMaxLockTTL = startuprecovery.MaxSafeLockTTL(
	startuprecovery.DefaultHeartbeatInterval,
	startuprecovery.DefaultLeaderRetry,
)

var runningInContainer = func() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}
	return strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_HOST")) != ""
}

func (c *Config) Validate() error {
	if c.CORS.AllowCredentials && len(c.CORS.AllowOrigins) == 0 {
		return fmt.Errorf("cors.allow_origins must not be empty when cors.allow_credentials is true")
	}
	if strings.TrimSpace(c.SharedStorage.Type) == "shared_fs" {
		if strings.TrimSpace(c.SharedStorage.SharedFS.Root) == "" {
			return fmt.Errorf("shared_storage.shared_fs.root must not be empty when shared_storage.type is shared_fs")
		}
	}
	for _, origin := range c.CORS.AllowOrigins {
		if strings.TrimSpace(origin) == "" {
			return fmt.Errorf("cors.allow_origins must not contain empty origin")
		}
	}
	if c.Container.DefaultCPUQuota <= 0 || c.Container.DefaultCPUQuota > 16 {
		return fmt.Errorf("container.default_cpu_quota must be between 0 and 16 cores")
	}
	if c.Container.DefaultMemory < 64*1024*1024 || c.Container.DefaultMemory > 16*1024*1024*1024 {
		return fmt.Errorf("container.default_memory must be between 64MB and 16GB")
	}
	if c.Container.DefaultPidsLimit <= 0 || c.Container.DefaultPidsLimit > 10000 {
		return fmt.Errorf("container.default_pids_limit must be between 1 and 10000")
	}
	if c.Container.DefaultExposedPort <= 0 || c.Container.DefaultExposedPort > 65535 {
		return fmt.Errorf("container.default_exposed_port must be between 1 and 65535")
	}
	if c.Container.PortRangeStart <= 0 || c.Container.PortRangeStart > 65535 ||
		c.Container.PortRangeEnd <= 0 || c.Container.PortRangeEnd > 65535 {
		return fmt.Errorf("container.port_range_start and container.port_range_end must be between 1 and 65535")
	}
	if c.Container.PortRangeStart >= c.Container.PortRangeEnd {
		return fmt.Errorf("container.port_range_start must be less than container.port_range_end")
	}
	if c.Container.OrphanGracePeriod <= 0 {
		return fmt.Errorf("container.orphan_grace_period must be greater than 0")
	}
	if c.Container.CleanupLockTTL <= 0 {
		return fmt.Errorf("container.cleanup_lock_ttl must be greater than 0")
	}
	if c.Container.StartupRecoveryLockTTL <= 0 {
		return fmt.Errorf("container.startup_recovery_lock_ttl must be greater than 0")
	}
	if c.Container.StartupRecoveryLockTTL > containerStartupRecoveryMaxLockTTL {
		return fmt.Errorf("container.startup_recovery_lock_ttl must be less than or equal to %s", containerStartupRecoveryMaxLockTTL)
	}
	if c.Container.ProxyTicketTTL <= 0 {
		return fmt.Errorf("container.proxy_ticket_ttl must be greater than 0")
	}
	if c.Container.ProxyBodyPreviewSize <= 0 {
		return fmt.Errorf("container.proxy_body_preview_size must be greater than 0")
	}
	if c.Container.DefenseSSHEnabled {
		if strings.TrimSpace(c.Container.DefenseSSHHost) == "" {
			return fmt.Errorf("container.defense_ssh_host must not be empty when container.defense_ssh_enabled is true")
		}
		if c.Container.DefenseSSHPort <= 0 || c.Container.DefenseSSHPort > 65535 {
			return fmt.Errorf("container.defense_ssh_port must be between 1 and 65535")
		}
		hostKeyPath := filepath.Clean(strings.TrimSpace(c.Container.DefenseSSHHostKeyPath))
		if hostKeyPath == "" || hostKeyPath == "." {
			return fmt.Errorf("container.defense_ssh_host_key_path must not be empty when container.defense_ssh_enabled is true")
		}
	}
	if c.Container.DefenseWorkbenchReadOnlyEnabled {
		root := path.Clean(strings.TrimSpace(c.Container.DefenseWorkbenchRoot))
		if root == "" || root == "." || root == "/" || !path.IsAbs(root) {
			return fmt.Errorf("container.defense_workbench_root must be an absolute non-root path when container.defense_workbench_readonly_enabled is true")
		}
	}
	if err := validateContainerNetworkConfig(c.Container.Network); err != nil {
		return err
	}
	if c.Container.Registry.Enabled {
		if strings.TrimSpace(c.Container.Registry.Server) == "" {
			return fmt.Errorf("container.registry.server must not be empty when container.registry.enabled is true")
		}
		hasIdentityToken := strings.TrimSpace(c.Container.Registry.IdentityToken) != ""
		hasBasicAuth := strings.TrimSpace(c.Container.Registry.Username) != "" && strings.TrimSpace(c.Container.Registry.Password) != ""
		if !hasIdentityToken && !hasBasicAuth {
			return fmt.Errorf("container.registry requires username/password or identity_token when enabled")
		}
	}
	if c.Container.Registry.BuildEnabled {
		if strings.TrimSpace(c.Container.Registry.Server) == "" {
			return fmt.Errorf("container.registry.server must not be empty when container.registry.build_enabled is true")
		}
		if c.Container.Registry.BuildTimeout <= 0 {
			return fmt.Errorf("container.registry.build_timeout must be greater than 0")
		}
		if c.Container.Registry.BuildConcurrency <= 0 {
			return fmt.Errorf("container.registry.build_concurrency must be greater than 0")
		}
	}
	if (c.Container.Registry.Enabled || c.Container.Registry.BuildEnabled) &&
		runningInContainer() &&
		isLocalRegistryServer(c.Container.Registry.Server) &&
		strings.TrimSpace(c.Container.Registry.AccessServer) == "" {
		return fmt.Errorf("container.registry.access_server must not be empty when container.registry.server points to localhost and the backend runs inside a container")
	}
	if c.Container.Scheduler.Enabled {
		if c.Container.Scheduler.PollInterval <= 0 {
			return fmt.Errorf("container.scheduler.poll_interval must be greater than 0")
		}
		if c.Container.Scheduler.DesiredReconcileInterval <= 0 {
			return fmt.Errorf("container.scheduler.desired_reconcile_interval must be greater than 0")
		}
		if c.Container.Scheduler.DesiredReconcileFailureInitialBackoff <= 0 {
			return fmt.Errorf("container.scheduler.desired_reconcile_failure_initial_backoff must be greater than 0")
		}
		if c.Container.Scheduler.DesiredReconcileFailureMaxBackoff < c.Container.Scheduler.DesiredReconcileFailureInitialBackoff {
			return fmt.Errorf("container.scheduler.desired_reconcile_failure_max_backoff must be greater than or equal to container.scheduler.desired_reconcile_failure_initial_backoff")
		}
		if c.Container.Scheduler.DesiredReconcileSuppressAfterFailures < 0 {
			return fmt.Errorf("container.scheduler.desired_reconcile_suppress_after_failures must be greater than or equal to 0")
		}
		if c.Container.Scheduler.DesiredReconcileSuppressAfterFailures > 0 && c.Container.Scheduler.DesiredReconcileSuppressDuration <= 0 {
			return fmt.Errorf("container.scheduler.desired_reconcile_suppress_duration must be greater than 0 when desired_reconcile_suppress_after_failures is enabled")
		}
		if c.Container.Scheduler.BatchSize <= 0 {
			return fmt.Errorf("container.scheduler.batch_size must be greater than 0")
		}
		if c.Container.Scheduler.MaxConcurrentStarts <= 0 {
			return fmt.Errorf("container.scheduler.max_concurrent_starts must be greater than 0")
		}
		if c.Container.Scheduler.MaxActiveInstances < 0 {
			return fmt.Errorf("container.scheduler.max_active_instances must be greater than or equal to 0")
		}
		if c.Container.Scheduler.LockTTL <= 0 {
			return fmt.Errorf("container.scheduler.lock_ttl must be greater than 0")
		}
	}
	if c.Container.DeletePollInterval <= 0 {
		return fmt.Errorf("container.delete_poll_interval must be greater than 0")
	}
	if c.Container.DeleteMaxConcurrent <= 0 {
		return fmt.Errorf("container.delete_max_concurrent must be greater than 0")
	}
	if c.Recommendation.WeakThreshold < 0 || c.Recommendation.WeakThreshold > 1 {
		return fmt.Errorf("recommendation.weak_threshold must be between 0 and 1")
	}
	if c.Recommendation.CacheTTL < time.Minute {
		return fmt.Errorf("recommendation.cache_ttl must be at least 1 minute")
	}
	if c.Recommendation.DefaultLimit <= 0 {
		return fmt.Errorf("recommendation.default_limit must be greater than 0")
	}
	if c.Recommendation.MaxLimit < c.Recommendation.DefaultLimit {
		return fmt.Errorf("recommendation.max_limit must be greater than or equal to default_limit")
	}
	if strings.TrimSpace(c.Report.StorageDir) == "" {
		return fmt.Errorf("report.storage_dir must not be empty")
	}
	if c.Report.DefaultFormat != "pdf" && c.Report.DefaultFormat != "excel" {
		return fmt.Errorf("report.default_format must be pdf or excel")
	}
	if c.Report.PersonalTimeout <= 0 {
		return fmt.Errorf("report.personal_timeout must be greater than 0")
	}
	if c.Report.ClassTimeout <= 0 {
		return fmt.Errorf("report.class_timeout must be greater than 0")
	}
	if c.Report.FileTTL <= 0 {
		return fmt.Errorf("report.file_ttl must be greater than 0")
	}
	if c.Report.MaxWorkers <= 0 {
		return fmt.Errorf("report.max_workers must be greater than 0")
	}
	if c.Dashboard.CacheTTL <= 0 {
		return fmt.Errorf("dashboard.cache_ttl must be greater than 0")
	}
	if c.Dashboard.AlertThreshold <= 0 || c.Dashboard.AlertThreshold > 100 {
		return fmt.Errorf("dashboard.alert_threshold must be between 0 and 100")
	}
	if c.WebSocket.TicketTTL <= 0 {
		return fmt.Errorf("websocket.ticket_ttl must be greater than 0")
	}
	if strings.TrimSpace(c.WebSocket.TicketKeyPrefix) == "" {
		return fmt.Errorf("websocket.ticket_key_prefix must not be empty")
	}
	if c.WebSocket.HeartbeatInterval <= 0 {
		return fmt.Errorf("websocket.heartbeat_interval must be greater than 0")
	}
	if c.WebSocket.ReadTimeout <= 0 {
		return fmt.Errorf("websocket.read_timeout must be greater than 0")
	}
	if c.WebSocket.ReadTimeout <= c.WebSocket.HeartbeatInterval {
		return fmt.Errorf("websocket.read_timeout must be greater than heartbeat_interval")
	}
	if c.WebSocket.RetryInitialDelay <= 0 {
		return fmt.Errorf("websocket.retry_initial_delay must be greater than 0")
	}
	if c.WebSocket.RetryMaxDelay < c.WebSocket.RetryInitialDelay {
		return fmt.Errorf("websocket.retry_max_delay must be greater than or equal to retry_initial_delay")
	}
	if c.RuntimeAgent.Enabled {
		if strings.TrimSpace(c.RuntimeAgent.Endpoint) == "" {
			return fmt.Errorf("runtime_agent.endpoint must not be empty when runtime_agent.enabled is true")
		}
		if c.RuntimeAgent.DialTimeout <= 0 {
			return fmt.Errorf("runtime_agent.dial_timeout must be greater than 0 when runtime_agent.enabled is true")
		}
		if strings.TrimSpace(c.RuntimeAgent.ServerName) == "" {
			return fmt.Errorf("runtime_agent.server_name must not be empty when runtime_agent.enabled is true")
		}
		if strings.TrimSpace(c.RuntimeAgent.CAFile) == "" {
			return fmt.Errorf("runtime_agent.ca_file must not be empty when runtime_agent.enabled is true")
		}
		if strings.TrimSpace(c.RuntimeAgent.CertFile) == "" {
			return fmt.Errorf("runtime_agent.cert_file must not be empty when runtime_agent.enabled is true")
		}
		if strings.TrimSpace(c.RuntimeAgent.KeyFile) == "" {
			return fmt.Errorf("runtime_agent.key_file must not be empty when runtime_agent.enabled is true")
		}
	}
	if c.RuntimeAgent.Server.Enabled {
		if strings.TrimSpace(c.RuntimeAgent.Server.Host) == "" {
			return fmt.Errorf("runtime_agent.server.host must not be empty when runtime_agent.server.enabled is true")
		}
		if c.RuntimeAgent.Server.Port <= 0 || c.RuntimeAgent.Server.Port > 65535 {
			return fmt.Errorf("runtime_agent.server.port must be between 1 and 65535 when runtime_agent.server.enabled is true")
		}
		if strings.TrimSpace(c.RuntimeAgent.Server.CertFile) == "" {
			return fmt.Errorf("runtime_agent.server.cert_file must not be empty when runtime_agent.server.enabled is true")
		}
		if strings.TrimSpace(c.RuntimeAgent.Server.KeyFile) == "" {
			return fmt.Errorf("runtime_agent.server.key_file must not be empty when runtime_agent.server.enabled is true")
		}
		if strings.TrimSpace(c.RuntimeAgent.Server.ClientCAFile) == "" {
			return fmt.Errorf("runtime_agent.server.client_ca_file must not be empty when runtime_agent.server.enabled is true")
		}
		if c.RuntimeAgent.Server.ShutdownTimeout <= 0 {
			return fmt.Errorf("runtime_agent.server.shutdown_timeout must be greater than 0 when runtime_agent.server.enabled is true")
		}
	}
	if isProductionEnv(c.App.Env) {
		if isPlaceholderSecret(c.Postgres.Password) {
			return fmt.Errorf("postgres.password must be provided from a non-placeholder secret in prod")
		}
		if isPlaceholderSecret(c.Redis.Password) {
			return fmt.Errorf("redis.password must be provided from a non-placeholder secret in prod")
		}
	}
	redisMode := normalizedRedisMode(c.Redis.Mode)
	switch redisMode {
	case "single":
		if strings.TrimSpace(c.Redis.Addr) == "" {
			return fmt.Errorf("redis.addr must not be empty when redis.mode is single")
		}
	case "sentinel":
		if strings.TrimSpace(c.Redis.MasterName) == "" {
			return fmt.Errorf("redis.master_name must not be empty when redis.mode is sentinel")
		}
		if len(nonEmptyStrings(c.Redis.SentinelAddrs)) == 0 {
			return fmt.Errorf("redis.sentinel_addrs must not be empty when redis.mode is sentinel")
		}
	default:
		return fmt.Errorf("redis.mode must be one of single or sentinel, got %q", strings.TrimSpace(c.Redis.Mode))
	}
	if c.Contest.StatusUpdateInterval <= 0 {
		return fmt.Errorf("contest.status_update_interval must be greater than 0")
	}
	if c.Contest.StatusUpdateBatchSize <= 0 {
		return fmt.Errorf("contest.status_update_batch_size must be greater than 0")
	}
	if c.Contest.StatusUpdateLockTTL <= 0 {
		return fmt.Errorf("contest.status_update_lock_ttl must be greater than 0")
	}
	if c.Contest.SubmissionRateLimitTTL <= 0 {
		return fmt.Errorf("contest.submission_rate_limit_ttl must be greater than 0")
	}
	if c.Contest.AWD.SchedulerInterval <= 0 {
		return fmt.Errorf("contest.awd.scheduler_interval must be greater than 0")
	}
	if c.Contest.AWD.SchedulerLockTTL <= 0 {
		return fmt.Errorf("contest.awd.scheduler_lock_ttl must be greater than 0")
	}
	if c.Contest.AWD.SchedulerBatchSize <= 0 {
		return fmt.Errorf("contest.awd.scheduler_batch_size must be greater than 0")
	}
	if c.Contest.AWD.RoundInterval <= 0 {
		return fmt.Errorf("contest.awd.round_interval must be greater than 0")
	}
	if c.Contest.AWD.RoundLockTTL <= 0 {
		return fmt.Errorf("contest.awd.round_lock_ttl must be greater than 0")
	}
	if c.Contest.AWD.PreviousRoundGrace < 0 {
		return fmt.Errorf("contest.awd.previous_round_grace must be greater than or equal to 0")
	}
	if c.Contest.AWD.CheckerTimeout <= 0 {
		return fmt.Errorf("contest.awd.checker_timeout must be greater than 0")
	}
	if strings.TrimSpace(c.Contest.AWD.CheckerSandbox.Image) == "" {
		return fmt.Errorf("contest.awd.checker_sandbox.image must not be empty")
	}
	if strings.TrimSpace(c.Contest.AWD.CheckerSandbox.WorkDir) == "" {
		return fmt.Errorf("contest.awd.checker_sandbox.work_dir must not be empty")
	}
	if !strings.HasPrefix(strings.TrimSpace(c.Contest.AWD.CheckerSandbox.WorkDir), "/") || strings.TrimSpace(c.Contest.AWD.CheckerSandbox.WorkDir) == "/" {
		return fmt.Errorf("contest.awd.checker_sandbox.work_dir must be an absolute non-root path")
	}
	if c.Contest.AWD.CheckerSandbox.Timeout <= 0 {
		return fmt.Errorf("contest.awd.checker_sandbox.timeout must be greater than 0")
	}
	if c.Contest.AWD.CheckerSandbox.CPUQuota <= 0 || c.Contest.AWD.CheckerSandbox.CPUQuota > 4 {
		return fmt.Errorf("contest.awd.checker_sandbox.cpu_quota must be between 0 and 4 cores")
	}
	if c.Contest.AWD.CheckerSandbox.MemoryBytes < 32*1024*1024 || c.Contest.AWD.CheckerSandbox.MemoryBytes > 2*1024*1024*1024 {
		return fmt.Errorf("contest.awd.checker_sandbox.memory_bytes must be between 32MB and 2GB")
	}
	if c.Contest.AWD.CheckerSandbox.PidsLimit <= 0 || c.Contest.AWD.CheckerSandbox.PidsLimit > 1024 {
		return fmt.Errorf("contest.awd.checker_sandbox.pids_limit must be between 1 and 1024")
	}
	if c.Contest.AWD.CheckerSandbox.NofileLimit <= 0 || c.Contest.AWD.CheckerSandbox.NofileLimit > 4096 {
		return fmt.Errorf("contest.awd.checker_sandbox.nofile_limit must be between 1 and 4096")
	}
	if c.Contest.AWD.CheckerSandbox.OutputLimitBytes <= 0 || c.Contest.AWD.CheckerSandbox.OutputLimitBytes > 1024*1024 {
		return fmt.Errorf("contest.awd.checker_sandbox.output_limit_bytes must be between 1 and 1MB")
	}
	if c.Auth.CAS.Enabled {
		if strings.TrimSpace(c.Auth.CAS.BaseURL) == "" {
			return fmt.Errorf("auth.cas.base_url must not be empty when CAS is enabled")
		}
		if strings.TrimSpace(c.Auth.CAS.ServiceURL) == "" {
			return fmt.Errorf("auth.cas.service_url must not be empty when CAS is enabled")
		}
	}
	return nil
}

func isLocalRegistryServer(server string) bool {
	normalized := strings.TrimSpace(server)
	normalized = strings.TrimPrefix(normalized, "https://")
	normalized = strings.TrimPrefix(normalized, "http://")
	normalized = strings.Trim(normalized, "/")
	if slash := strings.Index(normalized, "/"); slash >= 0 {
		normalized = normalized[:slash]
	}
	host := normalized
	if colon := strings.Index(normalized, ":"); colon >= 0 {
		host = normalized[:colon]
	}
	return host == "127.0.0.1" || host == "localhost"
}

func isProductionEnv(env string) bool {
	switch strings.ToLower(strings.TrimSpace(env)) {
	case "prod", "production":
		return true
	default:
		return false
	}
}

func isPlaceholderSecret(value string) bool {
	normalized := strings.TrimSpace(value)
	return normalized == "" || normalized == "change_me"
}

func normalizedRedisMode(mode string) string {
	normalized := strings.ToLower(strings.TrimSpace(mode))
	if normalized == "" {
		return "single"
	}
	return normalized
}

func nonEmptyStrings(values []string) []string {
	filtered := make([]string, 0, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		filtered = append(filtered, trimmed)
	}
	return filtered
}

func validateContainerNetworkConfig(cfg ContainerNetworkConfig) error {
	singleContainerCIDR, err := validateSubnetPoolConfig(
		"container.network.single_container_subnet_base",
		"container.network.single_container_subnet_mask",
		cfg.SingleContainerSubnetBase,
		cfg.SingleContainerSubnetMask,
	)
	if err != nil {
		return err
	}

	topologyCIDR, err := validateSubnetPoolConfig(
		"container.network.topology_subnet_base",
		"container.network.topology_subnet_mask",
		cfg.TopologySubnetBase,
		cfg.TopologySubnetMask,
	)
	if err != nil {
		return err
	}

	if cidrOverlaps(singleContainerCIDR, topologyCIDR) {
		return fmt.Errorf("container.network.single_container_subnet_base and container.network.topology_subnet_base must not overlap")
	}
	return nil
}

func validateSubnetPoolConfig(baseFieldName, maskFieldName, base string, mask int) (*net.IPNet, error) {
	trimmedBase := strings.TrimSpace(base)
	if trimmedBase == "" {
		return nil, fmt.Errorf("%s must not be empty", baseFieldName)
	}

	ip, ipNet, err := net.ParseCIDR(trimmedBase)
	if err != nil {
		return nil, fmt.Errorf("%s must be a valid CIDR", baseFieldName)
	}

	ip4 := ip.To4()
	baseIP := ipNet.IP.To4()
	if ip4 == nil || baseIP == nil {
		return nil, fmt.Errorf("%s must be an IPv4 CIDR", baseFieldName)
	}
	if !ip4.Equal(baseIP) {
		return nil, fmt.Errorf("%s must use the network address of the CIDR", baseFieldName)
	}

	basePrefix, bits := ipNet.Mask.Size()
	if bits != 32 {
		return nil, fmt.Errorf("%s must be an IPv4 CIDR", baseFieldName)
	}
	if mask <= 0 || mask > 30 {
		return nil, fmt.Errorf("%s must be between 1 and 30", maskFieldName)
	}
	if mask <= basePrefix {
		return nil, fmt.Errorf("%s must be greater than the base CIDR prefix", maskFieldName)
	}

	return ipNet, nil
}

func cidrOverlaps(a, b *net.IPNet) bool {
	if a == nil || b == nil {
		return false
	}
	return a.Contains(b.IP) || b.Contains(a.IP)
}
