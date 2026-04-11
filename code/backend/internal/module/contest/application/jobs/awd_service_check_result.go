package jobs

import (
	"context"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) checkTeamChallengeServices(
	ctx context.Context,
	definition contestports.AWDServiceDefinition,
	instances []contestports.AWDServiceInstance,
	round *model.AWDRound,
	source string,
) (*awdServiceCheckOutcome, error) {
	checkerType := effectiveAWDCheckerType(definition.CheckerType)
	healthPath := resolveAWDCheckerHealthPath(definition.CheckerConfig, u.cfg.CheckerHealthPath)
	result := awdServiceCheckResult{
		CheckedAt:            time.Now().UTC().Format(time.RFC3339),
		CheckSource:          contestdomain.NormalizeAWDCheckSource(source),
		CheckerType:          checkerType,
		HealthPath:           healthPath,
		InstanceCount:        len(instances),
		HealthyInstanceCount: 0,
		FailedInstanceCount:  len(instances),
	}
	if len(instances) == 0 {
		outcome, err := buildAWDCheckOutcomeWithoutInstances(result)
		if err != nil {
			return nil, err
		}
		outcome.checkerType = checkerType
		return outcome, nil
	}

	outcome, err := u.buildAWDCheckOutcomeFromProbes(ctx, instances, healthPath, result)
	if err != nil {
		return nil, err
	}
	outcome.checkerType = checkerType
	if outcome.serviceStatus == model.AWDServiceStatusUp {
		outcome.slaScore = definition.SLAScore
		outcome.defenseScore = effectiveAWDDefenseScore(definition, round)
	}
	return outcome, nil
}
