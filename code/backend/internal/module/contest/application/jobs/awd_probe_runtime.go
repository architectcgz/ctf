package jobs

import (
	"context"
	"fmt"
	"time"

	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) probeServiceInstance(ctx context.Context, accessURL, runtimeDetails, healthPath string) awdInstanceProbeResult {
	startedAt := time.Now()
	attempts := make([]awdProbeAttemptResult, 0, 1)
	targetURL, err := buildAWDHealthCheckURL(accessURL, healthPath)
	if err == nil {
		if u.httpRuntime == nil {
			err = newAWDCheckError("http_request_failed", "http_runtime_unavailable")
		} else {
			response, execErr := u.httpRuntime.Execute(ctx, contestports.AWDHTTPRequest{
				AccessURL:      accessURL,
				RuntimeDetails: runtimeDetails,
				URL:            targetURL,
				Method:         "GET",
				ReadBody:       false,
				Timeout:        normalizedAWDCheckerTimeout(u.cfg.CheckerTimeout),
			})
			if execErr == nil {
				if response.StatusCode >= 200 && response.StatusCode < 400 {
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
				err = newAWDCheckError("unexpected_http_status", fmt.Sprintf("unexpected_http_status:%d", response.StatusCode))
			} else {
				err = newAWDCheckError("http_request_failed", sanitizeAWDCheckError(execErr))
			}
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
