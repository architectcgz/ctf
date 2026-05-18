package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	contestentity "ctf-platform/internal/module/contest/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type SolvedSubmissionRepository struct {
	source practiceports.PracticeSolvedSubmissionRepository
}

func NewSolvedSubmissionRepository(source practiceports.PracticeSolvedSubmissionRepository) *SolvedSubmissionRepository {
	if source == nil {
		return nil
	}
	return &SolvedSubmissionRepository{source: source}
}

func (r *SolvedSubmissionRepository) FindCorrectSubmission(ctx context.Context, userID, challengeID int64) (*contestentity.Submission, error) {
	submission, err := r.source.FindCorrectSubmission(ctx, userID, challengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeSolvedSubmissionNotFound
	}
	return submission, err
}

var _ practiceports.PracticeSolvedSubmissionRepository = (*SolvedSubmissionRepository)(nil)
