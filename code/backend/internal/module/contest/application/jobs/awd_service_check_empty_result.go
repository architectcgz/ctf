package jobs

import "ctf-platform/internal/model"

func buildAWDCheckOutcomeWithoutInstances(result awdServiceCheckResult) (*awdServiceCheckOutcome, error) {
	result.StatusReason = "no_running_instances"
	result.ErrorCode = "no_running_instances"
	result.Error = "no_running_instances"
	return buildAWDCheckOutcome(result, model.AWDServiceStatusDown)
}
