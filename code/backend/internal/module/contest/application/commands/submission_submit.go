package commands

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *SubmissionService) SubmitFlagInContest(ctx context.Context, userID, contestID, challengeID int64, flag string) (*dto.SubmissionResp, error) {
	contest, err := s.contestRepo.FindByID(ctx, contestID)
	if err != nil {
		if errors.Is(err, contestdomain.ErrContestNotFound) {
			return nil, errcode.ErrContestNotFound
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	now := time.Now()
	if !now.Before(contest.EndTime) {
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errcode.ErrChallengeNotInContest
		}
		return nil, errcode.ErrInternal.WithCause(err)
	}

	rateLimitKey := fmt.Sprintf("contest:submit:rate:%d:%d:%d", userID, contestID, challengeID)
	exists, err := s.redis.Exists(ctx, rateLimitKey).Result()
	if err == nil && exists > 0 {
		return nil, errcode.ErrSubmitTooFrequent
	}

	if s.flagValidator == nil {
		return nil, errcode.ErrInternal.WithCause(fmt.Errorf("challenge flag validator is nil"))
	}
	isCorrect, err := s.flagValidator.ValidateFlag(userID, challengeID, flag, "")
	if err != nil {
		return nil, err
	}

	submission := &model.Submission{
		UserID:      userID,
		ChallengeID: challengeID,
		ContestID:   &contestID,
		TeamID:      teamID,
		Flag:        flag,
		IsCorrect:   false,
		Score:       0,
		SubmittedAt: now,
	}

	if !isCorrect {
		_ = s.redis.Set(ctx, rateLimitKey, "1", 5*time.Second).Err()
		if err := s.repo.CreateSubmission(ctx, submission); err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
		return &dto.SubmissionResp{
			IsCorrect:   false,
			Message:     constants.MsgFlagIncorrect,
			SubmittedAt: submission.SubmittedAt,
		}, nil
	}

	finalScore, err := s.handleCorrectSubmission(ctx, submission, contestChallenge, teamID)
	if err != nil {
		return nil, err
	}

	return &dto.SubmissionResp{
		IsCorrect:   true,
		Message:     constants.MsgFlagCorrect,
		Points:      finalScore,
		SubmittedAt: submission.SubmittedAt,
	}, nil
}
