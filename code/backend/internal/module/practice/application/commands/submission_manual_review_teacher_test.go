package commands

import (
	"context"
	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
	"ctf-platform/internal/shared/taxonomy"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestReviewManualReviewSubmissionPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("review")
	expectedCtxValue := "ctx-review-manual"
	now := time.Now()
	updatedCalled := false
	findRequesterCalled := false
	findRecordCalled := false
	challengeLookupCalled := false
	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDFn: func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			findRecordCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected get-review-record ctx value %v, got %v", expectedCtxValue, got)
			}
			return &practiceports.TeacherManualReviewSubmissionRecord{
				Submission: practiceports.SubmissionRecord{
					ID:           id,
					UserID:       88,
					ChallengeID:  11,
					Flag:         "answer",
					ReviewStatus: contestentity.SubmissionReviewStatusPending,
					SubmittedAt:  now,
					UpdatedAt:    now,
				},
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
		updateSubmissionFn: func(ctx context.Context, submission *practiceports.SubmissionRecord) error {
			updatedCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			challengeLookupCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected challenge lookup ctx value %v, got %v", expectedCtxValue, got)
			}
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:       id,
				Category: taxonomy.DimensionWeb,
				Points:   120,
				Status:   challengecontracts.ChallengeStatusPublished,
				FlagType: challengecontracts.FlagTypeManualReview,
			}, nil
		},
	}
	service := wirePracticeManualReviewAdapters(
		NewService(
			repo,

			nil,
			nil,
			nil,
			nil,
			nil,
			&config.Config{},
			nil),

		repo,
		challengeRepo,
	)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.ReviewManualReviewSubmission(
		ctx,
		91,
		1001,
		identitycontracts.RoleTeacher,
		&practicecontracts.ReviewManualReviewSubmissionReq{ReviewStatus: contestentity.SubmissionReviewStatusApproved},
	); err != nil {
		t.Fatalf("ReviewManualReviewSubmission() error = %v", err)
	}
	if !findRecordCalled {
		t.Fatal("expected review record repository to be called")
	}
	if !findRequesterCalled {
		t.Fatal("expected requester repository to be called")
	}
	if !challengeLookupCalled {
		t.Fatal("expected challenge lookup to be called")
	}
	if !updatedCalled {
		t.Fatal("expected update submission repository to be called")
	}
}

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
		NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil),
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
	service := NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.GetTeacherManualReviewSubmission(context.Background(), 91, 1001, identitycontracts.RoleStudent)
	if err == nil {
		t.Fatal("expected student role to be rejected")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrForbidden.Code {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestReviewManualReviewSubmissionRejectsStudentRole(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDFn: func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			t.Fatal("did not expect review record lookup for student role")
			return nil, nil
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			t.Fatal("did not expect requester lookup for student role")
			return nil, nil
		},
		updateSubmissionFn: func(ctx context.Context, submission *practiceports.SubmissionRecord) error {
			t.Fatal("did not expect submission update for student role")
			return nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.ReviewManualReviewSubmission(
		context.Background(),
		91,
		1001,
		identitycontracts.RoleStudent,
		&practicecontracts.ReviewManualReviewSubmissionReq{ReviewStatus: contestentity.SubmissionReviewStatusApproved},
	)
	if err == nil {
		t.Fatal("expected student role to be rejected")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrForbidden.Code {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestReviewManualReviewSubmissionRejectsInvalidReviewStatus(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDFn: func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			t.Fatal("did not expect review record lookup for invalid review status")
			return nil, nil
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			t.Fatal("did not expect requester lookup for invalid review status")
			return nil, nil
		},
		updateSubmissionFn: func(ctx context.Context, submission *practiceports.SubmissionRecord) error {
			t.Fatal("did not expect submission update for invalid review status")
			return nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.ReviewManualReviewSubmission(
		context.Background(),
		91,
		1001,
		identitycontracts.RoleTeacher,
		&practicecontracts.ReviewManualReviewSubmissionReq{ReviewStatus: contestentity.SubmissionReviewStatusPending},
	)
	if err == nil {
		t.Fatal("expected invalid review status to be rejected")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestReviewManualReviewSubmissionRejectsOversizedReviewComment(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDFn: func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			t.Fatal("did not expect review record lookup for oversized review comment")
			return nil, nil
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			t.Fatal("did not expect requester lookup for oversized review comment")
			return nil, nil
		},
		updateSubmissionFn: func(ctx context.Context, submission *practiceports.SubmissionRecord) error {
			t.Fatal("did not expect submission update for oversized review comment")
			return nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.ReviewManualReviewSubmission(
		context.Background(),
		91,
		1001,
		identitycontracts.RoleTeacher,
		&practicecontracts.ReviewManualReviewSubmissionReq{
			ReviewStatus:  contestentity.SubmissionReviewStatusApproved,
			ReviewComment: strings.Repeat("a", 4001),
		},
	)
	if err == nil {
		t.Fatal("expected oversized review comment to be rejected")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != apperror.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestReviewManualReviewSubmissionRejectsApprovalAfterChallengeAlreadySolved(t *testing.T) {
	t.Parallel()

	now := time.Now()
	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDFn: func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			return &practiceports.TeacherManualReviewSubmissionRecord{
				Submission: practiceports.SubmissionRecord{
					ID:           id,
					UserID:       88,
					ChallengeID:  11,
					Flag:         "answer",
					ReviewStatus: contestentity.SubmissionReviewStatusPending,
					SubmittedAt:  now,
					UpdatedAt:    now,
				},
				StudentUsername: "student88",
				StudentName:     "Student 88",
				ClassName:       "Class A",
				ChallengeTitle:  "manual challenge",
			}, nil
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			return &identitycontracts.User{ID: userID, Role: identitycontracts.RoleTeacher, ClassName: "Class A"}, nil
		},
		findCorrectSubmissionFn: func(ctx context.Context, userID, challengeID int64) (*practiceports.SubmissionRecord, error) {
			return &practiceports.SubmissionRecord{
				ID:           99,
				UserID:       userID,
				ChallengeID:  challengeID,
				IsCorrect:    true,
				ReviewStatus: contestentity.SubmissionReviewStatusApproved,
				SubmittedAt:  now.Add(-time.Minute),
				UpdatedAt:    now.Add(-time.Minute),
			}, nil
		},
		updateSubmissionFn: func(ctx context.Context, submission *practiceports.SubmissionRecord) error {
			t.Fatal("did not expect submission update when challenge already solved")
			return nil
		},
	}
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:       id,
				Category: taxonomy.DimensionWeb,
				Points:   120,
				Status:   challengecontracts.ChallengeStatusPublished,
				FlagType: challengecontracts.FlagTypeManualReview,
			}, nil
		},
	}
	service := wirePracticeManualReviewAdapters(
		NewService(
			repo,

			nil,
			nil,
			nil,
			nil,
			nil,
			&config.Config{},
			nil),

		repo,
		challengeRepo,
	)

	_, err := service.ReviewManualReviewSubmission(
		context.Background(),
		91,
		1001,
		identitycontracts.RoleTeacher,
		&practicecontracts.ReviewManualReviewSubmissionReq{ReviewStatus: contestentity.SubmissionReviewStatusApproved},
	)
	if err == nil {
		t.Fatal("expected already solved approval to be rejected")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != challengecontracts.ErrAlreadySolved.Code {
		t.Fatalf("expected already solved error, got %v", err)
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
		NewService(repo, nil, nil, nil, nil, nil, &config.Config{}, nil),
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
