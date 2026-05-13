package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *SubmissionService) validateContestSubmission(ctx context.Context, userID, contestID, challengeID int64, flag string) (*validatedContestSubmission, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	submittedAt := time.Now().UTC()
	if !submittedAt.Before(contest.EndTime) {
		return nil, errcode.ErrContestEnded
	}
	if contest.Status != model.ContestStatusRunning && contest.Status != model.ContestStatusFrozen {
		return nil, errcode.ErrContestNotRunning
	}

	teamID, err := s.resolveTeamID(ctx, userID, contestID)
	if err != nil {
		return nil, err
	}

	contestChallenge, err := s.repo.FindContestChallenge(ctx, contestID, challengeID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestSubmissionChallengeNotFound) {
			return nil, errcode.ErrChallengeNotInContest
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	challengeItem, err := s.repo.FindChallengeByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestSubmissionChallengeEntityNotFound) {
			return nil, errcode.ErrChallengeNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if challengeItem.FlagType == model.FlagTypeManualReview {
		return nil, errcode.ErrInvalidParams.WithCause(errors.New("人工审核题暂不支持竞赛提交"))
	}

	if s.rateLimitStore == nil {
		return nil, errcode.ErrInternal.WithCause(fmt.Errorf("contest submission rate limit store is nil"))
	}
	exists, err := s.rateLimitStore.HasIncorrectSubmissionRateLimit(ctx, userID, contestID, challengeID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if exists {
		return nil, errcode.ErrSubmitTooFrequent
	}

	isCorrect := false
	if s.flagValidator == nil {
		return nil, errcode.ErrInternal.WithCause(fmt.Errorf("challenge flag validator is nil"))
	}
	isCorrect, err = s.flagValidator.ValidateFlag(ctx, userID, challengeID, flag, "")
	if err != nil {
		return nil, err
	}

	return &validatedContestSubmission{
		contestChallenge: contestChallenge,
		teamID:           teamID,
		submittedAt:      submittedAt,
		isCorrect:        isCorrect,
	}, nil
}
