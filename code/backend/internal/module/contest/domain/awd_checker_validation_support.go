package domain

import (
	"encoding/json"
	"strings"

	contestentity "ctf-platform/internal/module/contest/entity"
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
	CheckerType    contestentity.AWDCheckerType `json:"checker_type,omitempty"`
	ServiceStatus  string                       `json:"service_status"`
	CheckResult    map[string]any               `json:"check_result"`
	PreviewContext AWDCheckerPreviewContext     `json:"preview_context"`
	PreviewToken   string                       `json:"preview_token,omitempty"`
}

func NormalizeAWDCheckerValidationState(value string) contestentity.AWDCheckerValidationState {
	switch strings.TrimSpace(value) {
	case string(contestentity.AWDCheckerValidationStatePassed):
		return contestentity.AWDCheckerValidationStatePassed
	case string(contestentity.AWDCheckerValidationStateFailed):
		return contestentity.AWDCheckerValidationStateFailed
	case string(contestentity.AWDCheckerValidationStateStale):
		return contestentity.AWDCheckerValidationStateStale
	default:
		return contestentity.AWDCheckerValidationStatePending
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
