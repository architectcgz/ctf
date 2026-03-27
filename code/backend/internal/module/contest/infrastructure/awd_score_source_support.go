package infrastructure

import (
	"encoding/json"
	"strings"

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
