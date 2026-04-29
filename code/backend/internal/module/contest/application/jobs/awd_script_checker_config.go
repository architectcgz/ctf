package jobs

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

type awdScriptCheckerConfig struct {
	Runtime    string                         `json:"runtime"`
	Entry      string                         `json:"entry"`
	TimeoutSec int                            `json:"timeout_sec"`
	Args       []string                       `json:"args"`
	Env        map[string]string              `json:"env"`
	Output     string                         `json:"output"`
	Artifact   awdScriptCheckerArtifactConfig `json:"artifact"`
}

type awdScriptCheckerArtifactConfig struct {
	Entry       string `json:"entry"`
	StoragePath string `json:"storage_path"`
	SHA256      string `json:"sha256"`
	Size        int64  `json:"size"`
}

func parseAWDScriptCheckerConfig(raw string) (awdScriptCheckerConfig, error) {
	var cfg awdScriptCheckerConfig
	if strings.TrimSpace(raw) == "" {
		return cfg, fmt.Errorf("script checker config is required")
	}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return cfg, err
	}
	cfg.Runtime = strings.TrimSpace(cfg.Runtime)
	cfg.Entry = strings.TrimSpace(cfg.Entry)
	cfg.Output = strings.TrimSpace(cfg.Output)
	if cfg.Runtime == "" {
		cfg.Runtime = "python3"
	}
	if cfg.Output == "" {
		cfg.Output = "exit_code"
	}
	if !allowedAWDScriptCheckerRuntime(cfg.Runtime) {
		return cfg, fmt.Errorf("unsupported script checker runtime: %s", cfg.Runtime)
	}
	if err := validateAWDScriptCheckerEntry(cfg.Entry); err != nil {
		return cfg, err
	}
	if cfg.TimeoutSec < 0 || cfg.TimeoutSec > 60 {
		return cfg, fmt.Errorf("script checker timeout_sec must be between 0 and 60")
	}
	if cfg.Output != "exit_code" && cfg.Output != "json" {
		return cfg, fmt.Errorf("script checker output must be exit_code or json")
	}
	return cfg, nil
}

func (c awdScriptCheckerConfig) timeout(defaultTimeout time.Duration) time.Duration {
	if c.TimeoutSec > 0 {
		return time.Duration(c.TimeoutSec) * time.Second
	}
	if defaultTimeout > 0 {
		return defaultTimeout
	}
	return 10 * time.Second
}

func allowedAWDScriptCheckerRuntime(value string) bool {
	switch strings.TrimSpace(value) {
	case "python3":
		return true
	default:
		return false
	}
}

func validateAWDScriptCheckerEntry(value string) error {
	clean := filepath.Clean(strings.TrimSpace(value))
	if clean == "." || clean == "" || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return fmt.Errorf("script checker entry must be a package-relative path")
	}
	return nil
}
