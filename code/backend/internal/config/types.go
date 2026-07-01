package config

import (
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"
)

type Config struct {
	App            AppConfig            `mapstructure:"app"`
	HTTP           HTTPConfig           `mapstructure:"http"`
	Log            LogConfig            `mapstructure:"log"`
	Postgres       PostgresConfig       `mapstructure:"postgres"`
	Redis          RedisConfig          `mapstructure:"redis"`
	SharedStorage  SharedStorageConfig  `mapstructure:"shared_storage"`
	CORS           CORSConfig           `mapstructure:"cors"`
	Auth           AuthConfig           `mapstructure:"auth"`
	RateLimit      RateLimitConfig      `mapstructure:"rate_limit"`
	Container      ContainerConfig      `mapstructure:"container"`
	Pagination     PaginationConfig     `mapstructure:"pagination"`
	Challenge      ChallengeConfig      `mapstructure:"challenge"`
	Score          ScoreConfig          `mapstructure:"score"`
	Cache          CacheConfig          `mapstructure:"cache"`
	Assessment     AssessmentConfig     `mapstructure:"assessment"`
	Recommendation RecommendationConfig `mapstructure:"recommendation"`
	Report         ReportConfig         `mapstructure:"report"`
	Dashboard      DashboardConfig      `mapstructure:"dashboard"`
	WebSocket      WebSocketConfig      `mapstructure:"websocket"`
	Contest        ContestConfig        `mapstructure:"contest"`
	RuntimeAgent   RuntimeAgentConfig   `mapstructure:"runtime_agent"`
}

type SharedStorageConfig struct {
	Type     string                `mapstructure:"type"`
	SharedFS SharedFSStorageConfig `mapstructure:"shared_fs"`
}

type SharedFSStorageConfig struct {
	Root string `mapstructure:"root"`
}

type AppConfig struct {
	Name    string `mapstructure:"name"`
	Env     string `mapstructure:"env"`
	Version string `mapstructure:"version"`
}

type HTTPConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type LogConfig struct {
	Level            string   `mapstructure:"level"`
	Format           string   `mapstructure:"format"`
	OutputPaths      []string `mapstructure:"output_paths"`
	ErrorOutputPaths []string `mapstructure:"error_output_paths"`
}

type PostgresConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Database        string        `mapstructure:"database"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

type RedisConfig struct {
	Mode             string             `mapstructure:"mode"`
	Addr             string             `mapstructure:"addr"`
	Password         string             `mapstructure:"password"`
	DB               int                `mapstructure:"db"`
	Cluster          RedisClusterConfig `mapstructure:"cluster"`
	MasterName       string             `mapstructure:"master_name"`
	SentinelAddrs    []string           `mapstructure:"sentinel_addrs"`
	SentinelUsername string             `mapstructure:"sentinel_username"`
	SentinelPassword string             `mapstructure:"sentinel_password"`
	DialTimeout      time.Duration      `mapstructure:"dial_timeout"`
	ReadTimeout      time.Duration      `mapstructure:"read_timeout"`
	WriteTimeout     time.Duration      `mapstructure:"write_timeout"`
}

type RedisClusterConfig struct {
	Addrs          []string `mapstructure:"addrs"`
	RouteByLatency bool     `mapstructure:"route_by_latency"`
	RouteRandomly  bool     `mapstructure:"route_randomly"`
}

type CORSConfig struct {
	AllowOrigins     []string      `mapstructure:"allow_origins"`
	AllowMethods     []string      `mapstructure:"allow_methods"`
	AllowHeaders     []string      `mapstructure:"allow_headers"`
	ExposeHeaders    []string      `mapstructure:"expose_headers"`
	AllowCredentials bool          `mapstructure:"allow_credentials"`
	MaxAge           time.Duration `mapstructure:"max_age"`
}

type AuthConfig struct {
	SessionTTL            time.Duration   `mapstructure:"session_ttl"`
	SessionCookieName     string          `mapstructure:"session_cookie_name"`
	SessionCookiePath     string          `mapstructure:"session_cookie_path"`
	SessionCookieSecure   bool            `mapstructure:"session_cookie_secure"`
	SessionCookieHTTPOnly bool            `mapstructure:"session_cookie_http_only"`
	SessionCookieSameSite string          `mapstructure:"session_cookie_same_site"`
	SessionKeyPrefix      string          `mapstructure:"session_key_prefix"`
	OAuth                 AuthOAuthConfig `mapstructure:"oauth"`
	CAS                   CASConfig       `mapstructure:"cas"`
}

type AuthOAuthConfig struct {
	IssuerURL                  string        `mapstructure:"issuer_url"`
	LoginURL                   string        `mapstructure:"login_url"`
	AuthorizationCodeTTL       time.Duration `mapstructure:"authorization_code_ttl"`
	AccessTokenTTL             time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL            time.Duration `mapstructure:"refresh_token_ttl"`
	ClientRegistrationEnabled  bool          `mapstructure:"client_registration_enabled"`
	AllowedRedirectURIPrefixes []string      `mapstructure:"allowed_redirect_uri_prefixes"`
	RedisKeyPrefix             string        `mapstructure:"redis_key_prefix"`
}

type CASConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	BaseURL       string `mapstructure:"base_url"`
	LoginPath     string `mapstructure:"login_path"`
	ValidatePath  string `mapstructure:"validate_path"`
	ServiceURL    string `mapstructure:"service_url"`
	AutoProvision bool   `mapstructure:"auto_provision"`
}

type RateLimitConfig struct {
	RedisKeyPrefix string                `mapstructure:"redis_key_prefix"`
	Anonymous      RateLimitPolicyConfig `mapstructure:"anonymous"`
	Global         RateLimitPolicyConfig `mapstructure:"global"`
	Login          RateLimitPolicyConfig `mapstructure:"login"`
	LoginIP        RateLimitPolicyConfig `mapstructure:"login_ip"`
	FlagSubmit     RateLimitPolicyConfig `mapstructure:"flag_submit"`
	MCP            RateLimitPolicyConfig `mapstructure:"mcp"`
}

type RateLimitPolicyConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	Limit        int           `mapstructure:"limit"`
	Window       time.Duration `mapstructure:"window"`
	LockDuration time.Duration `mapstructure:"lock_duration"`
}

type ContainerConfig struct {
	DefaultCPUQuota                 float64                          `mapstructure:"default_cpu_quota"`
	DefaultMemory                   int64                            `mapstructure:"default_memory"`
	DefaultPidsLimit                int64                            `mapstructure:"default_pids_limit"`
	ReadonlyRootfs                  bool                             `mapstructure:"readonly_rootfs"`
	RunAsUser                       string                           `mapstructure:"run_as_user"`
	AllowedCapabilities             []string                         `mapstructure:"allowed_capabilities"`
	Seccomp                         string                           `mapstructure:"seccomp"`
	PortRangeStart                  int                              `mapstructure:"port_range_start"`
	PortRangeEnd                    int                              `mapstructure:"port_range_end"`
	DefaultExposedPort              int                              `mapstructure:"default_exposed_port"`
	MaxConcurrentPerUser            int                              `mapstructure:"max_concurrent_per_user"`
	DefaultTTL                      time.Duration                    `mapstructure:"default_ttl"`
	SolveGracePeriod                time.Duration                    `mapstructure:"solve_grace_period"`
	MaxExtends                      int                              `mapstructure:"max_extends"`
	ExtendDuration                  time.Duration                    `mapstructure:"extend_duration"`
	CleanupInterval                 string                           `mapstructure:"cleanup_interval"`
	CleanupLockTTL                  time.Duration                    `mapstructure:"cleanup_lock_ttl"`
	StartupRecoveryLockTTL          time.Duration                    `mapstructure:"startup_recovery_lock_ttl"`
	DeletePollInterval              time.Duration                    `mapstructure:"delete_poll_interval"`
	DeleteMaxConcurrent             int                              `mapstructure:"delete_max_concurrent"`
	OrphanGracePeriod               time.Duration                    `mapstructure:"orphan_grace_period"`
	CreateTimeout                   time.Duration                    `mapstructure:"create_timeout"`
	StartProbeTimeout               time.Duration                    `mapstructure:"start_probe_timeout"`
	StartProbeInterval              time.Duration                    `mapstructure:"start_probe_interval"`
	StartProbeAttempts              int                              `mapstructure:"start_probe_attempts"`
	FlagGlobalSecret                string                           `mapstructure:"flag_global_secret"`
	FlagGlobalSecretFile            string                           `mapstructure:"flag_global_secret_file"`
	FlagGlobalSecretKeyID           string                           `mapstructure:"flag_global_secret_key_id"`
	FlagGlobalSecretKeyring         []ContainerFlagSecretKeyConfig   `mapstructure:"flag_global_secret_keyring"`
	FlagGlobalSecretAllowRotation   bool                             `mapstructure:"flag_global_secret_allow_rotation"`
	ResolvedFlagSecretKeyID         string                           `mapstructure:"-"`
	ResolvedFlagSecrets             map[string]string                `mapstructure:"-"`
	PublicHost                      string                           `mapstructure:"public_host"`
	AccessHost                      string                           `mapstructure:"access_host"`
	ProxyTicketTTL                  time.Duration                    `mapstructure:"proxy_ticket_ttl"`
	ProxyBodyPreviewSize            int                              `mapstructure:"proxy_body_preview_size"`
	DefenseSSHEnabled               bool                             `mapstructure:"defense_ssh_enabled"`
	DefenseSSHHost                  string                           `mapstructure:"defense_ssh_host"`
	DefenseSSHPort                  int                              `mapstructure:"defense_ssh_port"`
	DefenseSSHHostKeyPath           string                           `mapstructure:"defense_ssh_host_key_path"`
	DefenseWorkbenchReadOnlyEnabled bool                             `mapstructure:"defense_workbench_readonly_enabled"`
	DefenseWorkbenchRoot            string                           `mapstructure:"defense_workbench_root"`
	Network                         ContainerNetworkConfig           `mapstructure:"network"`
	Registry                        ContainerRegistryConfig          `mapstructure:"registry"`
	Scheduler                       ContainerSchedulerConfig         `mapstructure:"scheduler"`
	RuntimeNodeHealth               ContainerRuntimeNodeHealthConfig `mapstructure:"runtime_node_health"`
}

type ContainerFlagSecretKeyConfig struct {
	KeyID  string `mapstructure:"key_id"`
	Secret string `mapstructure:"secret"`
}

type ContainerNetworkConfig struct {
	SingleContainerSubnetBase string `mapstructure:"single_container_subnet_base"`
	SingleContainerSubnetMask int    `mapstructure:"single_container_subnet_mask"`
	TopologySubnetBase        string `mapstructure:"topology_subnet_base"`
	TopologySubnetMask        int    `mapstructure:"topology_subnet_mask"`
}

type ContainerRegistryConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	Scheme           string        `mapstructure:"scheme"`
	Server           string        `mapstructure:"server"`
	AccessServer     string        `mapstructure:"access_server"`
	Username         string        `mapstructure:"username"`
	Password         string        `mapstructure:"password"`
	IdentityToken    string        `mapstructure:"identity_token"`
	BuildEnabled     bool          `mapstructure:"build_enabled"`
	BuildTimeout     time.Duration `mapstructure:"build_timeout"`
	BuildConcurrency int           `mapstructure:"build_concurrency"`
	DefaultTagPolicy string        `mapstructure:"default_tag_policy"`
}

type ContainerSchedulerConfig struct {
	Enabled                               bool          `mapstructure:"enabled"`
	PollInterval                          time.Duration `mapstructure:"poll_interval"`
	DesiredReconcileInterval              time.Duration `mapstructure:"desired_reconcile_interval"`
	DesiredReconcileFailureInitialBackoff time.Duration `mapstructure:"desired_reconcile_failure_initial_backoff"`
	DesiredReconcileFailureMaxBackoff     time.Duration `mapstructure:"desired_reconcile_failure_max_backoff"`
	DesiredReconcileSuppressAfterFailures int           `mapstructure:"desired_reconcile_suppress_after_failures"`
	DesiredReconcileSuppressDuration      time.Duration `mapstructure:"desired_reconcile_suppress_duration"`
	BatchSize                             int           `mapstructure:"batch_size"`
	MaxConcurrentStarts                   int           `mapstructure:"max_concurrent_starts"`
	MaxActiveInstances                    int           `mapstructure:"max_active_instances"`
	LockTTL                               time.Duration `mapstructure:"lock_ttl"`
}

type ContainerRuntimeNodeHealthConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	PollInterval     time.Duration `mapstructure:"poll_interval"`
	ProbeTimeout     time.Duration `mapstructure:"probe_timeout"`
	StaleAfter       time.Duration `mapstructure:"stale_after"`
	FailureThreshold int           `mapstructure:"failure_threshold"`
}

type PaginationConfig struct {
	DefaultPageSize int `mapstructure:"default_page_size"`
	MaxPageSize     int `mapstructure:"max_page_size"`
}

type CacheConfig struct {
	ProgressTTL time.Duration `mapstructure:"progress_ttl"`
}

type ScoreConfig struct {
	CacheTTL        time.Duration `mapstructure:"cache_ttl"`
	LockTimeout     time.Duration `mapstructure:"lock_timeout"`
	MaxRankingLimit int           `mapstructure:"max_ranking_limit"`
}

type ChallengeConfig struct {
	SolvedCountCacheTTL time.Duration               `mapstructure:"solved_count_cache_ttl"`
	PublishCheck        ChallengePublishCheckConfig `mapstructure:"publish_check"`
}

type ChallengePublishCheckConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	PollInterval time.Duration `mapstructure:"poll_interval"`
	BatchSize    int           `mapstructure:"batch_size"`
}

type AssessmentConfig struct {
	RedisKeyPrefix           string        `mapstructure:"redis_key_prefix"`
	DimensionTotalCacheTTL   time.Duration `mapstructure:"dimension_total_cache_ttl"`
	FullRebuildCron          string        `mapstructure:"full_rebuild_cron"`
	FullRebuildTimeout       time.Duration `mapstructure:"full_rebuild_timeout"`
	LockTTL                  time.Duration `mapstructure:"lock_ttl"`
	IncrementalUpdateDelay   time.Duration `mapstructure:"incremental_update_delay"`
	IncrementalUpdateTimeout time.Duration `mapstructure:"incremental_update_timeout"`
}

type RecommendationConfig struct {
	WeakThreshold float64       `mapstructure:"weak_threshold"`
	CacheTTL      time.Duration `mapstructure:"cache_ttl"`
	DefaultLimit  int           `mapstructure:"default_limit"`
	MaxLimit      int           `mapstructure:"max_limit"`
}

type ReportConfig struct {
	StorageDir      string        `mapstructure:"storage_dir"`
	DefaultFormat   string        `mapstructure:"default_format"`
	PersonalTimeout time.Duration `mapstructure:"personal_timeout"`
	ClassTimeout    time.Duration `mapstructure:"class_timeout"`
	FileTTL         time.Duration `mapstructure:"file_ttl"`
	MaxWorkers      int           `mapstructure:"max_workers"`
}

type DashboardConfig struct {
	CacheTTL       time.Duration `mapstructure:"cache_ttl"`
	AlertThreshold float64       `mapstructure:"alert_threshold"`
	RedisKeyPrefix string        `mapstructure:"redis_key_prefix"`
}

type WebSocketConfig struct {
	TicketTTL         time.Duration `mapstructure:"ticket_ttl"`
	TicketKeyPrefix   string        `mapstructure:"ticket_key_prefix"`
	HeartbeatInterval time.Duration `mapstructure:"heartbeat_interval"`
	ReadTimeout       time.Duration `mapstructure:"read_timeout"`
	RetryInitialDelay time.Duration `mapstructure:"retry_initial_delay"`
	RetryMaxDelay     time.Duration `mapstructure:"retry_max_delay"`
}

type RuntimeAgentConfig struct {
	Enabled            bool                     `mapstructure:"enabled"`
	AllowLocalFallback bool                     `mapstructure:"allow_local_fallback"`
	NodeName           string                   `mapstructure:"node_name"`
	Endpoint           string                   `mapstructure:"endpoint"`
	DialTimeout        time.Duration            `mapstructure:"dial_timeout"`
	KeepaliveTime      time.Duration            `mapstructure:"keepalive_time"`
	KeepaliveTimeout   time.Duration            `mapstructure:"keepalive_timeout"`
	ServerName         string                   `mapstructure:"server_name"`
	CAFile             string                   `mapstructure:"ca_file"`
	CertFile           string                   `mapstructure:"cert_file"`
	KeyFile            string                   `mapstructure:"key_file"`
	Server             RuntimeAgentServerConfig `mapstructure:"server"`
}

type RuntimeAgentServerConfig struct {
	Enabled          bool          `mapstructure:"enabled"`
	NodeName         string        `mapstructure:"node_name"`
	Host             string        `mapstructure:"host"`
	Port             int           `mapstructure:"port"`
	CertFile         string        `mapstructure:"cert_file"`
	KeyFile          string        `mapstructure:"key_file"`
	ClientCAFile     string        `mapstructure:"client_ca_file"`
	KeepaliveMinTime time.Duration `mapstructure:"keepalive_min_time"`
	ShutdownTimeout  time.Duration `mapstructure:"shutdown_timeout"`
}

type ContestConfig struct {
	StatusUpdateInterval   time.Duration    `mapstructure:"status_update_interval"`
	StatusUpdateBatchSize  int              `mapstructure:"status_update_batch_size"`
	StatusUpdateLockTTL    time.Duration    `mapstructure:"status_update_lock_ttl"`
	SubmissionRateLimitTTL time.Duration    `mapstructure:"submission_rate_limit_ttl"`
	BaseScore              float64          `mapstructure:"base_score"`
	MinScore               float64          `mapstructure:"min_score"`
	Decay                  float64          `mapstructure:"decay"`
	FirstBloodBonus        float64          `mapstructure:"first_blood_bonus"`
	AWD                    ContestAWDConfig `mapstructure:"awd"`
}

type ContestAWDConfig struct {
	SchedulerInterval  time.Duration        `mapstructure:"scheduler_interval"`
	SchedulerLockTTL   time.Duration        `mapstructure:"scheduler_lock_ttl"`
	SchedulerBatchSize int                  `mapstructure:"scheduler_batch_size"`
	RoundInterval      time.Duration        `mapstructure:"round_interval"`
	RoundLockTTL       time.Duration        `mapstructure:"round_lock_ttl"`
	PreviousRoundGrace time.Duration        `mapstructure:"previous_round_grace"`
	CheckerTimeout     time.Duration        `mapstructure:"checker_timeout"`
	CheckerHealthPath  string               `mapstructure:"checker_health_path"`
	CheckerSandbox     CheckerSandboxConfig `mapstructure:"checker_sandbox"`
}

type CheckerSandboxConfig struct {
	Image            string        `mapstructure:"image"`
	User             string        `mapstructure:"user"`
	WorkDir          string        `mapstructure:"work_dir"`
	Timeout          time.Duration `mapstructure:"timeout"`
	CPUQuota         float64       `mapstructure:"cpu_quota"`
	MemoryBytes      int64         `mapstructure:"memory_bytes"`
	PidsLimit        int64         `mapstructure:"pids_limit"`
	NofileLimit      int64         `mapstructure:"nofile_limit"`
	OutputLimitBytes int64         `mapstructure:"output_limit_bytes"`
	NetworkMode      string        `mapstructure:"network_mode"`
}

func (c RedisConfig) RedisClusterAddrs() []string {
	return nonEmptyStrings(c.Cluster.Addrs)
}

func postgresConnStringValue(value string) string {
	if value == "" {
		return "''"
	}
	if !postgresValueNeedsQuoting(value) {
		return value
	}
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `'`, `\'`)
	return "'" + escaped + "'"
}

func postgresValueNeedsQuoting(value string) bool {
	for _, r := range value {
		if unicode.IsSpace(r) || r == '\'' || r == '\\' {
			return true
		}
	}
	return false
}

func (c PostgresConfig) DSN() string {
	params := []struct {
		key   string
		value string
	}{
		{key: "host", value: c.Host},
		{key: "port", value: strconv.Itoa(c.Port)},
		{key: "user", value: c.Username},
		{key: "password", value: c.Password},
		{key: "dbname", value: c.Database},
		{key: "sslmode", value: c.SSLMode},
		{key: "TimeZone", value: "UTC"},
	}

	parts := make([]string, 0, len(params))
	for _, param := range params {
		parts = append(parts, param.key+"="+postgresConnStringValue(param.value))
	}
	return strings.Join(parts, " ")
}

func (c AuthConfig) CookieSameSite() http.SameSite {
	switch strings.ToLower(strings.TrimSpace(c.SessionCookieSameSite)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
