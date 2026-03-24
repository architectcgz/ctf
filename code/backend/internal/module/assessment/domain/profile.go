package domain

import (
	"time"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

type DimensionScore struct {
	Dimension  string
	TotalScore int
	UserScore  int
}

func BuildEmptyProfile(userID int64) *dto.SkillProfileResp {
	dimensions := make([]*dto.SkillDimension, 0, len(model.AllDimensions))
	for _, dim := range model.AllDimensions {
		dimensions = append(dimensions, &dto.SkillDimension{
			Dimension: dim,
			Score:     0,
		})
	}

	return &dto.SkillProfileResp{
		UserID:     userID,
		Dimensions: dimensions,
		UpdatedAt:  "",
	}
}

func NormalizeAssessmentConfig(cfg config.AssessmentConfig) config.AssessmentConfig {
	if cfg.RedisKeyPrefix == "" {
		cfg.RedisKeyPrefix = "ctf:assessment:skill-profile"
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
