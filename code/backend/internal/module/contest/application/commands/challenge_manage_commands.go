package commands

import (
	"context"

	"ctf-platform/pkg/errcode"
)

func (s *ChallengeService) RemoveChallengeFromContest(ctx context.Context, contestID, challengeID int64) error {
	_, err := s.ensureMutableContest(ctx, contestID)
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

	hasSubmissions, err := s.repo.HasSubmissions(ctx, contestID, challengeID)
	if err != nil {
		return errcode.ErrInternal.WithCause(err)
	}
	if hasSubmissions {
		return errcode.ErrContestChallengeHasSubs
	}

	if err := s.repo.RemoveChallenge(ctx, contestID, challengeID); err != nil {
		return err
	}
	return nil
}

func (s *ChallengeService) UpdateChallenge(ctx context.Context, contestID, challengeID int64, req UpdateContestChallengeInput) error {
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

	if err := s.repo.UpdateChallenge(ctx, contestID, challengeID, updates); err != nil {
		return err
	}
	return nil
}
