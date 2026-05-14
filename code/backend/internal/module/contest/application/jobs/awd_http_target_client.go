package jobs

import (
	"context"
	"errors"
	"net"
	"net/url"
	"strings"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) executeAWDHTTPRequest(ctx context.Context, request contestports.AWDHTTPRequest) (contestports.AWDHTTPResponse, error) {
	if u == nil || u.httpRuntime == nil {
		return contestports.AWDHTTPResponse{}, newAWDCheckError("http_request_failed", "http_runtime_not_configured")
	}
	request.Timeout = normalizedAWDCheckerTimeout(request.Timeout)
	return u.httpRuntime.Execute(ctx, request)
}

func normalizeAWDHTTPRuntimeError(err error) (string, string) {
	var runtimeErr *contestports.AWDHTTPRuntimeError
	if errors.As(err, &runtimeErr) {
		errorCode := "http_request_failed"
		if runtimeErr.Kind == contestports.AWDHTTPRuntimeErrorKindResponseRead {
			errorCode = "http_response_read_failed"
		}
		if runtimeErr.Err != nil {
			return errorCode, sanitizeAWDCheckError(runtimeErr.Err)
		}
		return errorCode, sanitizeAWDCheckError(runtimeErr)
	}
	return normalizeAWDCheckError(err, "http_request_failed")
}

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
