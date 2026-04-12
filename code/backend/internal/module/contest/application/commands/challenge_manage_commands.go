package commands

import (
	"context"
	"strings"

	redis "github.com/redis/go-redis/v9"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func (s *ChallengeService) RemoveChallengeFromContest(ctx context.Context, contestID, challengeID int64) error {
	if _, err := s.ensureMutableContest(ctx, contestID); err != nil {
		return err
	}

	exists, err := s.repo.Exists(ctx, contestID, challengeID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if !exists {
		return errcode.ErrChallengeNotInContest
	}

	hasSubmissions, err := s.repo.HasSubmissions(ctx, contestID, challengeID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if hasSubmissions {
		return errcode.ErrContestChallengeHasSubs
	}

	return s.repo.RemoveChallenge(ctx, contestID, challengeID)
}

func (s *ChallengeService) UpdateChallenge(ctx context.Context, contestID, challengeID int64, req *dto.UpdateContestChallengeReq) error {
	contest, err := s.ensureMutableContest(ctx, contestID)
	if err != nil {
		return err
	}

	exists, err := s.repo.Exists(ctx, contestID, challengeID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if !exists {
		return errcode.ErrChallengeNotInContest
	}
	current, err := s.repo.FindChallenge(ctx, contestID, challengeID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}

	updates := make(map[string]any)
	if req.Points != nil {
		updates["points"] = *req.Points
	}
	if req.Order != nil {
		updates["order"] = *req.Order
	}
	if req.IsVisible != nil {
		updates["is_visible"] = *req.IsVisible
	}
	checkerTypeValue := ""
	if req.AWDCheckerType != nil {
		checkerTypeValue = *req.AWDCheckerType
	}
	checkerType, checkerConfig, err := validateAndNormalizeContestAWDFields(
		contest,
		checkerTypeValue,
		req.AWDCheckerConfig,
		zeroIfNil(req.AWDSLAScore),
		zeroIfNil(req.AWDDefenseScore),
	)
	if err != nil {
		return err
	}
	if req.AWDCheckerType != nil {
		updates["awd_checker_type"] = checkerType
	}
	if req.AWDCheckerConfig != nil {
		updates["awd_checker_config"] = checkerConfig
	}
	if req.AWDSLAScore != nil {
		updates["awd_sla_score"] = *req.AWDSLAScore
	}
	if req.AWDDefenseScore != nil {
		updates["awd_defense_score"] = *req.AWDDefenseScore
	}
	nextCheckerType := current.AWDCheckerType
	if req.AWDCheckerType != nil {
		nextCheckerType = checkerType
	}
	nextCheckerConfig := current.AWDCheckerConfig
	if req.AWDCheckerConfig != nil {
		nextCheckerConfig = checkerConfig
	}
	previewToken := ""
	if req.AWDCheckerPreviewToken != nil {
		previewToken = *req.AWDCheckerPreviewToken
	}
	if validationUpdates, ok, validationErr := buildCheckerValidationUpdate(
		ctx,
		s.redis,
		current,
		contestID,
		challengeID,
		nextCheckerType,
		nextCheckerConfig,
		previewToken,
	); validationErr != nil {
		return errcode.ErrInternal.WithCause(validationErr)
	} else if ok {
		for key, value := range validationUpdates {
			updates[key] = value
		}
	}

	return s.repo.UpdateChallenge(ctx, contestID, challengeID, updates)
}

func zeroIfNil(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func buildCheckerValidationUpdate(
	ctx context.Context,
	redisClient *redis.Client,
	current *model.ContestChallenge,
	contestID, challengeID int64,
	nextCheckerType model.AWDCheckerType,
	nextCheckerConfig string,
	previewToken string,
) (map[string]any, bool, error) {
	state, previewAt, previewResult, err := consumeCheckerPreviewValidationState(
		ctx,
		redisClient,
		contestID,
		challengeID,
		nextCheckerType,
		nextCheckerConfig,
		previewToken,
	)
	if err != nil {
		return nil, false, err
	}
	if strings.TrimSpace(previewToken) != "" && previewResult != "" {
		return map[string]any{
			"awd_checker_validation_state":    state,
			"awd_checker_last_preview_at":     previewAt,
			"awd_checker_last_preview_result": previewResult,
		}, true, nil
	}

	configChanged := current.AWDCheckerType != nextCheckerType || current.AWDCheckerConfig != nextCheckerConfig
	if !configChanged {
		return nil, false, nil
	}

	nextState := model.AWDCheckerValidationStatePending
	if hasPersistedCheckerValidation(current) {
		nextState = model.AWDCheckerValidationStateStale
	}
	return map[string]any{
		"awd_checker_validation_state": nextState,
	}, true, nil
}

func hasPersistedCheckerValidation(value *model.ContestChallenge) bool {
	if value == nil {
		return false
	}
	if value.AWDCheckerLastPreviewAt != nil {
		return true
	}
	if strings.TrimSpace(value.AWDCheckerLastPreviewResult) != "" {
		return true
	}
	return value.AWDCheckerValidationState != "" && value.AWDCheckerValidationState != model.AWDCheckerValidationStatePending
}
