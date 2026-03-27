package domain

import (
	"strings"

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
