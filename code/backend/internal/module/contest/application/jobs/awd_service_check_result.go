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
	contest *model.Contest,
	contestID int64,
	teamID int64,
	definition contestports.AWDServiceDefinition,
	instances []contestports.AWDServiceInstance,
	round *model.AWDRound,
	source string,
) (*awdServiceCheckOutcome, error) {
	checkerType := effectiveAWDCheckerType(definition.CheckerType)
	var (
		outcome *awdServiceCheckOutcome
		err     error
	)
	switch checkerType {
	case model.AWDCheckerTypeHTTPStandard:
		outcome, err = u.buildAWDCheckOutcomeFromHTTPStandard(ctx, contest, contestID, round, teamID, definition, instances, source)
	case model.AWDCheckerTypeTCPStandard:
		outcome, err = u.buildAWDCheckOutcomeFromTCPStandard(ctx, contestID, round, teamID, definition, instances, source, "")
	case model.AWDCheckerTypeScript:
		roundFlag, flagErr := u.resolveRoundFlag(ctx, contestID, round, teamID, definition)
		if flagErr != nil {
			result := awdServiceCheckResult{
				CheckedAt:            time.Now().UTC().Format(time.RFC3339),
				CheckSource:          contestdomain.NormalizeAWDCheckSource(source),
				CheckerType:          checkerType,
				InstanceCount:        len(instances),
				HealthyInstanceCount: 0,
				FailedInstanceCount:  len(instances),
			}
			outcome, err = buildAWDDownCheckOutcome(result, "flag_unavailable", sanitizeAWDCheckError(flagErr))
			break
		}
		outcome, err = u.buildAWDCheckOutcomeFromScriptChecker(ctx, contestID, round, teamID, definition, instances, source, roundFlag)
	default:
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
			outcome, err = buildAWDCheckOutcomeWithoutInstances(result)
		} else {
			outcome, err = u.buildAWDCheckOutcomeFromProbes(ctx, instances, healthPath, result)
		}
	}
	if err != nil {
		return nil, err
	}
	outcome.checkerType = checkerType
	if outcome.serviceStatus == model.AWDServiceStatusUp {
		outcome.slaScore = definition.SLAScore
		outcome.defenseScore = effectiveAWDDefenseScore(definition, round)
	} else if exempt, err := u.repo.HasSystemRecoveryOperationAt(ctx, contestID, teamID, definition.ServiceID, time.Now().UTC()); err != nil {
		return nil, err
	} else if exempt {
		outcome.slaScore = definition.SLAScore
		outcome.checkResult = annotateAWDRecoverySLAExemption(outcome.checkResult)
	}
	return outcome, nil
}
