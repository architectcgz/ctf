package commands

import (
	"context"
	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	practiceports "ctf-platform/internal/module/practice/ports"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
	"errors"
	"gorm.io/gorm"
	"testing"
	"time"
)

func TestListMyChallengeSubmissionsTreatsPracticeChallengeNotFoundAsChallengeNotFound(t *testing.T) {
	t.Parallel()

	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(context.Context, int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}
	service := wirePracticeSubmissionHistoryAdapters(
		NewService(&stubPracticeRepository{}, nil, nil, nil, nil, nil, &config.Config{}, nil),
		challengeRepo,
	)

	_, err := service.ListMyChallengeSubmissions(context.Background(), 7, 11)
	if err == nil {
		t.Fatal("expected challenge not found")
	}
	var appErr *apperror.AppError
	if !errors.As(err, &appErr) || appErr.Code != challengecontracts.ErrChallengeNotFound.Code {
		t.Fatalf("expected challenge not found error, got %v", err)
	}
}

func TestListMyChallengeSubmissionsPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("list-submissions")
	expectedCtxValue := "ctx-list-submissions"
	challengeLookupCalled := false
	listCalled := false
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			challengeLookupCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected challenge lookup ctx value %v, got %v", expectedCtxValue, got)
			}
			return &challengecontracts.PracticeRuntimeChallenge{ID: id, Status: challengecontracts.ChallengeStatusPublished}, nil
		},
	}
	service := wirePracticeSubmissionHistoryAdapters(
		NewService(
			&stubPracticeRepository{
				listChallengeSubmissionsFn: func(ctx context.Context, userID, challengeID int64, limit int) ([]practiceports.SubmissionRecord, error) {
					listCalled = true
					if got := ctx.Value(ctxKey); got != expectedCtxValue {
						t.Fatalf("expected submission listing ctx value %v, got %v", expectedCtxValue, got)
					}
					return []practiceports.SubmissionRecord{{ID: 1, UserID: userID, ChallengeID: challengeID, SubmittedAt: time.Now()}}, nil
				},
			},
			nil,
			nil,
			nil,
			nil,
			nil,
			&config.Config{},
			nil,
		),
		challengeRepo,
	)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	items, err := service.ListMyChallengeSubmissions(ctx, 7, 11)
	if err != nil {
		t.Fatalf("ListMyChallengeSubmissions() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one submission item, got %+v", items)
	}
	if !challengeLookupCalled {
		t.Fatal("expected challenge lookup to be called")
	}
	if !listCalled {
		t.Fatal("expected submission listing to be called")
	}
}

func TestListMyChallengeSubmissionsMapsStoredHistory(t *testing.T) {
	t.Parallel()

	now := time.Now()
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:     id,
				Status: challengecontracts.ChallengeStatusPublished,
			}, nil
		},
	}
	service := wirePracticeSubmissionHistoryAdapters(
		NewService(
			&stubPracticeRepository{
				listChallengeSubmissionsFn: func(ctx context.Context, userID, challengeID int64, limit int) ([]practiceports.SubmissionRecord, error) {
					if userID != 7 || challengeID != 11 {
						t.Fatalf("unexpected query: user=%d challenge=%d", userID, challengeID)
					}
					if limit <= 0 {
						t.Fatalf("expected positive limit, got %d", limit)
					}
					return []practiceports.SubmissionRecord{
						{
							ID:           3,
							UserID:       7,
							ChallengeID:  11,
							IsCorrect:    true,
							ReviewStatus: contestentity.SubmissionReviewStatusNotRequired,
							SubmittedAt:  now.Add(-time.Minute),
						},
						{
							ID:           2,
							UserID:       7,
							ChallengeID:  11,
							IsCorrect:    false,
							ReviewStatus: contestentity.SubmissionReviewStatusPending,
							Flag:         "answer with reasoning",
							SubmittedAt:  now.Add(-2 * time.Minute),
						},
						{
							ID:           1,
							UserID:       7,
							ChallengeID:  11,
							IsCorrect:    false,
							ReviewStatus: contestentity.SubmissionReviewStatusNotRequired,
							SubmittedAt:  now.Add(-3 * time.Minute),
						},
					}, nil
				},
			},
			nil,
			nil,
			nil,
			nil,
			nil,
			&config.Config{},
			nil,
		),
		challengeRepo,
	)

	items, err := service.ListMyChallengeSubmissions(context.Background(), 7, 11)
	if err != nil {
		t.Fatalf("ListMyChallengeSubmissions() error = %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 records, got %d", len(items))
	}
	if items[0].Status != SubmissionStatusCorrect {
		t.Fatalf("unexpected correct record: %+v", items[0])
	}
	if items[1].Status != SubmissionStatusPendingReview {
		t.Fatalf("unexpected pending record: %+v", items[1])
	}
	if items[1].Answer != "answer with reasoning" {
		t.Fatalf("expected manual review answer to be preserved, got %+v", items[1])
	}
	if items[2].Status != SubmissionStatusIncorrect {
		t.Fatalf("unexpected incorrect record: %+v", items[2])
	}
}
