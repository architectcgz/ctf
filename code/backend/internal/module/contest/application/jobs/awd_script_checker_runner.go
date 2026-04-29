package jobs

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
)

const defaultAWDCheckerArtifactRoot = "./data/awd-checker-artifacts"

func (u *AWDRoundUpdater) buildAWDPreviewOutcomeFromScriptChecker(
	ctx context.Context,
	definition contestports.AWDServiceDefinition,
	instances []contestports.AWDServiceInstance,
	previewContext contestports.AWDCheckerPreviewContext,
) (*awdServiceCheckOutcome, error) {
	return u.buildAWDCheckOutcomeFromScriptChecker(ctx, 0, nil, previewContext.TeamID, definition, instances, awdCheckSourceCheckerPreview, previewContext.PreviewFlag)
}

func (u *AWDRoundUpdater) buildAWDCheckOutcomeFromScriptChecker(
	ctx context.Context,
	contestID int64,
	round *model.AWDRound,
	teamID int64,
	definition contestports.AWDServiceDefinition,
	instances []contestports.AWDServiceInstance,
	source string,
	flag string,
) (*awdServiceCheckOutcome, error) {
	cfg, err := parseAWDScriptCheckerConfig(definition.CheckerConfig)
	checkSource := contestdomain.NormalizeAWDCheckSource(source)
	if source == awdCheckSourceCheckerPreview {
		checkSource = awdCheckSourceCheckerPreview
	}
	result := awdServiceCheckResult{
		CheckedAt:            time.Now().UTC().Format(time.RFC3339),
		CheckSource:          checkSource,
		CheckerType:          model.AWDCheckerTypeScript,
		InstanceCount:        len(instances),
		HealthyInstanceCount: 0,
		FailedInstanceCount:  len(instances),
	}
	if err != nil {
		return buildAWDDownCheckOutcome(result, "invalid_checker_config", sanitizeAWDCheckError(err))
	}
	if u.checkerRunner == nil {
		return buildAWDDownCheckOutcome(result, "checker_runner_unavailable", "checker_runner_unavailable")
	}
	if len(instances) == 0 {
		return buildAWDCheckOutcomeWithoutInstances(result)
	}

	targets := make([]awdCheckTargetResult, 0, len(instances))
	status := model.AWDServiceStatusUp
	statusReason := "healthy"
	for _, instance := range instances {
		target, targetStatus, reason := u.runAWDScriptCheckerTarget(ctx, cfg, contestID, round, teamID, definition, instance, flag)
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

func (u *AWDRoundUpdater) runAWDScriptCheckerTarget(
	ctx context.Context,
	cfg awdScriptCheckerConfig,
	contestID int64,
	round *model.AWDRound,
	teamID int64,
	definition contestports.AWDServiceDefinition,
	instance contestports.AWDServiceInstance,
	flag string,
) (awdCheckTargetResult, string, string) {
	startedAt := time.Now()
	target := awdCheckTargetResult{
		AccessURL: instance.AccessURL,
		Probe:     string(model.AWDCheckerTypeScript),
	}
	job := contestports.CheckerRunJob{
		Runtime:         cfg.Runtime,
		Entry:           cfg.Entry,
		Args:            renderAWDScriptCheckerValues(cfg.Args, instance, definition, round, teamID, flag),
		Env:             renderAWDScriptCheckerEnv(cfg.Env, instance, definition, round, teamID, flag),
		OutputMode:      cfg.Output,
		TargetAllowlist: buildAWDScriptCheckerTargetAllowlist(instance.AccessURL),
		Timeout:         cfg.timeout(u.cfg.CheckerSandbox.Timeout),
		Limits: contestports.CheckerRunLimits{
			CPUQuota:         u.cfg.CheckerSandbox.CPUQuota,
			MemoryBytes:      u.cfg.CheckerSandbox.MemoryBytes,
			PidsLimit:        u.cfg.CheckerSandbox.PidsLimit,
			NofileLimit:      u.cfg.CheckerSandbox.NofileLimit,
			OutputLimitBytes: u.cfg.CheckerSandbox.OutputLimitBytes,
		},
		Metadata: contestports.CheckerRunMetadata{
			ContestID:   contestID,
			ServiceID:   definition.ServiceID,
			TeamID:      teamID,
			RoundNumber: awdScriptRoundNumber(round),
		},
	}
	if file, ok, err := loadAWDScriptCheckerArtifact(cfg); err != nil {
		target.ErrorCode = "checker_artifact_unavailable"
		target.Error = sanitizeAWDCheckError(err)
		return target, model.AWDServiceStatusDown, target.ErrorCode
	} else if ok {
		job.Files = []contestports.CheckerRunFile{file}
	}

	runResult, err := u.checkerRunner.RunChecker(ctx, job)
	target.LatencyMS = time.Since(startedAt).Milliseconds()
	if err != nil {
		target.ErrorCode = "checker_runner_error"
		target.Error = sanitizeAWDCheckError(err)
		return target, model.AWDServiceStatusDown, target.ErrorCode
	}
	if runResult.Status != contestports.CheckerRunStatusOK {
		reason := strings.TrimSpace(runResult.Reason)
		if reason == "" {
			reason = contestports.CheckerReasonFailed
		}
		target.ErrorCode = reason
		target.Error = sanitizeAWDScriptCheckerError(runResult)
		return target, model.AWDServiceStatusDown, reason
	}
	target.Healthy = true
	return target, model.AWDServiceStatusUp, "healthy"
}

func loadAWDScriptCheckerArtifact(cfg awdScriptCheckerConfig) (contestports.CheckerRunFile, bool, error) {
	storagePath := strings.TrimSpace(cfg.Artifact.StoragePath)
	if storagePath == "" {
		return contestports.CheckerRunFile{}, false, nil
	}
	if err := validateAWDScriptCheckerArtifactPath(storagePath); err != nil {
		return contestports.CheckerRunFile{}, false, err
	}
	content, err := os.ReadFile(storagePath)
	if err != nil {
		return contestports.CheckerRunFile{}, false, err
	}
	if expected := strings.TrimSpace(cfg.Artifact.SHA256); expected != "" {
		sum := sha256.Sum256(content)
		actual := hex.EncodeToString(sum[:])
		if !strings.EqualFold(actual, expected) {
			return contestports.CheckerRunFile{}, false, fmt.Errorf("script checker artifact sha256 mismatch")
		}
	}
	if cfg.Artifact.Size > 0 && int64(len(content)) != cfg.Artifact.Size {
		return contestports.CheckerRunFile{}, false, fmt.Errorf("script checker artifact size mismatch")
	}
	entry := strings.TrimSpace(cfg.Artifact.Entry)
	if entry == "" {
		entry = cfg.Entry
	}
	if err := validateAWDScriptCheckerEntry(entry); err != nil {
		return contestports.CheckerRunFile{}, false, err
	}
	return contestports.CheckerRunFile{
		Path:    entry,
		Content: content,
		Mode:    0o500,
	}, true, nil
}

func validateAWDScriptCheckerArtifactPath(storagePath string) error {
	root, err := filepath.Abs(awdCheckerArtifactRoot())
	if err != nil {
		return err
	}
	target, err := filepath.Abs(storagePath)
	if err != nil {
		return err
	}
	if resolvedRoot, err := filepath.EvalSymlinks(root); err == nil {
		root = resolvedRoot
	}
	if resolvedTarget, err := filepath.EvalSymlinks(target); err == nil {
		target = resolvedTarget
	}
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return err
	}
	if rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." {
		return fmt.Errorf("script checker artifact path is outside artifact root")
	}
	return nil
}

func awdCheckerArtifactRoot() string {
	if value := strings.TrimSpace(os.Getenv("AWD_CHECKER_ARTIFACT_DIR")); value != "" {
		return value
	}
	return defaultAWDCheckerArtifactRoot
}

func renderAWDScriptCheckerEnv(env map[string]string, instance contestports.AWDServiceInstance, definition contestports.AWDServiceDefinition, round *model.AWDRound, teamID int64, flag string) map[string]string {
	rendered := map[string]string{
		"TARGET_URL":       strings.TrimSpace(instance.AccessURL),
		"FLAG":             strings.TrimSpace(flag),
		"TEAM_ID":          fmt.Sprintf("%d", teamID),
		"CHALLENGE_ID":     fmt.Sprintf("%d", definition.AWDChallengeID),
		"AWD_CHALLENGE_ID": fmt.Sprintf("%d", definition.AWDChallengeID),
		"ROUND":            fmt.Sprintf("%d", awdScriptRoundNumber(round)),
	}
	for key, value := range env {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		rendered[key] = renderAWDScriptCheckerValue(value, instance, definition, round, teamID, flag)
	}
	return rendered
}

func renderAWDScriptCheckerValues(values []string, instance contestports.AWDServiceInstance, definition contestports.AWDServiceDefinition, round *model.AWDRound, teamID int64, flag string) []string {
	rendered := make([]string, 0, len(values))
	for _, value := range values {
		rendered = append(rendered, renderAWDScriptCheckerValue(value, instance, definition, round, teamID, flag))
	}
	return rendered
}

func renderAWDScriptCheckerValue(value string, instance contestports.AWDServiceInstance, definition contestports.AWDServiceDefinition, round *model.AWDRound, teamID int64, flag string) string {
	host, port := splitAWDScriptCheckerTarget(instance.AccessURL)
	replacer := strings.NewReplacer(
		"{{TARGET_URL}}", strings.TrimSpace(instance.AccessURL),
		"{{TARGET_HOST}}", host,
		"{{TARGET_PORT}}", port,
		"{{FLAG}}", strings.TrimSpace(flag),
		"{{ROUND}}", fmt.Sprintf("%d", awdScriptRoundNumber(round)),
		"{{TEAM_ID}}", fmt.Sprintf("%d", teamID),
		"{{CHALLENGE_ID}}", fmt.Sprintf("%d", definition.AWDChallengeID),
		"{{AWD_CHALLENGE_ID}}", fmt.Sprintf("%d", definition.AWDChallengeID),
	)
	return replacer.Replace(value)
}

func buildAWDScriptCheckerTargetAllowlist(accessURL string) []string {
	host, port := splitAWDScriptCheckerTarget(accessURL)
	if host == "" {
		return nil
	}
	if port == "" {
		return []string{host}
	}
	return []string{net.JoinHostPort(host, port)}
}

func splitAWDScriptCheckerTarget(accessURL string) (string, string) {
	parsed, err := url.Parse(strings.TrimSpace(accessURL))
	if err != nil || parsed.Host == "" {
		return "", ""
	}
	host := parsed.Hostname()
	port := parsed.Port()
	if port == "" {
		switch parsed.Scheme {
		case "https":
			port = "443"
		case "http":
			port = "80"
		}
	}
	return host, port
}

func awdScriptRoundNumber(round *model.AWDRound) int {
	if round == nil {
		return 0
	}
	return round.RoundNumber
}

func sanitizeAWDScriptCheckerError(result contestports.CheckerRunResult) string {
	if strings.TrimSpace(result.Stderr) != "" {
		return sanitizeAWDCheckError(fmt.Errorf("%s", result.Stderr))
	}
	if strings.TrimSpace(result.Stdout) != "" {
		return sanitizeAWDCheckError(fmt.Errorf("%s", result.Stdout))
	}
	return strings.TrimSpace(result.Reason)
}
