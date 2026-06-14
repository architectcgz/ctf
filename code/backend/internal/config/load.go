package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

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
	cfg.Container.FlagGlobalSecretFile = strings.TrimSpace(cfg.Container.FlagGlobalSecretFile)
	if cfg.Container.FlagGlobalSecret != "" && len(cfg.Container.FlagGlobalSecret) < 32 {
		return nil, fmt.Errorf("container.flag_global_secret must be at least 32 bytes, current length: %d", len(cfg.Container.FlagGlobalSecret))
	}
	resolvedFlagSecret, err := resolveContainerFlagGlobalSecret(cfg.Container.FlagGlobalSecret, cfg.Container.FlagGlobalSecretFile, !isProductionEnv(cfg.App.Env))
	if err != nil {
		return nil, err
	}
	cfg.Container.FlagGlobalSecret = resolvedFlagSecret
	if len(cfg.Container.FlagGlobalSecret) < 32 {
		return nil, fmt.Errorf("container.flag_global_secret must be at least 32 bytes, current length: %d", len(cfg.Container.FlagGlobalSecret))
	}
	if err := cfg.resolveContainerFlagSecretKeyring(); err != nil {
		return nil, err
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	return cfg, nil
}
