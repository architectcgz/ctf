package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

func Load(env string) (*Config, error) {
	return load(env, true, "validate config", (*Config).Validate)
}

func LoadRuntimeAgent(env string) (*Config, error) {
	return load(env, false, "validate runtime-agent config", (*Config).ValidateRuntimeAgent)
}

func load(env string, resolveContainerSecrets bool, validateLabel string, validate func(*Config) error) (*Config, error) {
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

	normalizeLoadedConfig(cfg, env)
	if resolveContainerSecrets {
		if err := cfg.resolveContainerFlagSecretForLoad(); err != nil {
			return nil, err
		}
	}
	if validate != nil {
		if err := validate(cfg); err != nil {
			return nil, fmt.Errorf("%s: %w", validateLabel, err)
		}
	}

	return cfg, nil
}

func normalizeLoadedConfig(cfg *Config, env string) {
	if cfg.App.Env == "" {
		cfg.App.Env = env
	}
	cfg.SharedStorage.Type = strings.TrimSpace(cfg.SharedStorage.Type)
	cfg.SharedStorage.SharedFS.Root = strings.TrimSpace(cfg.SharedStorage.SharedFS.Root)
	cfg.Auth.OAuth.IssuerURL = strings.TrimSpace(cfg.Auth.OAuth.IssuerURL)
	cfg.Auth.OAuth.RedisKeyPrefix = strings.TrimSpace(cfg.Auth.OAuth.RedisKeyPrefix)
	for i, prefix := range cfg.Auth.OAuth.AllowedRedirectURIPrefixes {
		cfg.Auth.OAuth.AllowedRedirectURIPrefixes[i] = strings.TrimSpace(prefix)
	}

	cfg.Container.FlagGlobalSecret = strings.TrimSpace(cfg.Container.FlagGlobalSecret)
	cfg.Container.FlagGlobalSecretFile = strings.TrimSpace(cfg.Container.FlagGlobalSecretFile)
	if cfg.Container.FlagGlobalSecretFile != "" {
		cfg.Container.FlagGlobalSecretFile = cfg.SharedStoragePath(cfg.Container.FlagGlobalSecretFile)
	}
	cfg.Container.DefenseSSHHostKeyPath = strings.TrimSpace(cfg.Container.DefenseSSHHostKeyPath)
	if cfg.Container.DefenseSSHHostKeyPath != "" {
		cfg.Container.DefenseSSHHostKeyPath = cfg.SharedStoragePath(cfg.Container.DefenseSSHHostKeyPath)
	}
	cfg.RuntimeAgent.NodeName = strings.TrimSpace(cfg.RuntimeAgent.NodeName)
	cfg.RuntimeAgent.Server.NodeName = strings.TrimSpace(cfg.RuntimeAgent.Server.NodeName)
}

func (c *Config) resolveContainerFlagSecretForLoad() error {
	if c == nil {
		return fmt.Errorf("config is nil")
	}
	if c.Container.FlagGlobalSecret != "" && len(c.Container.FlagGlobalSecret) < 32 {
		return fmt.Errorf("container.flag_global_secret must be at least 32 bytes, current length: %d", len(c.Container.FlagGlobalSecret))
	}
	resolvedFlagSecret, err := resolveContainerFlagGlobalSecret(c.Container.FlagGlobalSecret, c.Container.FlagGlobalSecretFile, !isProductionEnv(c.App.Env))
	if err != nil {
		return err
	}
	c.Container.FlagGlobalSecret = resolvedFlagSecret
	if len(c.Container.FlagGlobalSecret) < 32 {
		return fmt.Errorf("container.flag_global_secret must be at least 32 bytes, current length: %d", len(c.Container.FlagGlobalSecret))
	}
	if err := c.resolveContainerFlagSecretKeyring(); err != nil {
		return err
	}
	return nil
}
