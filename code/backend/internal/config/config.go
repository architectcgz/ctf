package config

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App            AppConfig            `mapstructure:"app"`
	HTTP           HTTPConfig           `mapstructure:"http"`
	Log            LogConfig            `mapstructure:"log"`
	Postgres       PostgresConfig       `mapstructure:"postgres"`
	Redis          RedisConfig          `mapstructure:"redis"`
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
	Addr         string        `mapstructure:"addr"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
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
	Issuer                string        `mapstructure:"issuer"`
	AccessTokenTTL        time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL       time.Duration `mapstructure:"refresh_token_ttl"`
	RefreshCookieName     string        `mapstructure:"refresh_cookie_name"`
	RefreshCookiePath     string        `mapstructure:"refresh_cookie_path"`
	RefreshCookieSecure   bool          `mapstructure:"refresh_cookie_secure"`
	RefreshCookieHTTPOnly bool          `mapstructure:"refresh_cookie_http_only"`
	RefreshCookieSameSite string        `mapstructure:"refresh_cookie_same_site"`
	PrivateKeyPath        string        `mapstructure:"private_key_path"`
	PublicKeyPath         string        `mapstructure:"public_key_path"`
	TokenBlacklistPrefix  string        `mapstructure:"token_blacklist_prefix"`
	CAS                   CASConfig     `mapstructure:"cas"`
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
}

type RateLimitPolicyConfig struct {
	Enabled      bool          `mapstructure:"enabled"`
	Limit        int           `mapstructure:"limit"`
	Window       time.Duration `mapstructure:"window"`
	LockDuration time.Duration `mapstructure:"lock_duration"`
}

type ContainerConfig struct {
	DefaultCPUQuota      float64                  `mapstructure:"default_cpu_quota"` // CPU 核心数，如 0.5 表示 0.5 核
	DefaultMemory        int64                    `mapstructure:"default_memory"`    // 内存限制（字节）
	DefaultPidsLimit     int64                    `mapstructure:"default_pids_limit"`
	ReadonlyRootfs       bool                     `mapstructure:"readonly_rootfs"`
	RunAsUser            string                   `mapstructure:"run_as_user"`
	AllowedCapabilities  []string                 `mapstructure:"allowed_capabilities"`
	Seccomp              string                   `mapstructure:"seccomp"`
	PortRangeStart       int                      `mapstructure:"port_range_start"`
	PortRangeEnd         int                      `mapstructure:"port_range_end"`
	DefaultExposedPort   int                      `mapstructure:"default_exposed_port"`
	MaxConcurrentPerUser int                      `mapstructure:"max_concurrent_per_user"`
	DefaultTTL           time.Duration            `mapstructure:"default_ttl"`
	SolveGracePeriod     time.Duration            `mapstructure:"solve_grace_period"`
	MaxExtends           int                      `mapstructure:"max_extends"`
	ExtendDuration       time.Duration            `mapstructure:"extend_duration"`
	CleanupInterval      string                   `mapstructure:"cleanup_interval"`
	CleanupLockTTL       time.Duration            `mapstructure:"cleanup_lock_ttl"`
	OrphanGracePeriod    time.Duration            `mapstructure:"orphan_grace_period"`
	CreateTimeout        time.Duration            `mapstructure:"create_timeout"`
	StartProbeTimeout    time.Duration            `mapstructure:"start_probe_timeout"`
	StartProbeInterval   time.Duration            `mapstructure:"start_probe_interval"`
	StartProbeAttempts   int                      `mapstructure:"start_probe_attempts"`
	FlagGlobalSecret     string                   `mapstructure:"flag_global_secret"`
	PublicHost           string                   `mapstructure:"public_host"`
	ProxyTicketTTL       time.Duration            `mapstructure:"proxy_ticket_ttl"`
	ProxyBodyPreviewSize int                      `mapstructure:"proxy_body_preview_size"`
	Scheduler            ContainerSchedulerConfig `mapstructure:"scheduler"`
}

type ContainerSchedulerConfig struct {
	Enabled             bool          `mapstructure:"enabled"`
	PollInterval        time.Duration `mapstructure:"poll_interval"`
	BatchSize           int           `mapstructure:"batch_size"`
	MaxConcurrentStarts int           `mapstructure:"max_concurrent_starts"`
	MaxActiveInstances  int           `mapstructure:"max_active_instances"`
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
	SchedulerInterval  time.Duration `mapstructure:"scheduler_interval"`
	SchedulerLockTTL   time.Duration `mapstructure:"scheduler_lock_ttl"`
	SchedulerBatchSize int           `mapstructure:"scheduler_batch_size"`
	RoundInterval      time.Duration `mapstructure:"round_interval"`
	RoundLockTTL       time.Duration `mapstructure:"round_lock_ttl"`
	PreviousRoundGrace time.Duration `mapstructure:"previous_round_grace"`
	CheckerTimeout     time.Duration `mapstructure:"checker_timeout"`
	CheckerHealthPath  string        `mapstructure:"checker_health_path"`
}

func Load(env string) (*Config, error) {
	if strings.TrimSpace(env) == "" {
		env = "dev"
	}

	cfg := &Config{}
	v := viper.New()
	v.SetConfigType("yaml")
	v.AddConfigPath("configs")
	v.SetEnvPrefix("CTF")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v)

	v.SetConfigName("config")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	v.SetConfigName(fmt.Sprintf("config.%s", env))
	if err := v.MergeInConfig(); err != nil {
		return nil, fmt.Errorf("merge env config: %w", err)
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if cfg.App.Env == "" {
		cfg.App.Env = env
	}

	cfg.Container.FlagGlobalSecret = strings.TrimSpace(cfg.Container.FlagGlobalSecret)
	if cfg.Container.FlagGlobalSecret == "" {
		return nil, fmt.Errorf("container.flag_global_secret must be set via CTF_CONTAINER_FLAG_GLOBAL_SECRET environment variable")
	}
	if len(cfg.Container.FlagGlobalSecret) < 32 {
		return nil, fmt.Errorf("container.flag_global_secret must be at least 32 bytes, current length: %d", len(cfg.Container.FlagGlobalSecret))
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.CORS.AllowCredentials && len(c.CORS.AllowOrigins) == 0 {
		return fmt.Errorf("cors.allow_origins must not be empty when cors.allow_credentials is true")
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
	if c.Container.ProxyTicketTTL <= 0 {
		return fmt.Errorf("container.proxy_ticket_ttl must be greater than 0")
	}
	if c.Container.ProxyBodyPreviewSize <= 0 {
		return fmt.Errorf("container.proxy_body_preview_size must be greater than 0")
	}
	if c.Container.Scheduler.Enabled {
		if c.Container.Scheduler.PollInterval <= 0 {
			return fmt.Errorf("container.scheduler.poll_interval must be greater than 0")
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

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host,
		c.Port,
		c.Username,
		c.Password,
		c.Database,
		c.SSLMode,
	)
}

func (c AuthConfig) CookieSameSite() http.SameSite {
	switch strings.ToLower(strings.TrimSpace(c.RefreshCookieSameSite)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "ctf-platform")
	v.SetDefault("app.env", "dev")
	v.SetDefault("app.version", "dev")
	v.SetDefault("log.level", "info")
	v.SetDefault("log.format", "json")
	v.SetDefault("log.output_paths", []string{"stdout"})
	v.SetDefault("log.error_output_paths", []string{"stderr"})
	v.SetDefault("postgres.conn_max_lifetime", 30*time.Minute)
	v.SetDefault("redis.dial_timeout", 5*time.Second)
	v.SetDefault("redis.read_timeout", 3*time.Second)
	v.SetDefault("redis.write_timeout", 3*time.Second)
	v.SetDefault("cors.allow_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	v.SetDefault("cors.allow_headers", []string{"Authorization", "Content-Type", "X-Request-ID"})
	v.SetDefault("cors.expose_headers", []string{"X-Request-ID", "X-RateLimit-Limit", "X-RateLimit-Remaining", "X-RateLimit-Reset", "Retry-After"})
	v.SetDefault("cors.max_age", 12*time.Hour)
	v.SetDefault("auth.issuer", "ctf-platform")
	v.SetDefault("auth.access_token_ttl", 15*time.Minute)
	v.SetDefault("auth.refresh_token_ttl", 7*24*time.Hour)
	v.SetDefault("auth.refresh_cookie_name", "ctf_refresh_token")
	v.SetDefault("auth.refresh_cookie_path", "/api/v1/auth")
	v.SetDefault("auth.refresh_cookie_http_only", true)
	v.SetDefault("auth.refresh_cookie_same_site", "lax")
	v.SetDefault("auth.token_blacklist_prefix", "ctf:auth:blacklist")
	v.SetDefault("auth.cas.enabled", false)
	v.SetDefault("auth.cas.login_path", "/login")
	v.SetDefault("auth.cas.validate_path", "/serviceValidate")
	v.SetDefault("auth.cas.auto_provision", false)
	v.SetDefault("rate_limit.redis_key_prefix", "ctf:ratelimit")
	v.SetDefault("rate_limit.anonymous.enabled", true)
	v.SetDefault("rate_limit.anonymous.limit", 300)
	v.SetDefault("rate_limit.anonymous.window", time.Minute)
	v.SetDefault("rate_limit.global.enabled", true)
	v.SetDefault("rate_limit.global.limit", 600)
	v.SetDefault("rate_limit.global.window", time.Minute)
	v.SetDefault("rate_limit.login.enabled", true)
	v.SetDefault("rate_limit.login.limit", 10)
	v.SetDefault("rate_limit.login.window", time.Minute)
	v.SetDefault("rate_limit.login.lock_duration", 15*time.Minute)
	v.SetDefault("rate_limit.login_ip.enabled", true)
	v.SetDefault("rate_limit.login_ip.limit", 300)
	v.SetDefault("rate_limit.login_ip.window", time.Minute)
	v.SetDefault("rate_limit.flag_submit.enabled", true)
	v.SetDefault("rate_limit.flag_submit.limit", 5)
	v.SetDefault("rate_limit.flag_submit.window", time.Minute)
	v.SetDefault("container.default_cpu_quota", 0.5)
	v.SetDefault("container.default_memory", 268435456)
	v.SetDefault("container.default_pids_limit", 100)
	v.SetDefault("container.readonly_rootfs", false)
	v.SetDefault("container.run_as_user", "")
	v.SetDefault("container.allowed_capabilities", []string{"CHOWN", "SETUID", "SETGID"})
	v.SetDefault("container.seccomp", "default")
	v.SetDefault("container.port_range_start", 30000)
	v.SetDefault("container.port_range_end", 40000)
	v.SetDefault("container.default_exposed_port", 8080)
	v.SetDefault("container.max_concurrent_per_user", 3)
	v.SetDefault("container.default_ttl", 2*time.Hour)
	v.SetDefault("container.max_extends", 2)
	v.SetDefault("container.extend_duration", 1*time.Hour)
	v.SetDefault("container.cleanup_interval", "*/5 * * * *")
	v.SetDefault("container.cleanup_lock_ttl", 2*time.Minute)
	v.SetDefault("container.orphan_grace_period", 5*time.Minute)
	v.SetDefault("container.create_timeout", 30*time.Second)
	v.SetDefault("container.start_probe_timeout", 800*time.Millisecond)
	v.SetDefault("container.start_probe_interval", 300*time.Millisecond)
	v.SetDefault("container.start_probe_attempts", 5)
	v.SetDefault("container.flag_global_secret", "")
	v.SetDefault("container.public_host", "localhost")
	v.SetDefault("container.proxy_ticket_ttl", 15*time.Minute)
	v.SetDefault("container.proxy_body_preview_size", 1024)
	v.SetDefault("container.scheduler.enabled", true)
	v.SetDefault("container.scheduler.poll_interval", time.Second)
	v.SetDefault("container.scheduler.batch_size", 4)
	v.SetDefault("container.scheduler.max_concurrent_starts", 4)
	v.SetDefault("container.scheduler.max_active_instances", 60)
	v.SetDefault("pagination.default_page_size", 20)
	v.SetDefault("pagination.max_page_size", 100)
	v.SetDefault("challenge.solved_count_cache_ttl", 5*time.Minute)
	v.SetDefault("challenge.publish_check.enabled", true)
	v.SetDefault("challenge.publish_check.poll_interval", 2*time.Second)
	v.SetDefault("challenge.publish_check.batch_size", 1)
	v.SetDefault("score.cache_ttl", 5*time.Minute)
	v.SetDefault("score.lock_timeout", 5*time.Second)
	v.SetDefault("score.max_ranking_limit", 100)
	v.SetDefault("cache.progress_ttl", 10*time.Minute)
	v.SetDefault("assessment.redis_key_prefix", "ctf:assessment:skill-profile")
	v.SetDefault("assessment.full_rebuild_cron", "0 0 * * *")
	v.SetDefault("assessment.full_rebuild_timeout", 30*time.Minute)
	v.SetDefault("assessment.lock_ttl", 10*time.Second)
	v.SetDefault("assessment.incremental_update_delay", 100*time.Millisecond)
	v.SetDefault("assessment.incremental_update_timeout", 5*time.Second)
	v.SetDefault("recommendation.weak_threshold", 0.4)
	v.SetDefault("recommendation.cache_ttl", time.Hour)
	v.SetDefault("recommendation.default_limit", 6)
	v.SetDefault("recommendation.max_limit", 20)
	v.SetDefault("report.storage_dir", "storage/exports")
	v.SetDefault("report.default_format", "pdf")
	v.SetDefault("report.personal_timeout", 30*time.Second)
	v.SetDefault("report.class_timeout", 2*time.Minute)
	v.SetDefault("report.file_ttl", 7*24*time.Hour)
	v.SetDefault("report.max_workers", 2)
	v.SetDefault("dashboard.cache_ttl", 30*time.Second)
	v.SetDefault("dashboard.alert_threshold", 80.0)
	v.SetDefault("dashboard.redis_key_prefix", "ctf:dashboard")
	v.SetDefault("websocket.ticket_ttl", 30*time.Second)
	v.SetDefault("websocket.ticket_key_prefix", "ctf:ws:ticket")
	v.SetDefault("websocket.heartbeat_interval", 30*time.Second)
	v.SetDefault("websocket.read_timeout", 75*time.Second)
	v.SetDefault("websocket.retry_initial_delay", time.Second)
	v.SetDefault("websocket.retry_max_delay", 30*time.Second)
	v.SetDefault("contest.status_update_interval", 1*time.Minute)
	v.SetDefault("contest.status_update_batch_size", 1000)
	v.SetDefault("contest.status_update_lock_ttl", 30*time.Second)
	v.SetDefault("contest.submission_rate_limit_ttl", 5*time.Second)
	v.SetDefault("contest.base_score", 1000.0)
	v.SetDefault("contest.min_score", 100.0)
	v.SetDefault("contest.decay", 0.9)
	v.SetDefault("contest.first_blood_bonus", 0.1)
	v.SetDefault("contest.awd.scheduler_interval", 30*time.Second)
	v.SetDefault("contest.awd.scheduler_lock_ttl", 30*time.Second)
	v.SetDefault("contest.awd.scheduler_batch_size", 200)
	v.SetDefault("contest.awd.round_interval", 5*time.Minute)
	v.SetDefault("contest.awd.round_lock_ttl", 30*time.Second)
	v.SetDefault("contest.awd.previous_round_grace", 30*time.Second)
	v.SetDefault("contest.awd.checker_timeout", 3*time.Second)
	v.SetDefault("contest.awd.checker_health_path", "/health")
}
