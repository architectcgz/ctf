package jobs

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"
)

type awdTCPCheckerConfig struct {
	TimeoutMS int                        `json:"timeout_ms"`
	Connect   awdTCPCheckerConnectConfig `json:"connect"`
	Steps     []awdTCPCheckerStepConfig  `json:"steps"`
	Havoc     []awdTCPCheckerStepConfig  `json:"havoc"`
}

type awdTCPCheckerConnectConfig struct {
	Host string `json:"host"`
	Port any    `json:"port"`
}

type awdTCPCheckerStepConfig struct {
	Send           string `json:"send"`
	SendTemplate   string `json:"send_template"`
	SendHex        string `json:"send_hex"`
	ExpectContains string `json:"expect_contains"`
	ExpectRegex    string `json:"expect_regex"`
	TimeoutMS      int    `json:"timeout_ms"`
}

func parseAWDTCPCheckerConfig(raw string) (awdTCPCheckerConfig, error) {
	var cfg awdTCPCheckerConfig
	if strings.TrimSpace(raw) == "" {
		return cfg, fmt.Errorf("tcp checker config is required")
	}
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return cfg, err
	}
	if cfg.TimeoutMS < 0 || cfg.TimeoutMS > 60000 {
		return cfg, fmt.Errorf("tcp checker timeout_ms must be between 0 and 60000")
	}
	for _, step := range append(append([]awdTCPCheckerStepConfig{}, cfg.Steps...), cfg.Havoc...) {
		if step.TimeoutMS < 0 || step.TimeoutMS > 60000 {
			return cfg, fmt.Errorf("tcp checker step timeout_ms must be between 0 and 60000")
		}
		if strings.TrimSpace(step.SendHex) != "" && (step.Send != "" || step.SendTemplate != "") {
			return cfg, fmt.Errorf("tcp checker step send_hex cannot be combined with send or send_template")
		}
		if step.Send != "" && step.SendTemplate != "" {
			return cfg, fmt.Errorf("tcp checker step send cannot be combined with send_template")
		}
		if strings.TrimSpace(step.ExpectRegex) != "" {
			if _, err := regexp.Compile(step.ExpectRegex); err != nil {
				return cfg, fmt.Errorf("tcp checker expect_regex invalid: %w", err)
			}
		}
	}
	return cfg, nil
}

func (c awdTCPCheckerConfig) timeout(defaultTimeout time.Duration) time.Duration {
	if c.TimeoutMS > 0 {
		return time.Duration(c.TimeoutMS) * time.Millisecond
	}
	return normalizedAWDCheckerTimeout(defaultTimeout)
}

func (s awdTCPCheckerStepConfig) timeout(defaultTimeout time.Duration) time.Duration {
	if s.TimeoutMS > 0 {
		return time.Duration(s.TimeoutMS) * time.Millisecond
	}
	return defaultTimeout
}
