package jobs

import (
	"context"
	"encoding/json"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) checkTeamChallengeServices(ctx context.Context, instances []contestports.AWDServiceInstance, source string) (*awdServiceCheckOutcome, error) {
	healthPath := normalizedAWDCheckerHealthPath(u.cfg.CheckerHealthPath)
	result := awdServiceCheckResult{
		CheckedAt:            time.Now().UTC().Format(time.RFC3339),
		CheckSource:          contestdomain.NormalizeAWDCheckSource(source),
		HealthPath:           healthPath,
		InstanceCount:        len(instances),
		HealthyInstanceCount: 0,
		FailedInstanceCount:  len(instances),
	}
	if len(instances) == 0 {
		result.StatusReason = "no_running_instances"
		result.ErrorCode = "no_running_instances"
		result.Error = "no_running_instances"
		raw, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		return &awdServiceCheckOutcome{
			serviceStatus: model.AWDServiceStatusDown,
			checkResult:   string(raw),
		}, nil
	}

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
