package commands

import (
	"context"
	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/internal/module/practice/testsupport/contestentity"
	"errors"
	"testing"
	"time"
)

func TestGetTeacherManualReviewSubmissionPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("get-review")
	expectedCtxValue := "ctx-get-review"
	now := time.Now()
	getCalled := false
	findRequesterCalled := false
	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDFn: func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			getCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected get-review ctx value %v, got %v", expectedCtxValue, got)
			}
			return &practiceports.TeacherManualReviewSubmissionRecord{
				Submission:      practiceports.SubmissionRecord{ID: id, UserID: 88, ChallengeID: 11, ReviewStatus: contestentity.SubmissionReviewStatusPending, SubmittedAt: now, UpdatedAt: now},
				StudentUsername: "student88",
				StudentName:     "Student 88",
				ClassName:       "Class A",
				ChallengeTitle:  "manual challenge",
			}, nil
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			findRequesterCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &identitycontracts.User{ID: userID, Role: identitycontracts.RoleTeacher, ClassName: "Class A"}, nil
		},
	}
	service := wirePracticeManualReviewAdapters(
		newServiceCore(repo, nil, nil, nil, nil, nil, &config.Config{}, nil),
		repo,
		nil,
	)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.GetTeacherManualReviewSubmission(ctx, 91, 1001, identitycontracts.RoleTeacher); err != nil {
		t.Fatalf("GetTeacherManualReviewSubmission() error = %v", err)
	}
	if !getCalled {
		t.Fatal("expected get manual review repository to be called")
	}
	if !findRequesterCalled {
		t.Fatal("expected requester repository to be called")
	}
}

func TestGetTeacherManualReviewSubmissionRejectsStudentRole(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDFn: func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			t.Fatal("did not expect get repository call for student role")
			return nil, nil
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			t.Fatal("did not expect requester lookup for student role")
			return nil, nil
		},
	}
	service := newServiceCore(repo, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.GetTeacherManualReviewSubmission(context.Background(), 91, 1001, identitycontracts.RoleStudent)
	if err == nil {
		t.Fatal("expected student role to be rejected")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrForbidden.Code {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestGetTeacherManualReviewSubmissionTreatsPracticeManualReviewSubmissionNotFoundAsNotFound(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDFn: func(context.Context, int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			return nil, practiceports.ErrPracticeManualReviewSubmissionNotFound
		},
	}
	service := wirePracticeManualReviewAdapters(
		newServiceCore(repo, nil, nil, nil, nil, nil, &config.Config{}, nil),
		repo,
		nil,
	)

	_, err := service.GetTeacherManualReviewSubmission(context.Background(), 91, 1001, identitycontracts.RoleTeacher)
	if err == nil {
		t.Fatal("expected manual review detail lookup to fail")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrNotFound.Code {
		t.Fatalf("expected not found error, got %v", err)
	}
}
