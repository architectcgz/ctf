package jobs

import (
	"context"
	"time"

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
		return buildAWDCheckOutcomeWithoutInstances(result)
	}

	return u.buildAWDCheckOutcomeFromProbes(ctx, instances, healthPath, result)
}
