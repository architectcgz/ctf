package domain

import (
	"time"

	"ctf-platform/internal/config"
)

func NormalizeRecommendationConfig(cfg config.RecommendationConfig) config.RecommendationConfig {
	if cfg.WeakThreshold < 0 || cfg.WeakThreshold > 1 {
		cfg.WeakThreshold = 0.4
	}
	if cfg.CacheTTL < time.Minute {
		cfg.CacheTTL = time.Hour
	}
	if cfg.DefaultLimit <= 0 {
		cfg.DefaultLimit = 6
	}
	if cfg.MaxLimit < cfg.DefaultLimit {
		cfg.MaxLimit = 20
	}
	return cfg
}
