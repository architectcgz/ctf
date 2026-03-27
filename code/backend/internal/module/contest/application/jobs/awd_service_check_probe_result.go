package jobs

import (
	"context"
	"encoding/json"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) buildAWDCheckOutcomeFromProbes(ctx context.Context, instances []contestports.AWDServiceInstance, healthPath string, result awdServiceCheckResult) (*awdServiceCheckOutcome, error) {
	healthyCount := 0
	bestLatency := int64(0)
	bestProbe := ""
	targets := make([]awdCheckTargetResult, 0, len(instances))
	lastErrCode := ""
	firstErr := ""
	firstErrCode := ""

	for _, instance := range instances {
		probe := u.probeServiceInstance(ctx, instance.AccessURL, healthPath)
		target := awdCheckTargetResult{
			AccessURL: instance.AccessURL,
			Healthy:   probe.healthy,
			Probe:     probe.probe,
			LatencyMS: probe.latencyMS,
			ErrorCode: probe.errorCode,
			Error:     probe.err,
			Attempts:  probe.attempts,
		}
		targets = append(targets, target)
		if probe.healthy {
			healthyCount++
			if bestLatency == 0 || probe.latencyMS < bestLatency {
				bestLatency = probe.latencyMS
				bestProbe = probe.probe
			}
			continue
		}
		lastErrCode = probe.errorCode
		if firstErr == "" {
			firstErr = probe.err
			firstErrCode = probe.errorCode
		}
	}

	result.Targets = targets
	result.HealthyInstanceCount = healthyCount
	result.FailedInstanceCount = len(instances) - healthyCount
	if bestLatency > 0 {
		result.LatencyMS = bestLatency
	}
	if bestProbe != "" {
		result.Probe = bestProbe
	}

	status := model.AWDServiceStatusDown
	if healthyCount > 0 {
		status = model.AWDServiceStatusUp
		if healthyCount == len(instances) {
			result.StatusReason = "healthy"
		} else {
			result.StatusReason = "partial_available"
		}
	} else {
		if firstErrCode != "" {
			result.ErrorCode = firstErrCode
		} else if lastErrCode != "" {
			result.ErrorCode = lastErrCode
		}
		if firstErr != "" {
			result.Error = firstErr
		}
		if result.ErrorCode != "" {
			result.StatusReason = result.ErrorCode
		} else {
			result.StatusReason = "all_probes_failed"
		}
	}

	raw, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &awdServiceCheckOutcome{
		serviceStatus: status,
		checkResult:   string(raw),
	}, nil
}
