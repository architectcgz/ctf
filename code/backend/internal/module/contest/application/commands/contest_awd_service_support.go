package commands

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	redislib "github.com/redis/go-redis/v9"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func buildContestAWDServiceScoreConfig(points, slaScore, defenseScore int) string {
	value := map[string]any{
		"points":            points,
		"awd_sla_score":     slaScore,
		"awd_defense_score": defenseScore,
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func buildContestAWDServiceRuntimeConfig(
	challengeID int64,
	checkerType model.AWDCheckerType,
	checkerConfig string,
	extraRuntimeConfig string,
) string {
	value := map[string]any{
		// challenge_id remains an internal compatibility shadow field.
		// New code must use contest_awd_services.challenge_id as the source of truth
		// and must not expose or accept this runtime field as an independent contract.
		"challenge_id":   challengeID,
		"checker_type":   contestdomain.NormalizeAWDCheckerType(string(checkerType)),
		"checker_config": contestdomain.ParseAWDCheckerConfig(checkerConfig),
	}
	if normalizedCheckerConfig, err := contestdomain.MarshalAWDCheckerConfig(contestdomain.ParseAWDCheckerConfig(checkerConfig)); err == nil {
		value["checker_config_raw"] = normalizedCheckerConfig
	}
	if extra := contestdomain.ParseAWDCheckerConfig(extraRuntimeConfig); len(extra) > 0 {
		value["template_runtime"] = extra
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func normalizeContestAWDServiceRuntimeFields(
	contest *model.Contest,
	defaultCheckerType model.AWDCheckerType,
	defaultCheckerConfig string,
	checkerTypeOverride *string,
	checkerConfigOverride map[string]any,
	defaultSLAScore, defaultDefenseScore int,
	slaScoreOverride, defenseScoreOverride *int,
) (model.AWDCheckerType, string, int, int, error) {
	checkerTypeValue := string(defaultCheckerType)
	if checkerTypeOverride != nil {
		checkerTypeValue = strings.TrimSpace(*checkerTypeOverride)
	}
	checkerConfigValue := contestdomain.ParseAWDCheckerConfig(defaultCheckerConfig)
	if checkerConfigOverride != nil {
		checkerConfigValue = checkerConfigOverride
	}
	slaScore := defaultSLAScore
	if slaScoreOverride != nil {
		slaScore = *slaScoreOverride
	}
	defenseScore := defaultDefenseScore
	if defenseScoreOverride != nil {
		defenseScore = *defenseScoreOverride
	}

	checkerType, checkerConfig, err := validateAndNormalizeContestAWDFields(
		contest,
		checkerTypeValue,
		checkerConfigValue,
		slaScore,
		defenseScore,
	)
	if err != nil {
		return "", "", 0, 0, err
	}
	return checkerType, checkerConfig, slaScore, defenseScore, nil
}

func parseContestAWDServiceRuntimeChecker(runtimeConfig string) (model.AWDCheckerType, string) {
	value := contestdomain.ParseAWDCheckerConfig(runtimeConfig)

	checkerType := model.AWDCheckerType("")
	if raw, ok := value["checker_type"]; ok {
		if typed, ok := raw.(string); ok {
			checkerType = contestdomain.NormalizeAWDCheckerType(typed)
		}
	}
	if raw, ok := value["checker_config_raw"]; ok {
		if typed, ok := raw.(string); ok && strings.TrimSpace(typed) != "" {
			return checkerType, typed
		}
	}
	if raw, ok := value["checker_config"]; ok {
		if encoded, err := json.Marshal(raw); err == nil {
			return checkerType, string(encoded)
		}
	}
	return checkerType, "{}"
}

func parseContestAWDServiceTemplateRuntimeConfig(runtimeConfig string) string {
	value := contestdomain.ParseAWDCheckerConfig(runtimeConfig)
	raw, ok := value["template_runtime"]
	if !ok {
		return ""
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func parseContestAWDServiceScore(scoreConfig string, key string) (int, bool) {
	value := contestdomain.ParseAWDCheckerConfig(scoreConfig)
	raw, ok := value[key]
	if !ok {
		return 0, false
	}
	switch typed := raw.(type) {
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

func buildContestAWDServiceValidationUpdate(
	ctx context.Context,
	redisClient *redislib.Client,
	current *model.ContestAWDService,
	contestID int64,
	nextCheckerType model.AWDCheckerType,
	nextCheckerConfig string,
	previewToken string,
) (model.AWDCheckerValidationState, *time.Time, string, bool, error) {
	if current == nil {
		return model.AWDCheckerValidationStatePending, nil, "", false, nil
	}

	state, previewAt, previewResult, err := consumeCheckerPreviewValidationState(
		ctx,
		redisClient,
		contestID,
		current.ID,
		current.ChallengeID,
		nextCheckerType,
		nextCheckerConfig,
		previewToken,
	)
	if err != nil {
		return model.AWDCheckerValidationStatePending, nil, "", false, err
	}
	if strings.TrimSpace(previewToken) != "" && previewResult != "" {
		return state, previewAt, previewResult, true, nil
	}

	currentCheckerType, currentCheckerConfig := parseContestAWDServiceRuntimeChecker(current.RuntimeConfig)
	if currentCheckerType == nextCheckerType && currentCheckerConfig == nextCheckerConfig {
		return current.ValidationState, current.LastPreviewAt, current.LastPreviewResult, false, nil
	}

	nextState := model.AWDCheckerValidationStatePending
	if current.LastPreviewAt != nil ||
		strings.TrimSpace(current.LastPreviewResult) != "" ||
		(current.ValidationState != "" && current.ValidationState != model.AWDCheckerValidationStatePending) {
		nextState = model.AWDCheckerValidationStateStale
	}
	return nextState, current.LastPreviewAt, current.LastPreviewResult, true, nil
}
