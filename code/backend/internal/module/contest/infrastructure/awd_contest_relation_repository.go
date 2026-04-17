package infrastructure

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

type awdContestServiceRuntimeRow struct {
	ChallengeID             int64                           `gorm:"column:challenge_id"`
	FlagPrefix              string                          `gorm:"column:flag_prefix"`
	DisplayName             string                          `gorm:"column:display_name"`
	RuntimeConfig           string                          `gorm:"column:runtime_config"`
	ScoreConfig             string                          `gorm:"column:score_config"`
	LegacyCheckerType       model.AWDCheckerType            `gorm:"column:legacy_checker_type"`
	LegacyCheckerConfig     string                          `gorm:"column:legacy_checker_config"`
	LegacySLAScore          int                             `gorm:"column:legacy_sla_score"`
	LegacyDefenseScore      int                             `gorm:"column:legacy_defense_score"`
	LegacyValidationState   model.AWDCheckerValidationState `gorm:"column:legacy_validation_state"`
	LegacyLastPreviewAt     *time.Time                      `gorm:"column:legacy_last_preview_at"`
	LegacyLastPreviewResult string                          `gorm:"column:legacy_last_preview_result"`
	LegacyChallengeTitle    string                          `gorm:"column:legacy_challenge_title"`
}

func (r *AWDRepository) ListSchedulableAWDContests(ctx context.Context, now, recentCutoff time.Time, limit int) ([]model.Contest, error) {
	var contests []model.Contest
	query := r.dbWithContext(ctx).
		Where("mode = ?", model.ContestModeAWD).
		Where("status IN ?", []string{
			model.ContestStatusRegistration,
			model.ContestStatusRunning,
			model.ContestStatusFrozen,
			model.ContestStatusEnded,
		}).
		Where("start_time <= ?", now).
		Where("end_time > ?", recentCutoff).
		Order("start_time ASC, id ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	if err := query.Find(&contests).Error; err != nil {
		return nil, err
	}
	return contests, nil
}

func (r *AWDRepository) ListChallengesByContest(ctx context.Context, contestID int64) ([]model.Challenge, error) {
	var challenges []model.Challenge
	if err := r.dbWithContext(ctx).
		Table("challenges AS c").
		Select("c.*").
		Joins("JOIN contest_challenges AS cc ON cc.challenge_id = c.id").
		Where("cc.contest_id = ?", contestID).
		Order("c.id ASC").
		Scan(&challenges).Error; err != nil {
		return nil, err
	}
	return challenges, nil
}

func (r *AWDRepository) ListServiceDefinitionsByContest(ctx context.Context, contestID int64) ([]contestports.AWDServiceDefinition, error) {
	rows, err := r.listContestAWDServiceRuntimeRows(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return r.listLegacyServiceDefinitionsByContest(ctx, contestID)
	}

	definitions := make([]contestports.AWDServiceDefinition, 0, len(rows))
	for _, row := range rows {
		runtimeConfig := contestdomain.ParseAWDCheckerConfig(row.RuntimeConfig)
		scoreConfig := contestdomain.ParseAWDCheckerConfig(row.ScoreConfig)
		definitions = append(definitions, contestports.AWDServiceDefinition{
			ChallengeID:   row.ChallengeID,
			FlagPrefix:    row.FlagPrefix,
			CheckerType:   resolveContestAWDServiceCheckerType(runtimeConfig, row.LegacyCheckerType),
			CheckerConfig: resolveContestAWDServiceCheckerConfig(runtimeConfig, row.LegacyCheckerConfig),
			SLAScore:      resolveContestAWDServiceScore(scoreConfig, "awd_sla_score", row.LegacySLAScore),
			DefenseScore:  resolveContestAWDServiceScore(scoreConfig, "awd_defense_score", row.LegacyDefenseScore),
		})
	}
	return definitions, nil
}

func (r *AWDRepository) ListReadinessChallengesByContest(ctx context.Context, contestID int64) ([]contestports.AWDReadinessChallengeRecord, error) {
	rows, err := r.listContestAWDServiceRuntimeRows(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return r.listLegacyReadinessChallengesByContest(ctx, contestID)
	}

	records := make([]contestports.AWDReadinessChallengeRecord, 0, len(rows))
	for _, row := range rows {
		runtimeConfig := contestdomain.ParseAWDCheckerConfig(row.RuntimeConfig)
		title := strings.TrimSpace(row.DisplayName)
		if title == "" {
			title = row.LegacyChallengeTitle
		}
		records = append(records, contestports.AWDReadinessChallengeRecord{
			ChallengeID:       row.ChallengeID,
			Title:             title,
			CheckerType:       resolveContestAWDServiceCheckerType(runtimeConfig, row.LegacyCheckerType),
			CheckerConfig:     resolveContestAWDServiceCheckerConfig(runtimeConfig, row.LegacyCheckerConfig),
			ValidationState:   row.LegacyValidationState,
			LastPreviewAt:     row.LegacyLastPreviewAt,
			LastPreviewResult: row.LegacyLastPreviewResult,
		})
	}
	return records, nil
}

func (r *AWDRepository) listLegacyServiceDefinitionsByContest(ctx context.Context, contestID int64) ([]contestports.AWDServiceDefinition, error) {
	var definitions []contestports.AWDServiceDefinition
	if err := r.dbWithContext(ctx).
		Table("contest_challenges AS cc").
		Select(`
			cc.challenge_id AS challenge_id,
			c.flag_prefix AS flag_prefix,
			cc.awd_checker_type AS awd_checker_type,
			cc.awd_checker_config AS awd_checker_config,
			cc.awd_sla_score AS awd_sla_score,
			cc.awd_defense_score AS awd_defense_score
		`).
		Joins("JOIN challenges AS c ON c.id = cc.challenge_id").
		Where("cc.contest_id = ?", contestID).
		Order("cc.challenge_id ASC").
		Scan(&definitions).Error; err != nil {
		return nil, err
	}
	return definitions, nil
}

func (r *AWDRepository) listLegacyReadinessChallengesByContest(ctx context.Context, contestID int64) ([]contestports.AWDReadinessChallengeRecord, error) {
	var records []contestports.AWDReadinessChallengeRecord
	if err := r.dbWithContext(ctx).
		Table("contest_challenges AS cc").
		Select(`
			cc.challenge_id AS challenge_id,
			c.title AS title,
			cc.awd_checker_type AS awd_checker_type,
			cc.awd_checker_config AS awd_checker_config,
			cc.awd_checker_validation_state AS awd_checker_validation_state,
			cc.awd_checker_last_preview_at AS awd_checker_last_preview_at,
			cc.awd_checker_last_preview_result AS awd_checker_last_preview_result
		`).
		Joins("JOIN challenges AS c ON c.id = cc.challenge_id").
		Where("cc.contest_id = ?", contestID).
		Order("cc.challenge_id ASC").
		Scan(&records).Error; err != nil {
		return nil, err
	}
	return records, nil
}

func (r *AWDRepository) listContestAWDServiceRuntimeRows(ctx context.Context, contestID int64) ([]awdContestServiceRuntimeRow, error) {
	var rows []awdContestServiceRuntimeRow
	if err := r.dbWithContext(ctx).
		Table("contest_awd_services AS cas").
		Select(`
			cas.challenge_id AS challenge_id,
			c.flag_prefix AS flag_prefix,
			cas.display_name AS display_name,
			cas.runtime_config AS runtime_config,
			cas.score_config AS score_config,
			cc.awd_checker_type AS legacy_checker_type,
			cc.awd_checker_config AS legacy_checker_config,
			cc.awd_sla_score AS legacy_sla_score,
			cc.awd_defense_score AS legacy_defense_score,
			cc.awd_checker_validation_state AS legacy_validation_state,
			cc.awd_checker_last_preview_at AS legacy_last_preview_at,
			cc.awd_checker_last_preview_result AS legacy_last_preview_result,
			c.title AS legacy_challenge_title
		`).
		Joins("JOIN challenges AS c ON c.id = cas.challenge_id").
		Joins("LEFT JOIN contest_challenges AS cc ON cc.contest_id = cas.contest_id AND cc.challenge_id = cas.challenge_id").
		Where("cas.contest_id = ?", contestID).
		Order("cas.\"order\" ASC, cas.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func resolveContestAWDServiceCheckerType(runtimeConfig map[string]any, fallback model.AWDCheckerType) model.AWDCheckerType {
	if runtimeConfig != nil {
		if raw, ok := runtimeConfig["checker_type"]; ok {
			if value, ok := raw.(string); ok {
				if normalized := contestdomain.NormalizeAWDCheckerType(value); normalized != "" {
					return normalized
				}
			}
		}
	}
	return fallback
}

func resolveContestAWDServiceCheckerConfig(runtimeConfig map[string]any, fallback string) string {
	if runtimeConfig != nil {
		if raw, ok := runtimeConfig["checker_config"]; ok {
			if encoded := marshalContestAWDServiceJSON(raw); encoded != "" {
				return encoded
			}
		}
	}
	return fallback
}

func resolveContestAWDServiceScore(scoreConfig map[string]any, key string, fallback int) int {
	if scoreConfig != nil {
		if raw, ok := scoreConfig[key]; ok {
			if value, ok := normalizeContestAWDServiceInt(raw); ok {
				return value
			}
		}
	}
	return fallback
}

func normalizeContestAWDServiceInt(value any) (int, bool) {
	switch typed := value.(type) {
	case int:
		return typed, true
	case int32:
		return int(typed), true
	case int64:
		return int(typed), true
	case float64:
		return int(typed), true
	case json.Number:
		next, err := typed.Int64()
		if err != nil {
			return 0, false
		}
		return int(next), true
	default:
		return 0, false
	}
}

func marshalContestAWDServiceJSON(value any) string {
	raw, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(raw)
}

func (r *AWDRepository) ContestHasChallenge(ctx context.Context, contestID, challengeID int64) (bool, error) {
	var count int64
	if err := r.dbWithContext(ctx).
		Model(&model.ContestChallenge{}).
		Where("contest_id = ? AND challenge_id = ?", contestID, challengeID).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AWDRepository) FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	var challenge model.Challenge
	if err := r.dbWithContext(ctx).Where("id = ?", challengeID).First(&challenge).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}
