package domain

import (
	"encoding/json"
	"strings"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func NormalizeAWDCheckerType(value string) contestentity.AWDCheckerType {
	switch strings.TrimSpace(value) {
	case string(contestentity.AWDCheckerTypeLegacyProbe):
		return contestentity.AWDCheckerTypeLegacyProbe
	case string(contestentity.AWDCheckerTypeHTTPStandard):
		return contestentity.AWDCheckerTypeHTTPStandard
	case string(contestentity.AWDCheckerTypeTCPStandard):
		return contestentity.AWDCheckerTypeTCPStandard
	case string(contestentity.AWDCheckerTypeScript):
		return contestentity.AWDCheckerTypeScript
	default:
		return ""
	}
}

func MarshalAWDCheckerConfig(value map[string]any) (string, error) {
	if len(value) == 0 {
		return "{}", nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(raw), nil
}

func ParseAWDCheckerConfig(value string) map[string]any {
	if strings.TrimSpace(value) == "" {
		return map[string]any{}
	}
	result := make(map[string]any)
	if err := json.Unmarshal([]byte(value), &result); err != nil {
		return map[string]any{}
	}
	return result
}
