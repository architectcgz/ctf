package jobs

import (
	"context"

	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) buildAWDCheckOutcomeFromProbes(ctx context.Context, instances []contestports.AWDServiceInstance, healthPath string, result awdServiceCheckResult) (*awdServiceCheckOutcome, error) {
	aggregate := u.collectAWDProbeAggregate(ctx, instances, healthPath)
	status := applyAWDProbeAggregateResult(&result, aggregate, len(instances))
	return buildAWDCheckOutcome(result, status)
}
