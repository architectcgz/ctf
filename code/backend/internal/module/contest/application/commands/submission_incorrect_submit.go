package commands

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func (s *SubmissionService) handleIncorrectSubmission(ctx context.Context, submission *model.Submission, rateLimitKey string) (*dto.SubmissionResp, error) {
	_ = s.redis.Set(ctx, rateLimitKey, "1", s.cfg.Contest.SubmissionRateLimitTTL).Err()
	if err := s.repo.CreateSubmission(ctx, submission); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.SubmissionResp{
		IsCorrect:   false,
		SubmittedAt: submission.SubmittedAt,
	}, nil
}
