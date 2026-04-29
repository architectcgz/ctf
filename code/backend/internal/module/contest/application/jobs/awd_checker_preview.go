package jobs

import (
	"context"
	"strings"
	"time"

	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
)

const (
	awdCheckSourceCheckerPreview = "checker_preview"
	defaultAWDCheckerPreviewFlag = "flag{preview}"
)

func (u *AWDRoundUpdater) PreviewServiceCheck(ctx context.Context, req contestports.AWDServicePreviewRequest) (*contestports.AWDServicePreviewResult, error) {
	checkerType := effectiveAWDCheckerType(req.CheckerType)
	previewContext := contestports.AWDCheckerPreviewContext{
		ServiceID:      req.ServiceID,
		AccessURL:      strings.TrimSpace(req.AccessURL),
		PreviewFlag:    normalizeAWDCheckerPreviewFlag(req.PreviewFlag),
		RoundNumber:    0,
		TeamID:         0,
		AWDChallengeID: req.AWDChallengeID,
	}
	definition := contestports.AWDServiceDefinition{
		ServiceID:      req.ServiceID,
		AWDChallengeID: req.AWDChallengeID,
		CheckerType:    checkerType,
		CheckerConfig:  req.CheckerConfig,
	}
	instances := []contestports.AWDServiceInstance{
		{
			ServiceID:      req.ServiceID,
			AWDChallengeID: req.AWDChallengeID,
			AccessURL:      previewContext.AccessURL,
		},
	}

	outcome, err := u.previewServiceCheck(ctx, definition, instances, previewContext)
	if err != nil {
		return nil, err
	}
	outcome.checkerType = checkerType

	return &contestports.AWDServicePreviewResult{
		ServiceStatus:  previewContextServiceStatus(outcome),
		CheckerType:    checkerType,
		CheckResult:    outcome.checkResult,
		PreviewContext: previewContext,
	}, nil
}

func (u *AWDRoundUpdater) previewServiceCheck(
	ctx context.Context,
	definition contestports.AWDServiceDefinition,
	instances []contestports.AWDServiceInstance,
	previewContext contestports.AWDCheckerPreviewContext,
) (*awdServiceCheckOutcome, error) {
	checkerType := effectiveAWDCheckerType(definition.CheckerType)
	switch checkerType {
	case model.AWDCheckerTypeHTTPStandard:
		return u.buildAWDPreviewOutcomeFromHTTPStandard(ctx, definition, instances, previewContext)
	case model.AWDCheckerTypeTCPStandard:
		return u.buildAWDPreviewOutcomeFromTCPStandard(ctx, definition, instances, previewContext)
	case model.AWDCheckerTypeScript:
		return u.buildAWDPreviewOutcomeFromScriptChecker(ctx, definition, instances, previewContext)
	default:
		healthPath := resolveAWDCheckerHealthPath(definition.CheckerConfig, u.cfg.CheckerHealthPath)
		result := awdServiceCheckResult{
			CheckedAt:            time.Now().UTC().Format(time.RFC3339),
			CheckSource:          awdCheckSourceCheckerPreview,
			CheckerType:          checkerType,
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
}

func (u *AWDRoundUpdater) buildAWDPreviewOutcomeFromHTTPStandard(
	ctx context.Context,
	definition contestports.AWDServiceDefinition,
	instances []contestports.AWDServiceInstance,
	previewContext contestports.AWDCheckerPreviewContext,
) (*awdServiceCheckOutcome, error) {
	config, err := parseAWDHTTPCheckerConfig(definition.CheckerConfig)
	result := awdServiceCheckResult{
		CheckedAt:            time.Now().UTC().Format(time.RFC3339),
		CheckSource:          awdCheckSourceCheckerPreview,
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

	templateData := awdHTTPCheckerTemplateData{
		Flag:           previewContext.PreviewFlag,
		Round:          previewContext.RoundNumber,
		TeamID:         previewContext.TeamID,
		AWDChallengeID: previewContext.AWDChallengeID,
	}
	acceptedFlags := []string{previewContext.PreviewFlag}
	targets := make([]awdHTTPCheckerTargetRuntimeResult, 0, len(instances))
	for _, instance := range instances {
		targets = append(targets, u.runAWDHTTPCheckerPreviewTarget(ctx, instance, config, templateData, acceptedFlags))
	}

	status := applyAWDHTTPCheckerAggregateResult(&result, targets)
	return buildAWDCheckOutcome(result, status)
}

func (u *AWDRoundUpdater) runAWDHTTPCheckerPreviewTarget(
	ctx context.Context,
	instance contestports.AWDServiceInstance,
	config awdHTTPCheckerConfig,
	templateData awdHTTPCheckerTemplateData,
	acceptedFlags []string,
) awdHTTPCheckerTargetRuntimeResult {
	startedAt := time.Now()
	target := awdCheckTargetResult{
		AccessURL: instance.AccessURL,
		Probe:     string(model.AWDCheckerTypeHTTPStandard),
	}

	if awdHTTPCheckerActionEnabled(config.PutFlag) {
		putResult := u.runAWDHTTPCheckerAction(ctx, instance.AccessURL, config.PutFlag, templateData, nil)
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
	getResult := u.runAWDHTTPCheckerAction(ctx, instance.AccessURL, getAction, templateData, expectedSubstrings)
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
		havocResult := u.runAWDHTTPCheckerAction(ctx, instance.AccessURL, config.Havoc, templateData, nil)
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

func normalizeAWDCheckerPreviewFlag(value string) string {
	if strings.TrimSpace(value) == "" {
		return defaultAWDCheckerPreviewFlag
	}
	return strings.TrimSpace(value)
}

func previewContextServiceStatus(outcome *awdServiceCheckOutcome) string {
	if outcome == nil {
		return model.AWDServiceStatusDown
	}
	return outcome.serviceStatus
}
