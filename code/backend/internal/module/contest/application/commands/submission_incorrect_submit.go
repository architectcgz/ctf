package commands

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func (s *SubmissionService) handleIncorrectSubmission(ctx context.Context, submission *model.Submission, rateLimitKey string) (*dto.SubmissionResp, error) {
	if err := s.redis.Set(ctx, rateLimitKey, "1", s.cfg.Contest.SubmissionRateLimitTTL).Err(); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if err := s.repo.CreateSubmission(ctx, submission); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return contestResponseMapperInst.ToSubmissionRespPtr(submissionRespSource{
		IsCorrect:   false,
		SubmittedAt: submission.SubmittedAt,
	}), nil
}
