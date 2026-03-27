package commands

import (
	"context"
	"time"

	"ctf-platform/internal/constants"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func (s *SubmissionService) handleIncorrectSubmission(ctx context.Context, submission *model.Submission, rateLimitKey string) (*dto.SubmissionResp, error) {
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
