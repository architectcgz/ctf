package domain

import (
	"encoding/json"
	"strings"
	"time"

	"ctf-platform/internal/model"
)

const (
	AWDCheckSourceScheduler      = "scheduler"
	AWDCheckSourceManualCurrent  = "manual_current_round"
	AWDCheckSourceManualSelected = "manual_selected_round"
	AWDCheckSourceManualService  = "manual_service_check"
)

func NormalizeAWDAttackSource(value string) string {
	switch strings.TrimSpace(value) {
	case model.AWDAttackSourceManual:
		return model.AWDAttackSourceManual
	case model.AWDAttackSourceSubmission:
		return model.AWDAttackSourceSubmission
	default:
		return model.AWDAttackSourceLegacy
	}
}

func NormalizeAWDCheckSource(value any) string {
	raw, ok := value.(string)
	if !ok {
		return ""
	}
	switch strings.TrimSpace(raw) {
	case AWDCheckSourceScheduler:
		return AWDCheckSourceScheduler
	case AWDCheckSourceManualCurrent:
		return AWDCheckSourceManualCurrent
	case AWDCheckSourceManualSelected:
		return AWDCheckSourceManualSelected
	case AWDCheckSourceManualService:
		return AWDCheckSourceManualService
	default:
		return ""
	}
}

func NormalizedAWDCheckSource(value string) string {
	switch strings.TrimSpace(value) {
	case AWDCheckSourceManualCurrent:
		return AWDCheckSourceManualCurrent
	case AWDCheckSourceManualSelected:
		return AWDCheckSourceManualSelected
	case AWDCheckSourceManualService:
		return AWDCheckSourceManualService
	default:
		return AWDCheckSourceScheduler
	}
}

func MarshalAWDCheckResult(value map[string]any) (string, error) {
	if len(value) == 0 {
		return "{}", nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func NormalizeManualAWDCheckResult(value map[string]any) map[string]any {
	result := make(map[string]any, len(value)+2)
	for key, item := range value {
		result[key] = item
	}
	result["check_source"] = AWDCheckSourceManualService
	if checkedAt, ok := result["checked_at"].(string); !ok || strings.TrimSpace(checkedAt) == "" {
		result["checked_at"] = time.Now().UTC().Format(time.RFC3339)
	}
	return result
}

func ParseAWDCheckResult(value string) map[string]any {
	if strings.TrimSpace(value) == "" {
		return map[string]any{}
	}
	result := make(map[string]any)
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return map[string]any{}
	}
	return result
}
