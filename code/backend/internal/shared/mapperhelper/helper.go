package mapperhelper

import (
	"strings"
	"time"
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

func CopyTimeToPtr(value time.Time) *time.Time {
	copied := value
	return &copied
}

func SingleString(value string) []string {
	return []string{value}
}
