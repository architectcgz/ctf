package commands

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func buildContestAWDServiceSnapshot(awdChallenge *model.AWDChallenge) string {
	if awdChallenge == nil {
		return "{}"
	}
	snapshot := model.ContestAWDServiceSnapshot{
		Name:             strings.TrimSpace(awdChallenge.Name),
		Category:         strings.TrimSpace(awdChallenge.Category),
		Difficulty:       strings.TrimSpace(awdChallenge.Difficulty),
		Description:      strings.TrimSpace(awdChallenge.Description),
		ServiceType:      awdChallenge.ServiceType,
		DeploymentMode:   awdChallenge.DeploymentMode,
		FlagMode:         strings.TrimSpace(awdChallenge.FlagMode),
		FlagConfig:       parseContestAWDServiceJSONMap(awdChallenge.FlagConfig),
		DefenseEntryMode: strings.TrimSpace(awdChallenge.DefenseEntryMode),
		AccessConfig:     parseContestAWDServiceJSONMap(awdChallenge.AccessConfig),
		RuntimeConfig:    parseContestAWDServiceJSONMap(awdChallenge.RuntimeConfig),
	}
	raw, err := model.EncodeContestAWDServiceSnapshot(snapshot)
	if err != nil {
		return "{}"
	}
	return raw
}

func parseContestAWDServiceJSONMap(raw string) map[string]any {
	value := strings.TrimSpace(raw)
	if value == "" {
		return map[string]any{}
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(value), &payload); err != nil {
		return map[string]any{}
	}
	return payload
}

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
	checkerType model.AWDCheckerType,
	checkerConfig string,
	extraRuntimeConfig string,
) string {
	value := map[string]any{
		"checker_type":   contestdomain.NormalizeAWDCheckerType(string(checkerType)),
		"checker_config": contestdomain.ParseAWDCheckerConfig(checkerConfig),
	}
	if normalizedCheckerConfig, err := contestdomain.MarshalAWDCheckerConfig(contestdomain.ParseAWDCheckerConfig(checkerConfig)); err == nil {
		value["checker_config_raw"] = normalizedCheckerConfig
	}
	if extra := contestdomain.ParseAWDCheckerConfig(extraRuntimeConfig); len(extra) > 0 {
		value["challenge_runtime"] = extra
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
	if slaScoreOverride == nil && slaScore == 0 {
		slaScore = contestdomain.AWDDefaultServiceSLAScore
	}
	if slaScoreOverride != nil {
		slaScore = *slaScoreOverride
	}
	defenseScore := defaultDefenseScore
	if defenseScoreOverride == nil && defenseScore == 0 {
		defenseScore = contestdomain.AWDDefaultServiceDefenseScore
	}
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

func parseContestAWDChallengeRuntimeConfig(runtimeConfig string) string {
	value := contestdomain.ParseAWDCheckerConfig(runtimeConfig)
	raw, ok := value["challenge_runtime"]
	if !ok {
		return ""
	}
	encoded, err := json.Marshal(raw)
	if err != nil {
		return ""
	}
	return string(encoded)
}

func parseContestAWDServiceCheckerTokenEnv(runtimeConfig string) string {
	value := contestdomain.ParseAWDCheckerConfig(runtimeConfig)
	if checkerTokenEnv := strings.TrimSpace(readStringFromAny(value["checker_token_env"])); checkerTokenEnv != "" {
		return checkerTokenEnv
	}
	challengeRuntime, _ := value["challenge_runtime"].(map[string]any)
	return strings.TrimSpace(readStringFromAny(challengeRuntime["checker_token_env"]))
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
	previewTokenStore contestports.AWDCheckerPreviewTokenStore,
	current *model.ContestAWDService,
	contestID int64,
	nextCheckerType model.AWDCheckerType,
	nextCheckerConfig string,
	nextCheckerTokenEnv string,
	previewToken string,
) (model.AWDCheckerValidationState, *time.Time, string, bool, error) {
	if current == nil {
		return model.AWDCheckerValidationStatePending, nil, "", false, nil
	}

	state, previewAt, previewResult, err := consumeCheckerPreviewValidationState(
		ctx,
		previewTokenStore,
		contestID,
		current.ID,
		current.AWDChallengeID,
		nextCheckerType,
		nextCheckerConfig,
		nextCheckerTokenEnv,
		previewToken,
	)
	if err != nil {
		return model.AWDCheckerValidationStatePending, nil, "", false, err
	}
	if err := ensureCheckerPreviewTokenConsumed(previewToken, previewResult); err != nil {
		return model.AWDCheckerValidationStatePending, nil, "", false, err
	}
	if strings.TrimSpace(previewToken) != "" && previewResult != "" {
		return state, previewAt, previewResult, true, nil
	}

	currentCheckerType, currentCheckerConfig := parseContestAWDServiceRuntimeChecker(current.RuntimeConfig)
	currentCheckerTokenEnv := parseContestAWDServiceCheckerTokenEnv(current.RuntimeConfig)
	if currentCheckerType == nextCheckerType &&
		currentCheckerConfig == nextCheckerConfig &&
		currentCheckerTokenEnv == nextCheckerTokenEnv {
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

func ensureCheckerPreviewTokenConsumed(previewToken, previewResult string) error {
	if strings.TrimSpace(previewToken) == "" {
		return nil
	}
	if strings.TrimSpace(previewResult) != "" {
		return nil
	}
	return errcode.ErrAWDCheckerPreviewExpired
}
