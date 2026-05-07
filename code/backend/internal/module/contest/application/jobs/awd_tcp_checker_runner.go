package jobs

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (u *AWDRoundUpdater) buildAWDPreviewOutcomeFromTCPStandard(
	ctx context.Context,
	definition contestports.AWDServiceDefinition,
	instances []contestports.AWDServiceInstance,
	previewContext contestports.AWDCheckerPreviewContext,
) (*awdServiceCheckOutcome, error) {
	return u.buildAWDCheckOutcomeFromTCPStandard(ctx, 0, nil, previewContext.TeamID, definition, instances, awdCheckSourceCheckerPreview, previewContext.PreviewFlag)
}

func (u *AWDRoundUpdater) buildAWDCheckOutcomeFromTCPStandard(
	ctx context.Context,
	contestID int64,
	round *model.AWDRound,
	teamID int64,
	definition contestports.AWDServiceDefinition,
	instances []contestports.AWDServiceInstance,
	source string,
	flag string,
) (*awdServiceCheckOutcome, error) {
	config, err := parseAWDTCPCheckerConfig(definition.CheckerConfig)
	result := awdServiceCheckResult{
		CheckedAt:            time.Now().UTC().Format(time.RFC3339),
		CheckSource:          contestdomain.NormalizeAWDCheckSource(source),
		CheckerType:          model.AWDCheckerTypeTCPStandard,
		InstanceCount:        len(instances),
		HealthyInstanceCount: 0,
		FailedInstanceCount:  len(instances),
	}
	if source == awdCheckSourceCheckerPreview {
		result.CheckSource = awdCheckSourceCheckerPreview
	}
	if err != nil {
		return buildAWDDownCheckOutcome(result, "invalid_checker_config", sanitizeAWDCheckError(err))
	}
	if source != awdCheckSourceCheckerPreview && flag == "" && needsAWDTCPCheckerFlags(config) {
		flag, err = u.resolveRoundFlag(ctx, contestID, round, teamID, definition)
		if err != nil {
			return buildAWDDownCheckOutcome(result, "flag_unavailable", sanitizeAWDCheckError(err))
		}
	}
	if len(instances) == 0 {
		return buildAWDCheckOutcomeWithoutInstances(result)
	}
	checkerToken, err := u.resolveAWDCheckerToken(definition, contestID, teamID)
	if err != nil {
		return buildAWDDownCheckOutcome(result, "checker_token_unavailable", sanitizeAWDCheckError(err))
	}

	targets := make([]awdCheckTargetResult, 0, len(instances))
	status := model.AWDServiceStatusUp
	statusReason := "healthy"
	for _, instance := range instances {
		target, targetStatus, reason := u.runAWDTCPCheckerTarget(ctx, config, contestID, round, teamID, definition, instance, flag, checkerToken)
		targets = append(targets, target)
		if targetStatus != model.AWDServiceStatusUp && status == model.AWDServiceStatusUp {
			status = targetStatus
			statusReason = reason
		}
	}

	healthyCount := 0
	for _, target := range targets {
		if target.Healthy {
			healthyCount++
		}
	}
	result.Targets = targets
	result.HealthyInstanceCount = healthyCount
	result.FailedInstanceCount = len(targets) - healthyCount
	result.StatusReason = statusReason
	if status != model.AWDServiceStatusUp {
		result.ErrorCode = statusReason
		result.Error = statusReason
	}
	return buildAWDCheckOutcome(result, status)
}

func needsAWDTCPCheckerFlags(config awdTCPCheckerConfig) bool {
	if strings.Contains(config.Connect.Host, "{{FLAG}}") || strings.Contains(awdTCPCheckerPortString(config.Connect.Port), "{{FLAG}}") {
		return true
	}
	for _, step := range append(append([]awdTCPCheckerStepConfig{}, config.Steps...), config.Havoc...) {
		if strings.Contains(step.SendTemplate, "{{FLAG}}") || strings.Contains(step.ExpectContains, "{{FLAG}}") || strings.Contains(step.ExpectRegex, "{{FLAG}}") {
			return true
		}
	}
	return false
}

func (u *AWDRoundUpdater) runAWDTCPCheckerTarget(
	ctx context.Context,
	config awdTCPCheckerConfig,
	contestID int64,
	round *model.AWDRound,
	teamID int64,
	definition contestports.AWDServiceDefinition,
	instance contestports.AWDServiceInstance,
	flag string,
	checkerToken string,
) (awdCheckTargetResult, string, string) {
	startedAt := time.Now()
	target := awdCheckTargetResult{
		AccessURL: instance.AccessURL,
		Probe:     string(model.AWDCheckerTypeTCPStandard),
	}
	auditJob := contestports.CheckerRunJob{
		Metadata: contestports.CheckerRunMetadata{
			ContestID:   contestID,
			ServiceID:   definition.ServiceID,
			TeamID:      teamID,
			RoundNumber: awdScriptRoundNumber(round),
		},
	}
	timeout := config.timeout(u.cfg.CheckerTimeout)
	address, err := resolveAWDTCPCheckerAddress(config, instance, definition, round, teamID, flag, checkerToken)
	if err != nil {
		target.ErrorCode = "invalid_access_url"
		target.Error = sanitizeAWDCheckErrorWithSecrets(err, flag, checkerToken)
		target.Audit = buildAWDCheckerAuditRecord(auditJob, model.AWDCheckerTypeTCPStandard, "", contestports.CheckerRunResult{Duration: time.Since(startedAt)}, target.ErrorCode, flag, checkerToken)
		return target, model.AWDServiceStatusDown, target.ErrorCode
	}

	dialer := net.Dialer{Timeout: timeout}
	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	conn, err := dialer.DialContext(runCtx, "tcp", address)
	if err != nil {
		target.ErrorCode = "tcp_connect_failed"
		target.Error = sanitizeAWDCheckErrorWithSecrets(err, flag, checkerToken)
		target.Audit = buildAWDCheckerAuditRecord(auditJob, model.AWDCheckerTypeTCPStandard, "", contestports.CheckerRunResult{Duration: time.Since(startedAt)}, target.ErrorCode, flag, checkerToken)
		return target, model.AWDServiceStatusDown, target.ErrorCode
	}
	defer conn.Close()

	for _, step := range append(append([]awdTCPCheckerStepConfig{}, config.Steps...), config.Havoc...) {
		if err := runAWDTCPCheckerStep(conn, step, timeout, instance, definition, round, teamID, flag, checkerToken); err != nil {
			code, message := normalizeAWDCheckError(err, "tcp_step_failed")
			target.ErrorCode = code
			target.Error = sanitizeAWDCheckerText(message, flag, checkerToken)
			target.LatencyMS = time.Since(startedAt).Milliseconds()
			target.Audit = buildAWDCheckerAuditRecord(auditJob, model.AWDCheckerTypeTCPStandard, "", contestports.CheckerRunResult{Duration: time.Since(startedAt)}, code, flag, checkerToken)
			return target, model.AWDServiceStatusDown, code
		}
	}

	target.Healthy = true
	target.LatencyMS = time.Since(startedAt).Milliseconds()
	target.Audit = buildAWDCheckerAuditRecord(auditJob, model.AWDCheckerTypeTCPStandard, "", contestports.CheckerRunResult{Duration: time.Since(startedAt)}, "healthy", flag, checkerToken)
	return target, model.AWDServiceStatusUp, "healthy"
}

func resolveAWDTCPCheckerAddress(config awdTCPCheckerConfig, instance contestports.AWDServiceInstance, definition contestports.AWDServiceDefinition, round *model.AWDRound, teamID int64, flag string, checkerToken string) (string, error) {
	host, port := splitAWDScriptCheckerTarget(instance.AccessURL)
	if configuredHost := strings.TrimSpace(config.Connect.Host); configuredHost != "" {
		host = renderAWDScriptCheckerValue(configuredHost, instance, definition, round, teamID, flag, checkerToken)
	}
	if configuredPort := awdTCPCheckerPortString(config.Connect.Port); configuredPort != "" {
		port = renderAWDScriptCheckerValue(configuredPort, instance, definition, round, teamID, flag, checkerToken)
	}
	if host == "" || port == "" {
		return "", fmt.Errorf("tcp checker target host or port is empty")
	}
	return net.JoinHostPort(host, port), nil
}

func awdTCPCheckerPortString(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return strings.TrimSpace(typed)
	case float64:
		return strconv.Itoa(int(typed))
	default:
		return strings.TrimSpace(fmt.Sprint(typed))
	}
}

func runAWDTCPCheckerStep(conn net.Conn, step awdTCPCheckerStepConfig, defaultTimeout time.Duration, instance contestports.AWDServiceInstance, definition contestports.AWDServiceDefinition, round *model.AWDRound, teamID int64, flag string, checkerToken string) error {
	timeout := step.timeout(defaultTimeout)
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return newAWDCheckError("tcp_deadline_failed", sanitizeAWDCheckError(err))
	}
	payload, err := renderAWDTCPCheckerPayload(step, instance, definition, round, teamID, flag, checkerToken)
	if err != nil {
		return err
	}
	if len(payload) > 0 {
		if _, err := conn.Write(payload); err != nil {
			return newAWDCheckError("tcp_send_failed", sanitizeAWDCheckError(err))
		}
	}
	if strings.TrimSpace(step.ExpectContains) == "" && strings.TrimSpace(step.ExpectRegex) == "" {
		return nil
	}
	return readAWDTCPCheckerExpectation(conn, step, instance, definition, round, teamID, flag, checkerToken)
}

func renderAWDTCPCheckerPayload(step awdTCPCheckerStepConfig, instance contestports.AWDServiceInstance, definition contestports.AWDServiceDefinition, round *model.AWDRound, teamID int64, flag string, checkerToken string) ([]byte, error) {
	if strings.TrimSpace(step.SendHex) != "" {
		payload, err := hex.DecodeString(strings.ReplaceAll(strings.TrimSpace(step.SendHex), " ", ""))
		if err != nil {
			return nil, newAWDCheckError("invalid_tcp_payload", sanitizeAWDCheckError(err))
		}
		return payload, nil
	}
	if step.SendTemplate != "" {
		return []byte(renderAWDScriptCheckerValue(step.SendTemplate, instance, definition, round, teamID, flag, checkerToken)), nil
	}
	return []byte(step.Send), nil
}

func readAWDTCPCheckerExpectation(conn net.Conn, step awdTCPCheckerStepConfig, instance contestports.AWDServiceInstance, definition contestports.AWDServiceDefinition, round *model.AWDRound, teamID int64, flag string, checkerToken string) error {
	expectedContains := renderAWDScriptCheckerValue(step.ExpectContains, instance, definition, round, teamID, flag, checkerToken)
	expectedRegex := renderAWDScriptCheckerValue(step.ExpectRegex, instance, definition, round, teamID, flag, checkerToken)
	var compiled *regexp.Regexp
	if strings.TrimSpace(expectedRegex) != "" {
		regex, err := regexp.Compile(expectedRegex)
		if err != nil {
			return newAWDCheckError("invalid_tcp_expectation", sanitizeAWDCheckError(err))
		}
		compiled = regex
	}
	buffer := make([]byte, 0, 4096)
	chunk := make([]byte, 512)
	for len(buffer) < 64*1024 {
		n, err := conn.Read(chunk)
		if n > 0 {
			buffer = append(buffer, chunk[:n]...)
			text := string(buffer)
			if expectedContains != "" && strings.Contains(text, expectedContains) {
				return nil
			}
			if compiled != nil && compiled.Match(buffer) {
				return nil
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return newAWDCheckError("tcp_read_failed", sanitizeAWDCheckError(err))
		}
	}
	return newAWDCheckError("tcp_expectation_failed", "tcp_expectation_failed")
}
