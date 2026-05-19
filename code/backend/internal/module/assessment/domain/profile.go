package domain

import (
	"time"

	"ctf-platform/internal/config"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmententity "ctf-platform/internal/module/assessment/entity"
	"ctf-platform/internal/shared/taxonomy"
)

type DimensionScore struct {
	Dimension  string
	TotalScore int
	UserScore  int
}

func BuildEmptyProfileContract(userID int64) *assessmentcontracts.SkillProfile {
	return BuildSkillProfileContract(userID, nil)
}

func BuildSkillProfileContract(userID int64, profiles []*assessmententity.SkillProfile) *assessmentcontracts.SkillProfile {
	dimensionMap, latestUpdate := buildProfileSnapshot(profiles)
	dimensions := make([]*assessmentcontracts.SkillDimension, 0, len(taxonomy.AllDimensions))
	for _, dim := range taxonomy.AllDimensions {
		dimensions = append(dimensions, &assessmentcontracts.SkillDimension{
			Dimension: dim,
			Score:     dimensionMap[dim],
		})
	}

	resp := &assessmentcontracts.SkillProfile{
		UserID:     userID,
		Dimensions: dimensions,
		UpdatedAt:  "",
	}
	if !latestUpdate.IsZero() {
		resp.UpdatedAt = latestUpdate.Format(time.RFC3339)
	}
	return resp
}

func buildProfileSnapshot(profiles []*assessmententity.SkillProfile) (map[string]float64, time.Time) {
	dimensionMap := make(map[string]float64, len(profiles))
	var latestUpdate time.Time
	for _, profile := range profiles {
		if profile == nil {
			continue
		}
		dimensionMap[profile.Dimension] = profile.Score
		if profile.UpdatedAt.After(latestUpdate) {
			latestUpdate = profile.UpdatedAt
		}
	}
	return dimensionMap, latestUpdate
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
