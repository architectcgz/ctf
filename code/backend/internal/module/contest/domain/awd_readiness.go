package domain

import (
	"encoding/json"
	"strings"
	"time"

	"ctf-platform/internal/model"
)

const (
	AWDReadinessBlockingReasonMissingChecker       = "missing_checker"
	AWDReadinessBlockingReasonInvalidCheckerConfig = "invalid_checker_config"
	AWDReadinessBlockingReasonPendingValidation    = "pending_validation"
	AWDReadinessBlockingReasonLastPreviewFailed    = "last_preview_failed"
	AWDReadinessBlockingReasonValidationStale      = "validation_stale"

	AWDReadinessGlobalReasonNoChallenges = "no_challenges"

	AWDReadinessActionCreateRound          = "create_round"
	AWDReadinessActionRunCurrentRoundCheck = "run_current_round_check"
	AWDReadinessActionStartContest         = "start_contest"
)

var awdReadinessBlockingActions = []string{
	AWDReadinessActionCreateRound,
	AWDReadinessActionRunCurrentRoundCheck,
	AWDReadinessActionStartContest,
}

type AWDReadinessChallenge struct {
	ServiceID         int64
	AWDChallengeID    int64
	Title             string
	CheckerType       model.AWDCheckerType
	CheckerConfig     string
	ValidationState   model.AWDCheckerValidationState
	LastPreviewAt     *time.Time
	LastPreviewResult string
}

type AWDReadinessItem struct {
	ServiceID       int64
	AWDChallengeID  int64
	Title           string
	CheckerType     model.AWDCheckerType
	ValidationState string
	LastPreviewAt   *time.Time
	LastAccessURL   *string
	BlockingReason  string
}

type AWDReadinessSummary struct {
	ContestID                int64
	Ready                    bool
	TotalChallenges          int
	PassedChallenges         int
	PendingChallenges        int
	FailedChallenges         int
	StaleChallenges          int
	MissingCheckerChallenges int
	BlockingCount            int
	BlockingActions          []string
	GlobalBlockingReasons    []string
	Items                    []AWDReadinessItem
}

func BuildAWDReadiness(contestID int64, challenges []AWDReadinessChallenge) *AWDReadinessSummary {
	summary := &AWDReadinessSummary{
		ContestID:       contestID,
		TotalChallenges: len(challenges),
		Items:           make([]AWDReadinessItem, 0, len(challenges)),
	}
	if len(challenges) == 0 {
		summary.BlockingCount = 1
		summary.GlobalBlockingReasons = []string{AWDReadinessGlobalReasonNoChallenges}
		summary.BlockingActions = append([]string(nil), awdReadinessBlockingActions...)
		return summary
	}

	for _, challenge := range challenges {
		item := AWDReadinessItem{
			ServiceID:       challenge.ServiceID,
			AWDChallengeID:  challenge.AWDChallengeID,
			Title:           challenge.Title,
			CheckerType:     NormalizeAWDCheckerType(string(challenge.CheckerType)),
			ValidationState: string(NormalizeAWDCheckerValidationState(string(challenge.ValidationState))),
			LastPreviewAt:   challenge.LastPreviewAt,
			LastAccessURL:   extractAWDReadinessAccessURL(challenge.LastPreviewResult),
		}

		switch resolveAWDReadinessBlockingReason(challenge) {
		case AWDReadinessBlockingReasonMissingChecker:
			item.BlockingReason = AWDReadinessBlockingReasonMissingChecker
			summary.MissingCheckerChallenges++
			summary.BlockingCount++
		case AWDReadinessBlockingReasonInvalidCheckerConfig:
			item.BlockingReason = AWDReadinessBlockingReasonInvalidCheckerConfig
			summary.MissingCheckerChallenges++
			summary.BlockingCount++
		case AWDReadinessBlockingReasonPendingValidation:
			item.BlockingReason = AWDReadinessBlockingReasonPendingValidation
			summary.PendingChallenges++
			summary.BlockingCount++
		case AWDReadinessBlockingReasonLastPreviewFailed:
			item.BlockingReason = AWDReadinessBlockingReasonLastPreviewFailed
			summary.FailedChallenges++
			summary.BlockingCount++
		case AWDReadinessBlockingReasonValidationStale:
			item.BlockingReason = AWDReadinessBlockingReasonValidationStale
			summary.StaleChallenges++
			summary.BlockingCount++
		default:
			summary.PassedChallenges++
		}

		summary.Items = append(summary.Items, item)
	}

	summary.Ready = summary.BlockingCount == 0
	if !summary.Ready {
		summary.BlockingActions = append([]string(nil), awdReadinessBlockingActions...)
	}
	return summary
}

func resolveAWDReadinessBlockingReason(challenge AWDReadinessChallenge) string {
	if awdReadinessCheckerMissing(challenge.CheckerType, challenge.CheckerConfig) {
		return AWDReadinessBlockingReasonMissingChecker
	}
	if awdReadinessCheckerConfigInvalid(challenge.CheckerType, challenge.CheckerConfig) {
		return AWDReadinessBlockingReasonInvalidCheckerConfig
	}

	switch NormalizeAWDCheckerValidationState(string(challenge.ValidationState)) {
	case model.AWDCheckerValidationStatePassed:
		return ""
	case model.AWDCheckerValidationStateFailed:
		return AWDReadinessBlockingReasonLastPreviewFailed
	case model.AWDCheckerValidationStateStale:
		return AWDReadinessBlockingReasonValidationStale
	default:
		return AWDReadinessBlockingReasonPendingValidation
	}
}

func awdReadinessCheckerMissing(checkerType model.AWDCheckerType, checkerConfig string) bool {
	normalizedType := NormalizeAWDCheckerType(string(checkerType))
	if normalizedType == "" {
		return true
	}

	rawConfig := strings.TrimSpace(checkerConfig)
	if rawConfig == "" {
		return true
	}
	if awdReadinessCheckerConfigInvalid(checkerType, checkerConfig) {
		return false
	}

	parsedConfig := ParseAWDCheckerConfig(rawConfig)
	if normalizedType == model.AWDCheckerTypeHTTPStandard && rawConfig == "{}" {
		return true
	}
	if rawConfig == "{}" {
		return false
	}
	if len(parsedConfig) == 0 {
		return true
	}
	return false
}

func awdReadinessCheckerConfigInvalid(checkerType model.AWDCheckerType, checkerConfig string) bool {
	if NormalizeAWDCheckerType(string(checkerType)) == "" {
		return false
	}

	rawConfig := strings.TrimSpace(checkerConfig)
	if rawConfig == "" || rawConfig == "{}" {
		return false
	}

	if !json.Valid([]byte(rawConfig)) {
		return true
	}
	return len(ParseAWDCheckerConfig(rawConfig)) == 0
}

func extractAWDReadinessAccessURL(lastPreviewResult string) *string {
	preview := ParseAWDCheckerPreviewResult(lastPreviewResult)
	if preview == nil {
		return nil
	}
	accessURL := strings.TrimSpace(preview.PreviewContext.AccessURL)
	if accessURL == "" {
		return nil
	}
	return &accessURL
}
