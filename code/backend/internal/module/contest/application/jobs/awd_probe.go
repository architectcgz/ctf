package jobs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type awdProbeAttemptResult struct {
	Probe     string `json:"probe"`
	Healthy   bool   `json:"healthy"`
	LatencyMS int64  `json:"latency_ms,omitempty"`
	ErrorCode string `json:"error_code,omitempty"`
	Error     string `json:"error,omitempty"`
}

type awdInstanceProbeResult struct {
	healthy   bool
	latencyMS int64
	probe     string
	errorCode string
	err       string
	attempts  []awdProbeAttemptResult
}

type awdCheckError struct {
	code    string
	message string
}

func (e awdCheckError) Error() string {
	return e.message
}

func (u *AWDRoundUpdater) probeServiceInstance(ctx context.Context, accessURL, healthPath string) awdInstanceProbeResult {
	startedAt := time.Now()
	attempts := make([]awdProbeAttemptResult, 0, 1)
	targetURL, err := buildAWDHealthCheckURL(accessURL, healthPath)
	if err == nil {
		client := u.httpClient
		if client == nil {
			client = &http.Client{Timeout: normalizedAWDCheckerTimeout(u.cfg.CheckerTimeout)}
		}
		reqCtx, cancel := context.WithTimeout(ctx, normalizedAWDCheckerTimeout(u.cfg.CheckerTimeout))
		defer cancel()

		req, reqErr := http.NewRequestWithContext(reqCtx, http.MethodGet, targetURL, nil)
		if reqErr == nil {
			resp, doErr := client.Do(req)
			if doErr == nil {
				_, _ = io.Copy(io.Discard, resp.Body)
				_ = resp.Body.Close()
				if resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusBadRequest {
					attempts = append(attempts, awdProbeAttemptResult{
						Probe:     "http",
						Healthy:   true,
						LatencyMS: time.Since(startedAt).Milliseconds(),
					})
					return awdInstanceProbeResult{
						healthy:   true,
						latencyMS: time.Since(startedAt).Milliseconds(),
						probe:     "http",
						attempts:  attempts,
					}
				}
				err = newAWDCheckError("unexpected_http_status", fmt.Sprintf("unexpected_http_status:%d", resp.StatusCode))
			} else {
				err = newAWDCheckError("http_request_failed", sanitizeAWDCheckError(doErr))
			}
		} else {
			err = newAWDCheckError("http_request_failed", sanitizeAWDCheckError(reqErr))
		}
		errorCode, errorMessage := normalizeAWDCheckError(err, "http_request_failed")
		attempts = append(attempts, awdProbeAttemptResult{
			Probe:     "http",
			Healthy:   false,
			ErrorCode: errorCode,
			Error:     errorMessage,
		})
	} else {
		errorCode, errorMessage := normalizeAWDCheckError(err, "invalid_access_url")
		attempts = append(attempts, awdProbeAttemptResult{
			Probe:     "http",
			Healthy:   false,
			ErrorCode: errorCode,
			Error:     errorMessage,
		})
	}

	errorCode, errorMessage := normalizeAWDCheckError(err, "unknown_checker_error")
	if errorCode == "" && len(attempts) > 0 {
		lastAttempt := attempts[len(attempts)-1]
		errorCode = lastAttempt.ErrorCode
		errorMessage = lastAttempt.Error
	}
	return awdInstanceProbeResult{
		healthy:   false,
		probe:     "http",
		errorCode: errorCode,
		err:       errorMessage,
		attempts:  attempts,
	}
}
