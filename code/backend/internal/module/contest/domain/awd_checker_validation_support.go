package domain

import (
	"encoding/json"
	"strings"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func NormalizeAWDCheckerValidationState(value string) model.AWDCheckerValidationState {
	switch strings.TrimSpace(value) {
	case string(model.AWDCheckerValidationStatePassed):
		return model.AWDCheckerValidationStatePassed
	case string(model.AWDCheckerValidationStateFailed):
		return model.AWDCheckerValidationStateFailed
	case string(model.AWDCheckerValidationStateStale):
		return model.AWDCheckerValidationStateStale
	default:
		return model.AWDCheckerValidationStatePending
	}
}

func MarshalAWDCheckerPreviewResult(value *dto.AWDCheckerPreviewResp) (string, error) {
	if value == nil {
		return "", nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func ParseAWDCheckerPreviewResult(value string) *dto.AWDCheckerPreviewResp {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	var result dto.AWDCheckerPreviewResp
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil
	}
	result.CheckerType = NormalizeAWDCheckerType(string(result.CheckerType))
	if result.CheckResult == nil {
		result.CheckResult = map[string]any{}
	}
	return &result
}
