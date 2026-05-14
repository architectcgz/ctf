package jobs

import (
	"context"
	"fmt"
	"strings"

	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) runAWDHTTPCheckerAction(
	ctx context.Context,
	accessURL string,
	runtimeDetails string,
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
	if response, ok := u.runAWDHTTPCheckerActionInSandbox(ctx, targetURL, accessURL, runtimeDetails, action, bodyValue); ok {
		if response.Error != "" {
			summary.ErrorCode = "http_request_failed"
			summary.Error = response.Error
			return awdHTTPCheckerActionRuntimeResult{summary: summary}
		}
		summary.StatusCode = response.StatusCode
		if response.StatusCode != action.ExpectedStatus {
			summary.ErrorCode = "unexpected_http_status"
			summary.Error = fmt.Sprintf("unexpected_http_status:%d", response.StatusCode)
			return awdHTTPCheckerActionRuntimeResult{summary: summary, responseBody: response.Body}
		}
		if len(expectedSubstrings) > 0 && !containsAnyAWDExpectedSubstring(response.Body, expectedSubstrings) {
			summary.ErrorCode = "flag_mismatch"
			summary.Error = "flag_mismatch"
			return awdHTTPCheckerActionRuntimeResult{summary: summary, responseBody: response.Body}
		}
		summary.Healthy = true
		return awdHTTPCheckerActionRuntimeResult{summary: summary, responseBody: response.Body}
	}

	headers := make(map[string]string, len(action.Headers))
	for key, value := range action.Headers {
		headers[key] = renderAWDHTTPCheckerTemplate(value, templateData)
	}

	resp, err := u.executeAWDHTTPRequest(ctx, contestports.AWDHTTPRequest{
		AccessURL:      accessURL,
		RuntimeDetails: runtimeDetails,
		URL:            targetURL,
		Method:         action.Method,
		Headers:        headers,
		Body:           bodyValue,
		ReadBody:       true,
		Timeout:        u.cfg.CheckerTimeout,
	})
	if err != nil {
		summary.ErrorCode, summary.Error = normalizeAWDHTTPRuntimeError(err)
		return awdHTTPCheckerActionRuntimeResult{summary: summary}
	}

	summary.StatusCode = resp.StatusCode
	if resp.StatusCode != action.ExpectedStatus {
		summary.ErrorCode = "unexpected_http_status"
		summary.Error = fmt.Sprintf("unexpected_http_status:%d", resp.StatusCode)
		return awdHTTPCheckerActionRuntimeResult{summary: summary, responseBody: resp.Body}
	}

	if len(expectedSubstrings) > 0 && !containsAnyAWDExpectedSubstring(resp.Body, expectedSubstrings) {
		summary.ErrorCode = "flag_mismatch"
		summary.Error = "flag_mismatch"
		return awdHTTPCheckerActionRuntimeResult{summary: summary, responseBody: resp.Body}
	}

	summary.Healthy = true
	return awdHTTPCheckerActionRuntimeResult{summary: summary, responseBody: resp.Body}
}

func containsAnyAWDExpectedSubstring(body string, expectedSubstrings []string) bool {
	for _, item := range expectedSubstrings {
		if strings.TrimSpace(item) != "" && strings.Contains(body, item) {
			return true
		}
	}
	return false
}
