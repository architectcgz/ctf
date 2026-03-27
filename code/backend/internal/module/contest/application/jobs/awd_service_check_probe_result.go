package jobs

import (
	"context"
	"encoding/json"

	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) buildAWDCheckOutcomeFromProbes(ctx context.Context, instances []contestports.AWDServiceInstance, healthPath string, result awdServiceCheckResult) (*awdServiceCheckOutcome, error) {
	aggregate := u.collectAWDProbeAggregate(ctx, instances, healthPath)
	status := applyAWDProbeAggregateResult(&result, aggregate, len(instances))

	raw, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	return &awdServiceCheckOutcome{
		serviceStatus: status,
		checkResult:   string(raw),
	}, nil
}
