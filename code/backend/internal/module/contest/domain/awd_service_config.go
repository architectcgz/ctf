package domain

import (
	"encoding/json"
	"strings"

	"ctf-platform/internal/model"
)

func NormalizeAWDCheckerType(value string) model.AWDCheckerType {
	switch strings.TrimSpace(value) {
	case string(model.AWDCheckerTypeLegacyProbe):
		return model.AWDCheckerTypeLegacyProbe
	case string(model.AWDCheckerTypeHTTPStandard):
		return model.AWDCheckerTypeHTTPStandard
	case string(model.AWDCheckerTypeTCPStandard):
		return model.AWDCheckerTypeTCPStandard
	case string(model.AWDCheckerTypeScript):
		return model.AWDCheckerTypeScript
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
