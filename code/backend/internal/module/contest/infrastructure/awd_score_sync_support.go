package infrastructure

import (
	"encoding/json"
	"strings"
	"time"

	"ctf-platform/internal/model"
)

func normalizeAWDCheckSourceValue(value any) string {
	raw, ok := value.(string)
	if !ok {
		return ""
	}
	switch strings.TrimSpace(raw) {
	case "scheduler":
		return "scheduler"
	case "manual_current_round":
		return "manual_current_round"
	case "manual_selected_round":
		return "manual_selected_round"
	case "manual_service_check":
		return "manual_service_check"
	default:
		return ""
	}
}

func parseAWDCheckResultValue(value string) map[string]any {
	if strings.TrimSpace(value) == "" {
		return map[string]any{}
	}
	result := make(map[string]any)
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return map[string]any{}
	}
	return result
}

func normalizeAWDAttackSourceValue(value string) string {
	switch strings.TrimSpace(value) {
	case model.AWDAttackSourceManual:
		return model.AWDAttackSourceManual
	case model.AWDAttackSourceSubmission:
		return model.AWDAttackSourceSubmission
	default:
		return model.AWDAttackSourceLegacy
	}
}

func parseAWDScoreSyncTime(raw string) *time.Time {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	layouts := []string{
		time.RFC3339Nano,
		"2006-01-02 15:04:05.999999999-07:00",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		parsed, err := time.Parse(layout, trimmed)
		if err == nil {
			return &parsed
		}
	}
	return nil
}
