package config

import (
	"testing"
	"time"

	platformconfig "ctf-platform/internal/config"
	assessmententity "ctf-platform/internal/module/assessment/entity"
)

func TestNormalizeAssessmentConfigAppliesDefaults(t *testing.T) {
	cfg := NormalizeAssessmentConfig(platformconfig.AssessmentConfig{})

	if cfg.RedisKeyPrefix != "ctf:assessment:skill-profile" {
		t.Fatalf("RedisKeyPrefix = %q", cfg.RedisKeyPrefix)
	}
	if cfg.DimensionTotalCacheTTL != 5*time.Minute {
		t.Fatalf("DimensionTotalCacheTTL = %s", cfg.DimensionTotalCacheTTL)
	}
	if cfg.LockTTL != 10*time.Second {
		t.Fatalf("LockTTL = %s", cfg.LockTTL)
	}
	if cfg.FullRebuildCron != "0 0 * * *" {
		t.Fatalf("FullRebuildCron = %q", cfg.FullRebuildCron)
	}
	if cfg.FullRebuildTimeout != 30*time.Minute {
		t.Fatalf("FullRebuildTimeout = %s", cfg.FullRebuildTimeout)
	}
	if cfg.IncrementalUpdateDelay != 100*time.Millisecond {
		t.Fatalf("IncrementalUpdateDelay = %s", cfg.IncrementalUpdateDelay)
	}
	if cfg.IncrementalUpdateTimeout != 5*time.Second {
		t.Fatalf("IncrementalUpdateTimeout = %s", cfg.IncrementalUpdateTimeout)
	}
}

func TestNormalizeRecommendationConfigAppliesDefaults(t *testing.T) {
	cfg := NormalizeRecommendationConfig(platformconfig.RecommendationConfig{
		WeakThreshold: -1,
		CacheTTL:      time.Second,
	})

	if cfg.WeakThreshold != 0.4 {
		t.Fatalf("WeakThreshold = %v", cfg.WeakThreshold)
	}
	if cfg.CacheTTL != time.Hour {
		t.Fatalf("CacheTTL = %s", cfg.CacheTTL)
	}
	if cfg.DefaultLimit != 6 {
		t.Fatalf("DefaultLimit = %d", cfg.DefaultLimit)
	}
	if cfg.MaxLimit != 20 {
		t.Fatalf("MaxLimit = %d", cfg.MaxLimit)
	}
}

func TestNormalizeReportConfigAppliesDefaults(t *testing.T) {
	cfg := NormalizeReportConfig(platformconfig.ReportConfig{
		StorageDir:    " \t ",
		DefaultFormat: "docx",
	})

	if cfg.StorageDir != "storage/exports" {
		t.Fatalf("StorageDir = %q", cfg.StorageDir)
	}
	if cfg.DefaultFormat != assessmententity.ReportFormatPDF {
		t.Fatalf("DefaultFormat = %q", cfg.DefaultFormat)
	}
	if cfg.PersonalTimeout != 30*time.Second {
		t.Fatalf("PersonalTimeout = %s", cfg.PersonalTimeout)
	}
	if cfg.ClassTimeout != 2*time.Minute {
		t.Fatalf("ClassTimeout = %s", cfg.ClassTimeout)
	}
	if cfg.FileTTL != 7*24*time.Hour {
		t.Fatalf("FileTTL = %s", cfg.FileTTL)
	}
	if cfg.MaxWorkers != 2 {
		t.Fatalf("MaxWorkers = %d", cfg.MaxWorkers)
	}
}
