package commands

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func buildContestAWDServiceScoreConfig(points, slaScore, defenseScore int) string {
	value := map[string]any{
		"points":            points,
		"awd_sla_score":     slaScore,
		"awd_defense_score": defenseScore,
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func buildContestAWDServiceRuntimeConfig(
	challengeID int64,
	checkerType model.AWDCheckerType,
	checkerConfig string,
	extraRuntimeConfig string,
) string {
	value := map[string]any{
		"challenge_id":   challengeID,
		"checker_type":   contestdomain.NormalizeAWDCheckerType(string(checkerType)),
		"checker_config": contestdomain.ParseAWDCheckerConfig(checkerConfig),
	}
	if extra := contestdomain.ParseAWDCheckerConfig(extraRuntimeConfig); len(extra) > 0 {
		value["template_runtime"] = extra
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func (s *ChallengeService) syncContestAWDServiceForChallenge(
	ctx context.Context,
	contest *model.Contest,
	challenge *model.Challenge,
	contestChallenge *model.ContestChallenge,
	templateID *int64,
) error {
	if s.awdRepo == nil || contest == nil || contest.Mode != model.ContestModeAWD || contestChallenge == nil {
		return nil
	}

	displayName := strings.TrimSpace(fmt.Sprintf("Challenge #%d", contestChallenge.ChallengeID))
	if challenge != nil && strings.TrimSpace(challenge.Title) != "" {
		displayName = strings.TrimSpace(challenge.Title)
	}

	record := &model.ContestAWDService{
		ContestID:         contestChallenge.ContestID,
		ChallengeID:       contestChallenge.ChallengeID,
		TemplateID:        templateID,
		DisplayName:       displayName,
		Order:             contestChallenge.Order,
		IsVisible:         contestChallenge.IsVisible,
		ScoreConfig:       buildContestAWDServiceScoreConfig(contestChallenge.Points, contestChallenge.AWDSLAScore, contestChallenge.AWDDefenseScore),
		RuntimeConfig:     buildContestAWDServiceRuntimeConfig(contestChallenge.ChallengeID, contestChallenge.AWDCheckerType, contestChallenge.AWDCheckerConfig, ""),
		ValidationState:   contestChallenge.AWDCheckerValidationState,
		LastPreviewAt:     contestChallenge.AWDCheckerLastPreviewAt,
		LastPreviewResult: contestChallenge.AWDCheckerLastPreviewResult,
	}

	stored, err := s.awdRepo.FindContestAWDServiceByContestAndChallenge(ctx, contestChallenge.ContestID, contestChallenge.ChallengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return s.awdRepo.CreateContestAWDService(ctx, record)
	}
	if err != nil {
		return err
	}

	updates := map[string]any{
		"display_name":                    record.DisplayName,
		"order":                           record.Order,
		"is_visible":                      record.IsVisible,
		"score_config":                    record.ScoreConfig,
		"runtime_config":                  record.RuntimeConfig,
		"awd_checker_validation_state":    record.ValidationState,
		"awd_checker_last_preview_at":     record.LastPreviewAt,
		"awd_checker_last_preview_result": record.LastPreviewResult,
	}
	if templateID != nil {
		updates["template_id"] = templateID
	} else if stored.TemplateID != nil {
		updates["template_id"] = stored.TemplateID
	}
	return s.awdRepo.UpdateContestAWDService(ctx, contestChallenge.ContestID, contestChallenge.ChallengeID, updates)
}
