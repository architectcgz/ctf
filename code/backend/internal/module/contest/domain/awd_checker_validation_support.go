package domain

import (
	"encoding/json"
	"strings"

	"ctf-platform/internal/model"
)

type AWDCheckerPreviewContext struct {
	ServiceID      int64  `json:"service_id"`
	AccessURL      string `json:"access_url"`
	PreviewFlag    string `json:"preview_flag"`
	RoundNumber    int    `json:"round_number"`
	TeamID         int64  `json:"team_id"`
	AWDChallengeID int64  `json:"awd_challenge_id"`
}

type AWDCheckerPreviewResult struct {
	CheckerType    model.AWDCheckerType     `json:"checker_type,omitempty"`
	ServiceStatus  string                   `json:"service_status"`
	CheckResult    map[string]any           `json:"check_result"`
	PreviewContext AWDCheckerPreviewContext `json:"preview_context"`
	PreviewToken   string                   `json:"preview_token,omitempty"`
}

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

func MarshalAWDCheckerPreviewResult(value *AWDCheckerPreviewResult) (string, error) {
	if value == nil {
		return "", nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func ParseAWDCheckerPreviewResult(value string) *AWDCheckerPreviewResult {
	if strings.TrimSpace(value) == "" {
		return nil
	}

	var result AWDCheckerPreviewResult
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return nil
	}
	result.CheckerType = NormalizeAWDCheckerType(string(result.CheckerType))
	if result.CheckResult == nil {
		result.CheckResult = map[string]any{}
	}
	return &result
}
