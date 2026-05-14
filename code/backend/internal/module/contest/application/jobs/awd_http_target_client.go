package jobs

import (
	"net"
	"net/url"
	"strings"

	"ctf-platform/internal/model"
)

func resolveAWDHTTPDialOverride(accessURL, runtimeDetails string) (string, string) {
	parsed, err := url.Parse(strings.TrimSpace(accessURL))
	if err != nil {
		return "", ""
	}
	targetHost := strings.TrimSpace(parsed.Hostname())
	if targetHost == "" {
		return "", ""
	}

	resolved := model.ResolveRuntimeAliasAccessURL(accessURL, runtimeDetails)
	resolvedParsed, err := url.Parse(strings.TrimSpace(resolved))
	if err != nil {
		return "", ""
	}
	dialIP := strings.TrimSpace(resolvedParsed.Hostname())
	if dialIP == "" || strings.EqualFold(dialIP, targetHost) || net.ParseIP(dialIP) == nil {
		return "", ""
	}
	return targetHost, dialIP
}
