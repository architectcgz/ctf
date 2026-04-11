package commands

import (
	"context"

	"ctf-platform/internal/dto"
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

	return s.repo.UpdateChallenge(ctx, contestID, challengeID, updates)
}

func zeroIfNil(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}
