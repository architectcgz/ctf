package jobs

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) buildAWDCheckOutcomeFromHTTPStandard(
	ctx context.Context,
	contestID int64,
	round *model.AWDRound,
	teamID int64,
	definition contestports.AWDServiceDefinition,
	instances []contestports.AWDServiceInstance,
	source string,
) (*awdServiceCheckOutcome, error) {
	config, err := parseAWDHTTPCheckerConfig(definition.CheckerConfig)
	result := awdServiceCheckResult{
		CheckedAt:            time.Now().UTC().Format(time.RFC3339),
		CheckSource:          contestdomain.NormalizeAWDCheckSource(source),
		CheckerType:          model.AWDCheckerTypeHTTPStandard,
		HealthPath:           resolveAWDCheckerHealthPath(definition.CheckerConfig, u.cfg.CheckerHealthPath),
		InstanceCount:        len(instances),
		HealthyInstanceCount: 0,
		FailedInstanceCount:  len(instances),
	}
	if err != nil {
		return buildAWDDownCheckOutcome(result, "invalid_checker_config", sanitizeAWDCheckError(err))
	}
	if len(instances) == 0 {
		return buildAWDCheckOutcomeWithoutInstances(result)
	}

	acceptedFlags := []string{}
	currentFlag := ""
	if needsAWDHTTPCheckerFlags(config) {
		acceptedFlags, err = u.resolveAcceptedRoundFlags(ctx, contestID, round, teamID, definition, time.Now().UTC())
		if err != nil {
			return buildAWDDownCheckOutcome(result, "flag_unavailable", sanitizeAWDCheckError(err))
		}
		if len(acceptedFlags) > 0 {
			currentFlag = acceptedFlags[0]
		}
	}
	checkerToken := ""
	if needsAWDHTTPCheckerToken(config) {
		checkerToken, err = u.resolveAWDCheckerToken(definition, contestID, teamID)
		if err != nil {
			return buildAWDDownCheckOutcome(result, "checker_token_unavailable", sanitizeAWDCheckError(err))
		}
	}

	targets := make([]awdHTTPCheckerTargetRuntimeResult, 0, len(instances))
	for _, instance := range instances {
		targets = append(targets, u.runAWDHTTPCheckerTarget(ctx, instance, round, teamID, definition, config, currentFlag, checkerToken, acceptedFlags))
	}

	status := applyAWDHTTPCheckerAggregateResult(&result, targets)
	return buildAWDCheckOutcome(result, status)
}

func (u *AWDRoundUpdater) runAWDHTTPCheckerTarget(
	ctx context.Context,
	instance contestports.AWDServiceInstance,
	round *model.AWDRound,
	teamID int64,
	definition contestports.AWDServiceDefinition,
	config awdHTTPCheckerConfig,
	currentFlag string,
	checkerToken string,
	acceptedFlags []string,
) awdHTTPCheckerTargetRuntimeResult {
	startedAt := time.Now()
	target := awdCheckTargetResult{
		AccessURL: instance.AccessURL,
		Probe:     string(model.AWDCheckerTypeHTTPStandard),
	}

	templateData := awdHTTPCheckerTemplateData{
		Flag:           currentFlag,
		CheckerToken:   strings.TrimSpace(checkerToken),
		Round:          0,
		TeamID:         teamID,
		AWDChallengeID: definition.AWDChallengeID,
	}
	if round != nil {
		templateData.Round = round.RoundNumber
	}

	if awdHTTPCheckerActionEnabled(config.PutFlag) {
		putResult := u.runAWDHTTPCheckerAction(ctx, instance.AccessURL, instance.RuntimeDetails, config.PutFlag, templateData, nil)
		target.PutFlag = putResult.summary
		if !putResult.summary.Healthy {
			target.ErrorCode = putResult.summary.ErrorCode
			target.Error = putResult.summary.Error
			target.LatencyMS = time.Since(startedAt).Milliseconds()
			return awdHTTPCheckerTargetRuntimeResult{
				status:       model.AWDServiceStatusDown,
				statusReason: putResult.summary.ErrorCode,
				target:       target,
			}
		}
	}

	getAction := config.GetFlag
	if !awdHTTPCheckerActionEnabled(getAction) {
		getAction.Path = normalizedAWDCheckerHealthPath(u.cfg.CheckerHealthPath)
	}

	expectedSubstrings := renderAWDHTTPCheckerExpectedSubstrings(getAction.ExpectedSubstring, templateData, acceptedFlags)
	getResult := u.runAWDHTTPCheckerAction(ctx, instance.AccessURL, instance.RuntimeDetails, getAction, templateData, expectedSubstrings)
	target.GetFlag = getResult.summary
	if !getResult.summary.Healthy {
		target.ErrorCode = getResult.summary.ErrorCode
		target.Error = getResult.summary.Error
		target.LatencyMS = time.Since(startedAt).Milliseconds()
		status := model.AWDServiceStatusDown
		if getResult.summary.ErrorCode == "flag_mismatch" {
			status = model.AWDServiceStatusCompromised
		}
		return awdHTTPCheckerTargetRuntimeResult{
			status:       status,
			statusReason: getResult.summary.ErrorCode,
			target:       target,
		}
	}

	if awdHTTPCheckerActionEnabled(config.Havoc) {
		havocResult := u.runAWDHTTPCheckerAction(ctx, instance.AccessURL, instance.RuntimeDetails, config.Havoc, templateData, nil)
		target.Havoc = havocResult.summary
		if !havocResult.summary.Healthy {
			target.ErrorCode = havocResult.summary.ErrorCode
			target.Error = havocResult.summary.Error
			target.LatencyMS = time.Since(startedAt).Milliseconds()
			return awdHTTPCheckerTargetRuntimeResult{
				status:       model.AWDServiceStatusDown,
				statusReason: havocResult.summary.ErrorCode,
				target:       target,
			}
		}
	}

	target.Healthy = true
	target.LatencyMS = time.Since(startedAt).Milliseconds()
	return awdHTTPCheckerTargetRuntimeResult{
		status:       model.AWDServiceStatusUp,
		statusReason: "healthy",
		target:       target,
	}
}

func applyAWDHTTPCheckerAggregateResult(result *awdServiceCheckResult, targets []awdHTTPCheckerTargetRuntimeResult) string {
	result.Targets = make([]awdCheckTargetResult, 0, len(targets))
	var firstHealthy *awdHTTPCheckerTargetRuntimeResult
	var firstCompromised *awdHTTPCheckerTargetRuntimeResult
	var firstDown *awdHTTPCheckerTargetRuntimeResult
	healthyCount := 0

	for i := range targets {
		item := targets[i]
		result.Targets = append(result.Targets, item.target)
		switch item.status {
		case model.AWDServiceStatusUp:
			healthyCount++
			if firstHealthy == nil {
				firstHealthy = &targets[i]
			}
		case model.AWDServiceStatusCompromised:
			if firstCompromised == nil {
				firstCompromised = &targets[i]
			}
		default:
			if firstDown == nil {
				firstDown = &targets[i]
			}
		}
	}

	result.HealthyInstanceCount = healthyCount
	result.FailedInstanceCount = len(targets) - healthyCount

	selected := firstHealthy
	if selected == nil {
		selected = firstCompromised
	}
	if selected == nil {
		selected = firstDown
	}
	if selected != nil {
		result.PutFlag = selected.target.PutFlag
		result.GetFlag = selected.target.GetFlag
		result.Havoc = selected.target.Havoc
		result.Probe = selected.target.Probe
		result.LatencyMS = selected.target.LatencyMS
	}

	if healthyCount > 0 {
		if healthyCount == len(targets) {
			result.StatusReason = "healthy"
		} else {
			result.StatusReason = "partial_available"
		}
		return model.AWDServiceStatusUp
	}

	if firstCompromised != nil {
		result.StatusReason = firstCompromised.statusReason
		result.ErrorCode = firstCompromised.target.ErrorCode
		result.Error = firstCompromised.target.Error
		return model.AWDServiceStatusCompromised
	}

	if firstDown != nil {
		result.StatusReason = firstDown.statusReason
		result.ErrorCode = firstDown.target.ErrorCode
		result.Error = firstDown.target.Error
	} else {
		result.StatusReason = "all_probes_failed"
	}
	return model.AWDServiceStatusDown
}

func renderAWDHTTPCheckerExpectedSubstrings(templateValue string, templateData awdHTTPCheckerTemplateData, acceptedFlags []string) []string {
	if strings.TrimSpace(templateValue) == "" {
		return nil
	}
	if len(acceptedFlags) == 0 {
		return []string{renderAWDHTTPCheckerTemplate(templateValue, templateData)}
	}

	values := make([]string, 0, len(acceptedFlags))
	for _, flag := range acceptedFlags {
		current := templateData
		current.Flag = flag
		values = append(values, renderAWDHTTPCheckerTemplate(templateValue, current))
	}
	return values
}

func needsAWDHTTPCheckerFlags(config awdHTTPCheckerConfig) bool {
	if awdHTTPCheckerActionEnabled(config.PutFlag) {
		return true
	}
	return strings.Contains(config.GetFlag.ExpectedSubstring, "{{FLAG}}") ||
		strings.Contains(config.PutFlag.BodyTemplate, "{{FLAG}}") ||
		strings.Contains(config.GetFlag.BodyTemplate, "{{FLAG}}") ||
		strings.Contains(config.Havoc.BodyTemplate, "{{FLAG}}")
}

func needsAWDHTTPCheckerToken(config awdHTTPCheckerConfig) bool {
	return actionUsesAWDHTTPCheckerToken(config.PutFlag) ||
		actionUsesAWDHTTPCheckerToken(config.GetFlag) ||
		actionUsesAWDHTTPCheckerToken(config.Havoc)
}

func actionUsesAWDHTTPCheckerToken(action awdHTTPCheckerActionConfig) bool {
	if strings.Contains(action.BodyTemplate, "{{CHECKER_TOKEN}}") ||
		strings.Contains(action.ExpectedSubstring, "{{CHECKER_TOKEN}}") {
		return true
	}
	for _, value := range action.Headers {
		if strings.Contains(value, "{{CHECKER_TOKEN}}") {
			return true
		}
	}
	return false
}
