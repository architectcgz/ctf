package infrastructure

import (
	"context"
	"time"

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

func (r *SubmissionRepository) ConsumeSharedProof(ctx context.Context, sharedProofID int64, consumedAt time.Time) (bool, error) {
	result := r.dbWithContext(ctx).
		Model(&model.SharedProof{}).
		Where("id = ? AND status = ? AND consumed_at IS NULL AND expires_at > ?", sharedProofID, model.SharedProofStatusActive, consumedAt).
		Updates(map[string]any{
			"status":      model.SharedProofStatusConsumed,
			"consumed_at": consumedAt,
			"updated_at":  time.Now(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
