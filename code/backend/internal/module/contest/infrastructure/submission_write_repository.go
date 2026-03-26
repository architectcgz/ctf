package infrastructure

import (
	"context"

	"ctf-platform/internal/model"
)

func (r *SubmissionRepository) CreateSubmission(ctx context.Context, submission *model.Submission) error {
	return r.dbWithContext(ctx).Create(submission).Error
}

func (r *SubmissionRepository) UpdateSubmissionScore(ctx context.Context, submissionID int64, score int) error {
	return r.dbWithContext(ctx).
		Model(&model.Submission{}).
		Where("id = ?", submissionID).
		Update("score", score).Error
}
