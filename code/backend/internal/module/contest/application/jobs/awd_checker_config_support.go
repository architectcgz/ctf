package jobs

import (
	"strings"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func effectiveAWDCheckerType(value model.AWDCheckerType) model.AWDCheckerType {
	normalized := contestdomain.NormalizeAWDCheckerType(string(value))
	if normalized == "" {
		return model.AWDCheckerTypeLegacyProbe
	}
	return normalized
}

func resolveAWDCheckerHealthPath(checkerConfig, fallback string) string {
	if configuredPath := parseAWDCheckerHealthPath(checkerConfig); configuredPath != "" {
		return normalizedAWDCheckerHealthPath(configuredPath)
	}
	return normalizedAWDCheckerHealthPath(fallback)
}

func parseAWDCheckerHealthPath(value string) string {
	if strings.TrimSpace(value) == "" {
		return ""
	}

	config, err := parseAWDHTTPCheckerConfig(value)
	if err != nil {
		return ""
	}

	if path := strings.TrimSpace(config.GetFlag.Path); path != "" {
		return path
	}
	if path := strings.TrimSpace(config.Havoc.Path); path != "" {
		return path
	}
	return ""
}
