package commands

import (
	"context"
	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	"errors"
	"strings"
	"testing"
)

func TestListTeacherManualReviewSubmissionsPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("list-review")
	expectedCtxValue := "ctx-list-review"
	listCalled := false
	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &identitycontracts.User{ID: userID, Role: identitycontracts.RoleTeacher, ClassName: "Class A"}, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *practicecontracts.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			listCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected list-review ctx value %v, got %v", expectedCtxValue, got)
			}
			if query.ClassName != "Class A" {
				t.Fatalf("expected normalized class name, got %+v", query)
			}
			return []practiceports.TeacherManualReviewSubmissionRecord{}, 0, nil
		},
	}
	service := wirePracticeManualReviewAdapters(
		NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil),
		repo,
		nil,
	)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.ListTeacherManualReviewSubmissions(ctx, 1001, identitycontracts.RoleTeacher, &practicecontracts.TeacherManualReviewSubmissionQuery{}); err != nil {
		t.Fatalf("ListTeacherManualReviewSubmissions() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected list manual review repository to be called")
	}
}

func TestListTeacherManualReviewSubmissionsRejectsStudentRole(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			t.Fatal("did not expect requester lookup for student role")
			return nil, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *practicecontracts.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			t.Fatal("did not expect list repository call for student role")
			return nil, 0, nil
		},
	}
	service := wirePracticeManualReviewAdapters(
		NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil),
		repo,
		nil,
	)

	_, err := service.ListTeacherManualReviewSubmissions(context.Background(), 1001, identitycontracts.RoleStudent, &practicecontracts.TeacherManualReviewSubmissionQuery{})
	if err == nil {
		t.Fatal("expected student role to be rejected")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrForbidden.Code {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestListTeacherManualReviewSubmissionsRejectsInvalidReviewStatus(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			t.Fatal("did not expect requester lookup for invalid review status")
			return nil, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *practicecontracts.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			t.Fatal("did not expect list repository call for invalid review status")
			return nil, 0, nil
		},
	}
	service := wirePracticeManualReviewAdapters(
		NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil),
		repo,
		nil,
	)

	_, err := service.ListTeacherManualReviewSubmissions(
		context.Background(),
		1001,
		identitycontracts.RoleTeacher,
		&practicecontracts.TeacherManualReviewSubmissionQuery{ReviewStatus: "archived"},
	)
	if err == nil {
		t.Fatal("expected invalid review status to be rejected")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestListTeacherManualReviewSubmissionsRejectsOversizedPageSize(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			t.Fatal("did not expect requester lookup for oversized page size")
			return nil, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *practicecontracts.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			t.Fatal("did not expect list repository call for oversized page size")
			return nil, 0, nil
		},
	}
	service := wirePracticeManualReviewAdapters(
		NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil),
		repo,
		nil,
	)

	_, err := service.ListTeacherManualReviewSubmissions(
		context.Background(),
		1001,
		identitycontracts.RoleTeacher,
		&practicecontracts.TeacherManualReviewSubmissionQuery{Size: 101},
	)
	if err == nil {
		t.Fatal("expected oversized page size to be rejected")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestListTeacherManualReviewSubmissionsRejectsNonPositiveStudentID(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			t.Fatal("did not expect requester lookup for non-positive student id")
			return nil, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *practicecontracts.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			t.Fatal("did not expect list repository call for non-positive student id")
			return nil, 0, nil
		},
	}
	service := wirePracticeManualReviewAdapters(
		NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil),
		repo,
		nil,
	)
	studentID := int64(0)

	_, err := service.ListTeacherManualReviewSubmissions(
		context.Background(),
		1001,
		identitycontracts.RoleTeacher,
		&practicecontracts.TeacherManualReviewSubmissionQuery{StudentID: &studentID},
	)
	if err == nil {
		t.Fatal("expected non-positive student id to be rejected")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestListTeacherManualReviewSubmissionsRejectsNonPositiveChallengeID(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			t.Fatal("did not expect requester lookup for non-positive challenge id")
			return nil, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *practicecontracts.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			t.Fatal("did not expect list repository call for non-positive challenge id")
			return nil, 0, nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil)
	challengeID := int64(0)

	_, err := service.ListTeacherManualReviewSubmissions(
		context.Background(),
		1001,
		identitycontracts.RoleTeacher,
		&practicecontracts.TeacherManualReviewSubmissionQuery{ChallengeID: &challengeID},
	)
	if err == nil {
		t.Fatal("expected non-positive challenge id to be rejected")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestListTeacherManualReviewSubmissionsRejectsOversizedClassName(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			t.Fatal("did not expect requester lookup for oversized class name")
			return nil, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *practicecontracts.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			t.Fatal("did not expect list repository call for oversized class name")
			return nil, 0, nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.ListTeacherManualReviewSubmissions(
		context.Background(),
		1001,
		identitycontracts.RoleAdmin,
		&practicecontracts.TeacherManualReviewSubmissionQuery{ClassName: strings.Repeat("A", 129)},
	)
	if err == nil {
		t.Fatal("expected oversized class name to be rejected")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}
