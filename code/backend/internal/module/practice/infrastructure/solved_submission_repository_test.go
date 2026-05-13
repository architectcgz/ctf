package infrastructure

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

	"ctf-platform/internal/model"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type solvedSubmissionSourceStub struct {
	findCorrectSubmissionFn func(context.Context, int64, int64) (*model.Submission, error)
}

func (s solvedSubmissionSourceStub) FindCorrectSubmission(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
	return s.findCorrectSubmissionFn(ctx, userID, challengeID)
}

func TestSolvedSubmissionRepositoryMapsNotFoundErrors(t *testing.T) {
	t.Parallel()

	repo := NewSolvedSubmissionRepository(solvedSubmissionSourceStub{
		findCorrectSubmissionFn: func(context.Context, int64, int64) (*model.Submission, error) {
			return nil, gorm.ErrRecordNotFound
		},
	})

	if _, err := repo.FindCorrectSubmission(context.Background(), 1, 2); !errors.Is(err, practiceports.ErrPracticeSolvedSubmissionNotFound) {
		t.Fatalf("error = %v, want %v", err, practiceports.ErrPracticeSolvedSubmissionNotFound)
	}
}

func TestSolvedSubmissionRepositoryPassesThroughNonNotFoundErrors(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("boom")
	repo := NewSolvedSubmissionRepository(solvedSubmissionSourceStub{
		findCorrectSubmissionFn: func(context.Context, int64, int64) (*model.Submission, error) {
			return nil, expectedErr
		},
	})

	if _, err := repo.FindCorrectSubmission(context.Background(), 1, 2); !errors.Is(err, expectedErr) {
		t.Fatalf("error = %v, want %v", err, expectedErr)
	}
}
