package commands

import (
	"context"
	"ctf-platform/internal/apperror"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
	"ctf-platform/internal/shared/taxonomy"
	"errors"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"sync/atomic"
	"testing"
	"time"
)

func newPracticeFlagSubmitRateLimitStoreForTest(redisClient *redis.Client) practiceports.PracticeFlagSubmitRateLimitStore {
	return practiceinfra.NewFlagSubmitRateLimitStore(redisClient, "practice:test")
}

func TestSubmitFlagWithManualReviewChallengeCreatesPendingSubmission(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	var createdSubmission *practiceports.SubmissionRecord
	repo := &stubPracticeRepository{
		createSubmissionFn: func(ctx context.Context, submission *practiceports.SubmissionRecord) error {
			createdSubmission = submission
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
			newPracticeFlagSubmitRateLimitStoreForTest(redisClient),
			&config.Config{
				RateLimit: config.RateLimitConfig{
					RedisKeyPrefix: "practice:test",
					FlagSubmit: config.RateLimitPolicyConfig{
						Limit:  5,
						Window: time.Minute,
					},
				},
			},
			nil),

		repo,
		challengeRepo,
	)

	resp, err := service.SubmitFlag(context.Background(), 8, 18, "answer with reasoning")
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if resp.IsCorrect || resp.Status != SubmissionStatusPendingReview {
		t.Fatalf("expected pending-review response, got %+v", resp)
	}
	if createdSubmission == nil {
		t.Fatal("expected submission to be created")
	}
	if createdSubmission.Flag != "answer with reasoning" {
		t.Fatalf("expected raw answer stored for manual review, got %+v", createdSubmission)
	}
	if createdSubmission.ReviewStatus != contestentity.SubmissionReviewStatusPending {
		t.Fatalf("expected pending review status, got %+v", createdSubmission)
	}
}

func TestReviewManualReviewSubmissionApprovesAndTriggersScoreUpdate(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	now := time.Now()
	submissionID := int64(71)
	reviewerID := int64(301)
	studentID := int64(201)
	var updatedSubmission *practiceports.SubmissionRecord
	var scoreUpdateCalls atomic.Int32
	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDFn: func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			if id != submissionID {
				t.Fatalf("unexpected submission id: %d", id)
			}
			return &practiceports.TeacherManualReviewSubmissionRecord{
				Submission: practiceports.SubmissionRecord{
					ID:           submissionID,
					UserID:       studentID,
					ChallengeID:  19,
					Flag:         "answer text",
					ReviewStatus: contestentity.SubmissionReviewStatusPending,
					SubmittedAt:  now,
				},
				StudentUsername: "student",
				StudentName:     "Student",
				ClassName:       "Class 1",
				ChallengeTitle:  "manual challenge",
			}, nil
		},
		updateSubmissionFn: func(ctx context.Context, submission *practiceports.SubmissionRecord) error {
			updatedSubmission = submission
			return nil
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*identitycontracts.User, error) {
			return &identitycontracts.User{ID: userID, Username: "teacher", Role: identitycontracts.RoleTeacher, ClassName: "Class 1"}, nil
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
			&stubScoreUpdater{
				updateFn: func(ctx context.Context, userID int64) error {
					if userID != studentID {
						t.Fatalf("unexpected score update user: %d", userID)
					}
					scoreUpdateCalls.Add(1)
					return nil
				},
			},
			newPracticeFlagSubmitRateLimitStoreForTest(redisClient),
			&config.Config{
				RateLimit: config.RateLimitConfig{
					RedisKeyPrefix: "practice:test",
					FlagSubmit: config.RateLimitPolicyConfig{
						Limit:  5,
						Window: time.Minute,
					},
				},
				Cache: config.CacheConfig{
					ProgressTTL: time.Minute,
				},
			},
			nil),

		repo,
		challengeRepo,
	)

	service.StartBackgroundTasks(context.Background())

	resp, err := service.ReviewManualReviewSubmission(
		context.Background(),
		submissionID,
		reviewerID,
		identitycontracts.RoleTeacher,
		&practicecontracts.ReviewManualReviewSubmissionReq{
			ReviewStatus:  contestentity.SubmissionReviewStatusApproved,
			ReviewComment: "答案链路完整",
		},
	)
	if err != nil {
		t.Fatalf("ReviewManualReviewSubmission() error = %v", err)
	}
	if resp.ReviewStatus != contestentity.SubmissionReviewStatusApproved || !resp.IsCorrect || resp.Score != 120 {
		t.Fatalf("unexpected review response: %+v", resp)
	}
	if updatedSubmission == nil {
		t.Fatal("expected submission to be updated")
	}
	if updatedSubmission.ReviewStatus != contestentity.SubmissionReviewStatusApproved || !updatedSubmission.IsCorrect || updatedSubmission.Score != 120 {
		t.Fatalf("unexpected updated submission: %+v", updatedSubmission)
	}
	requireEventually(t, time.Second, func() bool {
		return scoreUpdateCalls.Load() == 1
	})
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
