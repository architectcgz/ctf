package jobs

import (
	"strings"

	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
)

func effectiveAWDCheckerType(value contestentity.AWDCheckerType) contestentity.AWDCheckerType {
	normalized := contestdomain.NormalizeAWDCheckerType(string(value))
	if normalized == "" {
		return contestentity.AWDCheckerTypeLegacyProbe
	}
	return normalized
}

func resolveAWDCheckerHealthPath(checkerConfig, fallback string) string {
	if configuredPath := parseAWDCheckerHealthPath(checkerConfig); configuredPath != "" {
		return normalizedAWDCheckerHealthPath(configuredPath)
	}
	return normalizedAWDCheckerHealthPath(fallback)
}

func resolveAWDPreviewWarmupHealthPath(checkerConfig, fallback string) string {
	if strings.TrimSpace(checkerConfig) != "" {
		config, err := parseAWDHTTPCheckerConfig(checkerConfig)
		if err == nil {
			if path := strings.TrimSpace(config.Havoc.Path); path != "" {
				return normalizedAWDCheckerHealthPath(path)
			}
		}
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
