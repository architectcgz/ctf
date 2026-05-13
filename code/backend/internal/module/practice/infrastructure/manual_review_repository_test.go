package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type manualReviewSourceStub struct {
	getTeacherManualReviewSubmissionByIDFn func(context.Context, int64) (*practiceports.TeacherManualReviewSubmissionRecord, error)
	listTeacherManualReviewSubmissionsFn   func(context.Context, *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error)
	findCorrectSubmissionFn                func(context.Context, int64, int64) (*model.Submission, error)
	updateSubmissionFn                     func(context.Context, *model.Submission) error
	findUserByIDFn                         func(context.Context, int64) (*model.User, error)
}

func (s manualReviewSourceStub) GetTeacherManualReviewSubmissionByID(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
	return s.getTeacherManualReviewSubmissionByIDFn(ctx, id)
}

func (s manualReviewSourceStub) ListTeacherManualReviewSubmissions(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
	if s.listTeacherManualReviewSubmissionsFn == nil {
		return []practiceports.TeacherManualReviewSubmissionRecord{}, 0, nil
	}
	return s.listTeacherManualReviewSubmissionsFn(ctx, query)
}

func (s manualReviewSourceStub) FindCorrectSubmission(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
	return s.findCorrectSubmissionFn(ctx, userID, challengeID)
}

func (s manualReviewSourceStub) UpdateSubmission(ctx context.Context, submission *model.Submission) error {
	if s.updateSubmissionFn == nil {
		return nil
	}
	return s.updateSubmissionFn(ctx, submission)
}

func (s manualReviewSourceStub) FindUserByID(ctx context.Context, userID int64) (*model.User, error) {
	return s.findUserByIDFn(ctx, userID)
}

func TestManualReviewRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewManualReviewRepository(manualReviewSourceStub{
		getTeacherManualReviewSubmissionByIDFn: func(context.Context, int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findCorrectSubmissionFn: func(context.Context, int64, int64) (*model.Submission, error) {
			return nil, gorm.ErrRecordNotFound
		},
		findUserByIDFn: func(context.Context, int64) (*model.User, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.GetTeacherManualReviewSubmissionByID(context.Background(), 1); !errors.Is(err, practiceports.ErrPracticeManualReviewSubmissionNotFound) {
		t.Fatalf("manual review error = %v, want %v", err, practiceports.ErrPracticeManualReviewSubmissionNotFound)
	}
	if _, err := repo.FindCorrectSubmission(context.Background(), 1, 2); !errors.Is(err, practiceports.ErrPracticeSolvedSubmissionNotFound) {
		t.Fatalf("solved submission error = %v, want %v", err, practiceports.ErrPracticeSolvedSubmissionNotFound)
	}
	if _, err := repo.FindUserByID(context.Background(), 3); !errors.Is(err, practiceports.ErrPracticeUserNotFound) {
		t.Fatalf("user error = %v, want %v", err, practiceports.ErrPracticeUserNotFound)
	}
}

func TestManualReviewRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewManualReviewRepository(manualReviewSourceStub{
		getTeacherManualReviewSubmissionByIDFn: func(context.Context, int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			return nil, expectedErr
		},
		findCorrectSubmissionFn: func(context.Context, int64, int64) (*model.Submission, error) {
			return &model.Submission{ID: 1}, nil
		},
		findUserByIDFn: func(context.Context, int64) (*model.User, error) {
			return &model.User{ID: 2}, nil
		},
	})

	if _, err := repo.GetTeacherManualReviewSubmissionByID(context.Background(), 1); !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}
