package commands

import (
	"context"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	"errors"
	"fmt"
	"time"

	"ctf-platform/internal/apperror"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestports "ctf-platform/internal/module/contest/ports"
)

func (s *SubmissionService) validateContestSubmission(ctx context.Context, userID, contestID, challengeID int64, flag string) (*validatedContestSubmission, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, contestcontracts.ErrContestNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}

	submittedAt := time.Now().UTC()
	if contestdomain.ContestHasEndedAt(contest, submittedAt) {
		return nil, contestcontracts.ErrContestEnded
	}
	if contest.Status != contestentity.ContestStatusRunning && contest.Status != contestentity.ContestStatusFrozen {
		return nil, contestcontracts.ErrContestNotRunning
	}

	teamID, err := s.resolveTeamID(ctx, userID, contestID)
	if err != nil {
		return nil, err
	}

	contestChallenge, err := s.repo.FindContestChallenge(ctx, contestID, challengeID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestSubmissionChallengeNotFound) {
			return nil, contestcontracts.ErrChallengeNotInContest
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}

	challengeItem, err := s.repo.FindChallengeByID(ctx, challengeID)
	if err != nil {
		if errors.Is(err, contestports.ErrContestSubmissionChallengeEntityNotFound) {
			return nil, challengecontracts.ErrChallengeNotFound
		}
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if challengeItem.FlagType == contestentity.ChallengeFlagTypeManualReview {
		return nil, apperror.ErrInvalidParams.WithCause(errors.New("人工审核题暂不支持竞赛提交"))
	}

	if s.rateLimitStore == nil {
		return nil, apperror.ErrInternal.WithCause(fmt.Errorf("contest submission rate limit store is nil"))
	}
	exists, err := s.rateLimitStore.HasIncorrectSubmissionRateLimit(ctx, userID, contestID, challengeID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}
	if exists {
		return nil, challengecontracts.ErrSubmitTooFrequent
	}

	isCorrect := false
	if s.flagValidator == nil {
		return nil, apperror.ErrInternal.WithCause(fmt.Errorf("challenge flag validator is nil"))
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
