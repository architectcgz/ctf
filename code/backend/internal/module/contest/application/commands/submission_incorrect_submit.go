package commands

import (
	"context"
	"fmt"

	contestentity "ctf-platform/internal/module/contest/entity"
	"ctf-platform/pkg/errcode"
)

func (s *SubmissionService) handleIncorrectSubmission(ctx context.Context, submission *contestentity.Submission) (*SubmissionResp, error) {
	if submission == nil || submission.ContestID == nil {
		return nil, errcode.ErrInternal.WithCause(fmt.Errorf("contest submission is incomplete"))
	}
	if s.rateLimitStore == nil {
		return nil, errcode.ErrInternal.WithCause(fmt.Errorf("contest submission rate limit store is nil"))
	}
	if err := s.rateLimitStore.SetIncorrectSubmissionRateLimit(ctx, submission.UserID, *submission.ContestID, submission.ChallengeID, s.cfg.Contest.SubmissionRateLimitTTL); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if err := s.repo.CreateSubmission(ctx, submission); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return contestResponseMapperInst.ToSubmissionRespPtr(submissionRespSource{
		IsCorrect:   false,
		Status:      SubmissionStatusIncorrect,
		SubmittedAt: submission.SubmittedAt,
	}), nil
}
