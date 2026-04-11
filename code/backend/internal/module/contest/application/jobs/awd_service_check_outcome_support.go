package jobs

import (
	"encoding/json"

	"ctf-platform/internal/model"
)

func buildAWDCheckOutcome(result awdServiceCheckResult, serviceStatus string) (*awdServiceCheckOutcome, error) {
	raw, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &awdServiceCheckOutcome{
		serviceStatus: serviceStatus,
		checkResult:   string(raw),
	}, nil
}

func buildAWDCheckOutcomeWithError(result awdServiceCheckResult, serviceStatus, errorCode, errorMessage string) (*awdServiceCheckOutcome, error) {
	result.StatusReason = errorCode
	result.ErrorCode = errorCode
	result.Error = errorMessage
	return buildAWDCheckOutcome(result, serviceStatus)
}

func buildAWDDownCheckOutcome(result awdServiceCheckResult, errorCode, errorMessage string) (*awdServiceCheckOutcome, error) {
	return buildAWDCheckOutcomeWithError(result, model.AWDServiceStatusDown, errorCode, errorMessage)
}
