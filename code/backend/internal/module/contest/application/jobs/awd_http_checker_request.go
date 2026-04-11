package jobs

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (u *AWDRoundUpdater) runAWDHTTPCheckerAction(
	ctx context.Context,
	accessURL string,
	action awdHTTPCheckerActionConfig,
	templateData awdHTTPCheckerTemplateData,
	expectedSubstrings []string,
) awdHTTPCheckerActionRuntimeResult {
	summary := &awdCheckActionResult{
		Healthy: false,
		Method:  action.Method,
		Path:    action.Path,
	}
	if !awdHTTPCheckerActionEnabled(action) {
		summary.ErrorCode = "checker_action_not_configured"
		summary.Error = "checker_action_not_configured"
		return awdHTTPCheckerActionRuntimeResult{summary: summary}
	}

	targetURL, err := buildAWDHealthCheckURL(accessURL, action.Path)
	if err != nil {
		errorCode, errorMessage := normalizeAWDCheckError(err, "invalid_access_url")
		summary.ErrorCode = errorCode
		summary.Error = errorMessage
		return awdHTTPCheckerActionRuntimeResult{summary: summary}
	}

	bodyValue := renderAWDHTTPCheckerTemplate(action.BodyTemplate, templateData)
	reqCtx, cancel := context.WithTimeout(ctx, normalizedAWDCheckerTimeout(u.cfg.CheckerTimeout))
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, action.Method, targetURL, bytes.NewBufferString(bodyValue))
	if err != nil {
		summary.ErrorCode = "http_request_failed"
		summary.Error = sanitizeAWDCheckError(err)
		return awdHTTPCheckerActionRuntimeResult{summary: summary}
	}
	for key, value := range action.Headers {
		req.Header.Set(key, renderAWDHTTPCheckerTemplate(value, templateData))
	}

	client := u.httpClient
	if client == nil {
		client = &http.Client{Timeout: normalizedAWDCheckerTimeout(u.cfg.CheckerTimeout)}
	}

	resp, err := client.Do(req)
	if err != nil {
		summary.ErrorCode = "http_request_failed"
		summary.Error = sanitizeAWDCheckError(err)
		return awdHTTPCheckerActionRuntimeResult{summary: summary}
	}
	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		summary.StatusCode = resp.StatusCode
		summary.ErrorCode = "http_response_read_failed"
		summary.Error = sanitizeAWDCheckError(readErr)
		return awdHTTPCheckerActionRuntimeResult{summary: summary}
	}

	summary.StatusCode = resp.StatusCode
	if resp.StatusCode != action.ExpectedStatus {
		summary.ErrorCode = "unexpected_http_status"
		summary.Error = fmt.Sprintf("unexpected_http_status:%d", resp.StatusCode)
		return awdHTTPCheckerActionRuntimeResult{summary: summary, responseBody: string(bodyBytes)}
	}

	if len(expectedSubstrings) > 0 && !containsAnyAWDExpectedSubstring(string(bodyBytes), expectedSubstrings) {
		summary.ErrorCode = "flag_mismatch"
		summary.Error = "flag_mismatch"
		return awdHTTPCheckerActionRuntimeResult{summary: summary, responseBody: string(bodyBytes)}
	}

	summary.Healthy = true
	return awdHTTPCheckerActionRuntimeResult{summary: summary, responseBody: string(bodyBytes)}
}

func containsAnyAWDExpectedSubstring(body string, expectedSubstrings []string) bool {
	for _, item := range expectedSubstrings {
		if strings.TrimSpace(item) != "" && strings.Contains(body, item) {
			return true
		}
	}
	return false
}
