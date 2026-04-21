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
	ServiceID         int64                           `gorm:"column:service_id"`
	ChallengeID       int64                           `gorm:"column:challenge_id"`
	DisplayName       string                          `gorm:"column:display_name"`
	ServiceSnapshot   string                          `gorm:"column:service_snapshot"`
	RuntimeConfig     string                          `gorm:"column:runtime_config"`
	ScoreConfig       string                          `gorm:"column:score_config"`
	ValidationState   model.AWDCheckerValidationState `gorm:"column:validation_state"`
	LastPreviewAt     *time.Time                      `gorm:"column:last_preview_at"`
	LastPreviewResult string                          `gorm:"column:last_preview_result"`
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
	rows, err := r.listContestAWDServiceRuntimeRows(ctx, contestID)
	if err != nil {
		return nil, err
	}

	challenges := make([]model.Challenge, 0, len(rows))
	for _, row := range rows {
		snapshot, decodeErr := model.DecodeContestAWDServiceSnapshot(row.ServiceSnapshot)
		if decodeErr != nil {
			return nil, decodeErr
		}
		challenges = append(challenges, model.Challenge{
			ID:         row.ChallengeID,
			Title:      resolveContestAWDServiceTitle(snapshot, row.DisplayName),
			Category:   snapshot.Category,
			Difficulty: snapshot.Difficulty,
			FlagType:   resolveContestAWDServiceFlagType(snapshot),
			FlagPrefix: resolveContestAWDServiceFlagPrefix(snapshot),
			Points:     resolveContestAWDServiceScore(contestdomain.ParseAWDCheckerConfig(row.ScoreConfig), "points"),
			Status:     model.ChallengeStatusPublished,
		})
	}
	return challenges, nil
}

func (r *AWDRepository) ListServiceDefinitionsByContest(ctx context.Context, contestID int64) ([]contestports.AWDServiceDefinition, error) {
	rows, err := r.listContestAWDServiceRuntimeRows(ctx, contestID)
	if err != nil {
		return nil, err
	}

	definitions := make([]contestports.AWDServiceDefinition, 0, len(rows))
	for _, row := range rows {
		snapshot, decodeErr := model.DecodeContestAWDServiceSnapshot(row.ServiceSnapshot)
		if decodeErr != nil {
			return nil, decodeErr
		}
		runtimeConfig := contestdomain.ParseAWDCheckerConfig(row.RuntimeConfig)
		scoreConfig := contestdomain.ParseAWDCheckerConfig(row.ScoreConfig)
		definitions = append(definitions, contestports.AWDServiceDefinition{
			ServiceID:     row.ServiceID,
			ChallengeID:   row.ChallengeID,
			FlagPrefix:    resolveContestAWDServiceFlagPrefix(snapshot),
			CheckerType:   resolveContestAWDServiceCheckerType(runtimeConfig),
			CheckerConfig: resolveContestAWDServiceCheckerConfig(runtimeConfig),
			SLAScore:      resolveContestAWDServiceScore(scoreConfig, "awd_sla_score"),
			DefenseScore:  resolveContestAWDServiceScore(scoreConfig, "awd_defense_score"),
		})
	}
	return definitions, nil
}

func (r *AWDRepository) ListReadinessChallengesByContest(ctx context.Context, contestID int64) ([]contestports.AWDReadinessChallengeRecord, error) {
	rows, err := r.listContestAWDServiceRuntimeRows(ctx, contestID)
	if err != nil {
		return nil, err
	}

	records := make([]contestports.AWDReadinessChallengeRecord, 0, len(rows))
	for _, row := range rows {
		snapshot, decodeErr := model.DecodeContestAWDServiceSnapshot(row.ServiceSnapshot)
		if decodeErr != nil {
			return nil, decodeErr
		}
		runtimeConfig := contestdomain.ParseAWDCheckerConfig(row.RuntimeConfig)
		records = append(records, contestports.AWDReadinessChallengeRecord{
			ServiceID:         row.ServiceID,
			ChallengeID:       row.ChallengeID,
			Title:             resolveContestAWDServiceTitle(snapshot, row.DisplayName),
			CheckerType:       resolveContestAWDServiceCheckerType(runtimeConfig),
			CheckerConfig:     resolveContestAWDServiceCheckerConfig(runtimeConfig),
			ValidationState:   row.ValidationState,
			LastPreviewAt:     row.LastPreviewAt,
			LastPreviewResult: row.LastPreviewResult,
		})
	}
	return records, nil
}

func (r *AWDRepository) listContestAWDServiceRuntimeRows(ctx context.Context, contestID int64) ([]awdContestServiceRuntimeRow, error) {
	var rows []awdContestServiceRuntimeRow
	if err := r.dbWithContext(ctx).
		Table("contest_awd_services AS cas").
		Select(`
			cas.id AS service_id,
			cas.challenge_id AS challenge_id,
			cas.display_name AS display_name,
			cas.service_snapshot AS service_snapshot,
			cas.runtime_config AS runtime_config,
			cas.score_config AS score_config,
			cas.awd_checker_validation_state AS validation_state,
			cas.awd_checker_last_preview_at AS last_preview_at,
			cas.awd_checker_last_preview_result AS last_preview_result
		`).
		Where("cas.contest_id = ?", contestID).
		Where("cas.deleted_at IS NULL").
		Order("cas.\"order\" ASC, cas.id ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func resolveContestAWDServiceCheckerType(runtimeConfig map[string]any) model.AWDCheckerType {
	if runtimeConfig != nil {
		if raw, ok := runtimeConfig["checker_type"]; ok {
			if value, ok := raw.(string); ok {
				if normalized := contestdomain.NormalizeAWDCheckerType(value); normalized != "" {
					return normalized
				}
			}
		}
	}
	return ""
}

func resolveContestAWDServiceTitle(snapshot model.ContestAWDServiceSnapshot, displayName string) string {
	if title := strings.TrimSpace(displayName); title != "" {
		return title
	}
	return strings.TrimSpace(snapshot.Name)
}

func resolveContestAWDServiceFlagPrefix(snapshot model.ContestAWDServiceSnapshot) string {
	if snapshot.FlagConfig != nil {
		if value, ok := snapshot.FlagConfig["flag_prefix"].(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return "flag"
}

func resolveContestAWDServiceFlagType(snapshot model.ContestAWDServiceSnapshot) string {
	if snapshot.FlagConfig != nil {
		if value, ok := snapshot.FlagConfig["flag_type"].(string); ok {
			if trimmed := strings.TrimSpace(value); trimmed != "" {
				return trimmed
			}
		}
	}
	return model.FlagTypeDynamic
}

func resolveContestAWDServiceCheckerConfig(runtimeConfig map[string]any) string {
	if runtimeConfig != nil {
		if raw, ok := runtimeConfig["checker_config_raw"]; ok {
			if value, ok := raw.(string); ok {
				return value
			}
		}
		if raw, ok := runtimeConfig["checker_config"]; ok {
			if encoded := marshalContestAWDServiceJSON(raw); encoded != "" {
				return encoded
			}
		}
	}
	return ""
}

func resolveContestAWDServiceScore(scoreConfig map[string]any, key string) int {
	if scoreConfig != nil {
		if raw, ok := scoreConfig[key]; ok {
			if value, ok := normalizeContestAWDServiceInt(raw); ok {
				return value
			}
		}
	}
	return 0
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

func (r *AWDRepository) FindChallengeByID(ctx context.Context, challengeID int64) (*model.Challenge, error) {
	var challenge model.Challenge
	if err := r.dbWithContext(ctx).Where("id = ?", challengeID).First(&challenge).Error; err != nil {
		return nil, err
	}
	return &challenge, nil
}
