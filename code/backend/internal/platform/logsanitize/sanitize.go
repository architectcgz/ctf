package logsanitize

import "strings"

const RedactedValue = "[redacted]"

func SanitizePassword(string) string {
	return RedactedValue
}

func SanitizeToken(string) string {
	return RedactedValue
}

func SanitizeSecret(string) string {
	return RedactedValue
}

func SanitizeKey(key string) string {
	trimmed := strings.TrimSpace(key)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= 12 {
		return trimmed
	}

	if separator := strings.LastIndex(trimmed, ":"); separator > 0 && separator < len(trimmed)-1 {
		namespace := trimmed[:separator+1]
		suffix := trimmed[separator+1:]
		if len(suffix) <= 5 {
			return namespace + "..."
		}
		return namespace + suffix[:5] + "..."
	}

	return trimmed[:8] + "..."
}
