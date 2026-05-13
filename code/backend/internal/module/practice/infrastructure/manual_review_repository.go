package infrastructure

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type manualReviewSource interface {
	practiceports.PracticeSubmissionUpdateRepository
	practiceports.PracticeSolvedSubmissionRepository
	practiceports.PracticeUserLookupRepository
	practiceports.PracticeManualReviewListRepository
	practiceports.PracticeManualReviewLookupRepository
}

type ManualReviewRepository struct {
	source manualReviewSource
}

func NewManualReviewRepository(source manualReviewSource) *ManualReviewRepository {
	if source == nil {
		return nil
	}
	return &ManualReviewRepository{source: source}
}

func (r *ManualReviewRepository) GetTeacherManualReviewSubmissionByID(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
	record, err := r.source.GetTeacherManualReviewSubmissionByID(ctx, id)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeManualReviewSubmissionNotFound
	}
	return record, err
}

func (r *ManualReviewRepository) ListTeacherManualReviewSubmissions(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
	return r.source.ListTeacherManualReviewSubmissions(ctx, query)
}

func (r *ManualReviewRepository) FindCorrectSubmission(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
	submission, err := r.source.FindCorrectSubmission(ctx, userID, challengeID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeSolvedSubmissionNotFound
	}
	return submission, err
}

func (r *ManualReviewRepository) UpdateSubmission(ctx context.Context, submission *model.Submission) error {
	return r.source.UpdateSubmission(ctx, submission)
}

func (r *ManualReviewRepository) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	user, err := r.source.FindUserByID(ctx, userID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, practiceports.ErrPracticeUserNotFound
	}
	return user, err
}

var _ practiceports.PracticeManualReviewRepository = (*ManualReviewRepository)(nil)
