package infrastructure

import (
	"encoding/json"
	"strings"

	contestentity "ctf-platform/internal/module/contest/entity"
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
	case contestentity.AWDAttackSourceManual:
		return contestentity.AWDAttackSourceManual
	case contestentity.AWDAttackSourceSubmission:
		return contestentity.AWDAttackSourceSubmission
	default:
		return contestentity.AWDAttackSourceLegacy
	}
}
