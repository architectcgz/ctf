package jobs

import (
	"context"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

type awdProbeAggregate struct {
	healthyCount int
	bestLatency  int64
	bestProbe    string
	targets      []awdCheckTargetResult
	lastErrCode  string
	firstErr     string
	firstErrCode string
}

func (u *AWDRoundUpdater) collectAWDProbeAggregate(ctx context.Context, instances []contestports.AWDServiceInstance, healthPath string) awdProbeAggregate {
	aggregate := awdProbeAggregate{
		targets: make([]awdCheckTargetResult, 0, len(instances)),
	}

	for _, instance := range instances {
		probe := u.probeServiceInstance(ctx, instance.AccessURL, healthPath)
		aggregate.targets = append(aggregate.targets, awdCheckTargetResult{
			AccessURL: instance.AccessURL,
			Healthy:   probe.healthy,
			Probe:     probe.probe,
			LatencyMS: probe.latencyMS,
			ErrorCode: probe.errorCode,
			Error:     probe.err,
			Attempts:  probe.attempts,
		})

		if probe.healthy {
			aggregate.healthyCount++
			if aggregate.bestLatency == 0 || probe.latencyMS < aggregate.bestLatency {
				aggregate.bestLatency = probe.latencyMS
				aggregate.bestProbe = probe.probe
			}
			continue
		}

		aggregate.lastErrCode = probe.errorCode
		if aggregate.firstErr == "" {
			aggregate.firstErr = probe.err
			aggregate.firstErrCode = probe.errorCode
		}
	}

	return aggregate
}

func applyAWDProbeAggregateResult(result *awdServiceCheckResult, aggregate awdProbeAggregate, totalInstances int) string {
	result.Targets = aggregate.targets
	result.HealthyInstanceCount = aggregate.healthyCount
	result.FailedInstanceCount = totalInstances - aggregate.healthyCount
	if aggregate.bestLatency > 0 {
		result.LatencyMS = aggregate.bestLatency
	}
	if aggregate.bestProbe != "" {
		result.Probe = aggregate.bestProbe
	}

	if aggregate.healthyCount > 0 {
		if aggregate.healthyCount == totalInstances {
			result.StatusReason = "healthy"
		} else {
			result.StatusReason = "partial_available"
		}
		return model.AWDServiceStatusUp
	}

	if aggregate.firstErrCode != "" {
		result.ErrorCode = aggregate.firstErrCode
	} else if aggregate.lastErrCode != "" {
		result.ErrorCode = aggregate.lastErrCode
	}
	if aggregate.firstErr != "" {
		result.Error = aggregate.firstErr
	}
	if result.ErrorCode != "" {
		result.StatusReason = result.ErrorCode
	} else {
		result.StatusReason = "all_probes_failed"
	}
	return model.AWDServiceStatusDown
}
