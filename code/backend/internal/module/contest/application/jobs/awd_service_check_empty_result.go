package jobs

import (
	"encoding/json"

	"ctf-platform/internal/model"
)

func buildAWDCheckOutcomeWithoutInstances(result awdServiceCheckResult) (*awdServiceCheckOutcome, error) {
	result.StatusReason = "no_running_instances"
	result.ErrorCode = "no_running_instances"
	result.Error = "no_running_instances"

	raw, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &awdServiceCheckOutcome{
		serviceStatus: model.AWDServiceStatusDown,
		checkResult:   string(raw),
	}, nil
}
