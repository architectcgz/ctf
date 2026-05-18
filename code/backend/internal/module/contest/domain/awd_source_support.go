package domain

import (
	"strings"

	contestentity "ctf-platform/internal/module/contest/entity"
)

const (
	AWDCheckSourceScheduler      = "scheduler"
	AWDCheckSourceManualCurrent  = "manual_current_round"
	AWDCheckSourceManualSelected = "manual_selected_round"
	AWDCheckSourceManualService  = "manual_service_check"
)

func NormalizeAWDAttackSource(value string) string {
	switch strings.TrimSpace(value) {
	case contestentity.AWDAttackSourceManual:
		return contestentity.AWDAttackSourceManual
	case contestentity.AWDAttackSourceSubmission:
		return contestentity.AWDAttackSourceSubmission
	default:
		return contestentity.AWDAttackSourceLegacy
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
