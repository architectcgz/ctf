package mapperhelper

import (
	"strings"
)

func NormalizeOptionalString(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func NormalizeOptionalTrimmedString(value string) *string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
