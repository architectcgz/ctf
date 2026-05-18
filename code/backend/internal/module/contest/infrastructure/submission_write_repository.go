package infrastructure

import (
	"context"

	contestentity "ctf-platform/internal/module/contest/entity"
)

func (r *SubmissionRepository) CreateSubmission(ctx context.Context, submission *contestentity.Submission) error {
	return r.dbWithContext(ctx).Create(submission).Error
}

func (r *SubmissionRepository) UpdateSubmissionScore(ctx context.Context, submissionID int64, score int) error {
	return r.dbWithContext(ctx).
		Model(&contestentity.Submission{}).
		Where("id = ?", submissionID).
		Update("score", score).Error
}
