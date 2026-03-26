package jobs

import (
	"errors"
	"net/url"
	"path"
	"strings"
	"time"
)

func buildAWDHealthCheckURL(accessURL, healthPath string) (string, error) {
	parsed, err := url.Parse(strings.TrimSpace(accessURL))
	if err != nil {
		return "", newAWDCheckError("invalid_access_url", sanitizeAWDCheckError(err))
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", newAWDCheckError("invalid_access_url", "invalid_access_url")
	}
	parsed.Path = path.Join("/", strings.TrimSpace(parsed.Path), strings.TrimSpace(healthPath))
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String(), nil
}

func newAWDCheckError(code, message string) error {
	message = strings.TrimSpace(message)
	if message == "" {
		message = code
	}
	return awdCheckError{code: code, message: message}
}

func normalizeAWDCheckError(err error, fallbackCode string) (string, string) {
	if err == nil {
		return "", ""
	}
	var typedErr awdCheckError
	if ok := errors.As(err, &typedErr); ok {
		return typedErr.code, sanitizeAWDCheckError(typedErr)
	}
	return fallbackCode, sanitizeAWDCheckError(err)
}

func sanitizeAWDCheckError(err error) string {
	if err == nil {
		return ""
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return "unknown_checker_error"
	}
	return msg
}

func normalizedAWDCheckerTimeout(value time.Duration) time.Duration {
	if value <= 0 {
		return 3 * time.Second
	}
	return value
}

func normalizedAWDCheckerHealthPath(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "/health"
	}
	if !strings.HasPrefix(trimmed, "/") {
		return "/" + trimmed
	}
	return trimmed
}
