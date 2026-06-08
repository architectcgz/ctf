package config

import (
	"strings"
	"time"

	platformconfig "ctf-platform/internal/config"
	assessmententity "ctf-platform/internal/module/assessment/entity"
)

func NormalizeAssessmentConfig(cfg platformconfig.AssessmentConfig) platformconfig.AssessmentConfig {
	if cfg.RedisKeyPrefix == "" {
		cfg.RedisKeyPrefix = "ctf:assessment:skill-profile"
	}
	if cfg.DimensionTotalCacheTTL <= 0 {
		cfg.DimensionTotalCacheTTL = 5 * time.Minute
	}
	if cfg.LockTTL <= 0 {
		cfg.LockTTL = 10 * time.Second
	}
	if cfg.FullRebuildCron == "" {
		cfg.FullRebuildCron = "0 0 * * *"
	}
	if cfg.FullRebuildTimeout <= 0 {
		cfg.FullRebuildTimeout = 30 * time.Minute
	}
	if cfg.IncrementalUpdateDelay <= 0 {
		cfg.IncrementalUpdateDelay = 100 * time.Millisecond
	}
	if cfg.IncrementalUpdateTimeout <= 0 {
		cfg.IncrementalUpdateTimeout = 5 * time.Second
	}
	return cfg
}

func NormalizeRecommendationConfig(cfg platformconfig.RecommendationConfig) platformconfig.RecommendationConfig {
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

func NormalizeReportConfig(cfg platformconfig.ReportConfig) platformconfig.ReportConfig {
	if strings.TrimSpace(cfg.StorageDir) == "" {
		cfg.StorageDir = "storage/exports"
	}
	if cfg.DefaultFormat != assessmententity.ReportFormatPDF && cfg.DefaultFormat != assessmententity.ReportFormatExcel {
		cfg.DefaultFormat = assessmententity.ReportFormatPDF
	}
	if cfg.PersonalTimeout <= 0 {
		cfg.PersonalTimeout = 30 * time.Second
	}
	if cfg.ClassTimeout <= 0 {
		cfg.ClassTimeout = 2 * time.Minute
	}
	if cfg.FileTTL <= 0 {
		cfg.FileTTL = 7 * 24 * time.Hour
	}
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = 2
	}
	return cfg
}
