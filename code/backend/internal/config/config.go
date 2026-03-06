package config

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App        AppConfig        `mapstructure:"app"`
	HTTP       HTTPConfig       `mapstructure:"http"`
	Log        LogConfig        `mapstructure:"log"`
	Postgres   PostgresConfig   `mapstructure:"postgres"`
	Redis      RedisConfig      `mapstructure:"redis"`
	CORS       CORSConfig       `mapstructure:"cors"`
	Auth       AuthConfig       `mapstructure:"auth"`
	RateLimit  RateLimitConfig  `mapstructure:"rate_limit"`
	Container  ContainerConfig  `mapstructure:"container"`
	Pagination PaginationConfig `mapstructure:"pagination"`
	Contest    ContestConfig    `mapstructure:"contest"`
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
}

type RateLimitConfig struct {
	RedisKeyPrefix string                `mapstructure:"redis_key_prefix"`
	Global         RateLimitPolicyConfig `mapstructure:"global"`
	Login          RateLimitPolicyConfig `mapstructure:"login"`
	FlagSubmit     RateLimitPolicyConfig `mapstructure:"flag_submit"`
}

type RateLimitPolicyConfig struct {
	Enabled bool          `mapstructure:"enabled"`
	Limit   int           `mapstructure:"limit"`
	Window  time.Duration `mapstructure:"window"`
}

type ContainerConfig struct {
	DefaultCPUQuota         int64         `mapstructure:"default_cpu_quota"`
	DefaultMemory           int64         `mapstructure:"default_memory"`
	DefaultPidsLimit        int64         `mapstructure:"default_pids_limit"`
	ReadonlyRootfs          bool          `mapstructure:"readonly_rootfs"`
	RunAsUser               string        `mapstructure:"run_as_user"`
	PortRangeStart          int           `mapstructure:"port_range_start"`
	PortRangeEnd            int           `mapstructure:"port_range_end"`
	MaxConcurrentPerUser    int           `mapstructure:"max_concurrent_per_user"`
	DefaultTTL              time.Duration `mapstructure:"default_ttl"`
	MaxExtends              int           `mapstructure:"max_extends"`
	ExtendDuration          time.Duration `mapstructure:"extend_duration"`
	CleanupInterval         string        `mapstructure:"cleanup_interval"`
	FlagGlobalSecret        string        `mapstructure:"flag_global_secret"`
}

type PaginationConfig struct {
	DefaultPageSize int `mapstructure:"default_page_size"`
	MaxPageSize     int `mapstructure:"max_page_size"`
}

type ContestConfig struct {
	StatusUpdateInterval  time.Duration `mapstructure:"status_update_interval"`
	StatusUpdateBatchSize int           `mapstructure:"status_update_batch_size"`
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

	return cfg, nil
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
	v.SetDefault("rate_limit.redis_key_prefix", "ctf:ratelimit")
	v.SetDefault("rate_limit.global.enabled", true)
	v.SetDefault("rate_limit.global.limit", 120)
	v.SetDefault("rate_limit.global.window", time.Minute)
	v.SetDefault("rate_limit.login.enabled", true)
	v.SetDefault("rate_limit.login.limit", 10)
	v.SetDefault("rate_limit.login.window", time.Minute)
	v.SetDefault("rate_limit.flag_submit.enabled", true)
	v.SetDefault("rate_limit.flag_submit.limit", 5)
	v.SetDefault("rate_limit.flag_submit.window", time.Minute)
	v.SetDefault("container.default_cpu_quota", 50000)
	v.SetDefault("container.default_memory", 268435456)
	v.SetDefault("container.default_pids_limit", 100)
	v.SetDefault("container.readonly_rootfs", false)
	v.SetDefault("container.run_as_user", "")
	v.SetDefault("container.port_range_start", 30000)
	v.SetDefault("container.port_range_end", 40000)
	v.SetDefault("container.max_concurrent_per_user", 3)
	v.SetDefault("container.default_ttl", 2*time.Hour)
	v.SetDefault("container.max_extends", 2)
	v.SetDefault("container.extend_duration", 1*time.Hour)
	v.SetDefault("container.cleanup_interval", "*/5 * * * *")
	v.SetDefault("pagination.default_page_size", 20)
	v.SetDefault("pagination.max_page_size", 100)
	v.SetDefault("contest.status_update_interval", 1*time.Minute)
	v.SetDefault("contest.status_update_batch_size", 1000)
}
