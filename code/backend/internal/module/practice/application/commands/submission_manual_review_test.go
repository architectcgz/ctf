package commands

import (
	"context"
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	"ctf-platform/internal/platform/events"
	flagcrypto "ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
	"errors"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestSubmitFlagWithRegexChallengeMatchesPattern(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	repo := &stubPracticeRepository{}
	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:        id,
					Category:  model.DimensionWeb,
					Points:    80,
					Status:    model.ChallengeStatusPublished,
					FlagType:  model.FlagTypeRegex,
					FlagRegex: `^flag\{regex-[0-9]{2}\}$`,
				}, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		redisClient,
		&config.Config{
			RateLimit: config.RateLimitConfig{
				RedisKeyPrefix: "practice:test",
				FlagSubmit: config.RateLimitPolicyConfig{
					Limit:  5,
					Window: time.Minute,
				},
			},
		},
		nil)

	resp, err := service.SubmitFlag(context.Background(), 9, 19, "flag{regex-42}")
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !resp.IsCorrect || resp.Status != dto.SubmissionStatusCorrect {
		t.Fatalf("expected regex submission success, got %+v", resp)
	}
}

func TestSubmitFlagWithManualReviewChallengeCreatesPendingSubmission(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	var createdSubmission *model.Submission
	repo := &stubPracticeRepository{
		createSubmissionFn: func(ctx context.Context, submission *model.Submission) error {
			createdSubmission = submission
			return nil
		},
	}
	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:       id,
					Category: model.DimensionWeb,
					Points:   120,
					Status:   model.ChallengeStatusPublished,
					FlagType: model.FlagTypeManualReview,
				}, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		redisClient,
		&config.Config{
			RateLimit: config.RateLimitConfig{
				RedisKeyPrefix: "practice:test",
				FlagSubmit: config.RateLimitPolicyConfig{
					Limit:  5,
					Window: time.Minute,
				},
			},
		},
		nil)

	resp, err := service.SubmitFlag(context.Background(), 8, 18, "answer with reasoning")
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if resp.IsCorrect || resp.Status != dto.SubmissionStatusPendingReview {
		t.Fatalf("expected pending-review response, got %+v", resp)
	}
	if createdSubmission == nil {
		t.Fatal("expected submission to be created")
	}
	if createdSubmission.Flag != "answer with reasoning" {
		t.Fatalf("expected raw answer stored for manual review, got %+v", createdSubmission)
	}
	if createdSubmission.ReviewStatus != model.SubmissionReviewStatusPending {
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
	var updatedSubmission *model.Submission
	var scoreUpdateCalls atomic.Int32
	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDFn: func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			if id != submissionID {
				t.Fatalf("unexpected submission id: %d", id)
			}
			return &practiceports.TeacherManualReviewSubmissionRecord{
				Submission: model.Submission{
					ID:           submissionID,
					UserID:       studentID,
					ChallengeID:  19,
					Flag:         "answer text",
					ReviewStatus: model.SubmissionReviewStatusPending,
					SubmittedAt:  now,
				},
				StudentUsername: "student",
				StudentName:     "Student",
				ClassName:       "Class 1",
				ChallengeTitle:  "manual challenge",
			}, nil
		},
		updateSubmissionFn: func(ctx context.Context, submission *model.Submission) error {
			updatedSubmission = submission
			return nil
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			return &model.User{ID: userID, Username: "teacher", Role: model.RoleTeacher, ClassName: "Class 1"}, nil
		},
	}
	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:       id,
					Category: model.DimensionWeb,
					Points:   120,
					Status:   model.ChallengeStatusPublished,
					FlagType: model.FlagTypeManualReview,
				}, nil
			},
		},
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
		redisClient,
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
		nil)

	service.StartBackgroundTasks(context.Background())

	resp, err := service.ReviewManualReviewSubmission(
		context.Background(),
		submissionID,
		reviewerID,
		model.RoleTeacher,
		&dto.ReviewManualReviewSubmissionReq{
			ReviewStatus:  model.SubmissionReviewStatusApproved,
			ReviewComment: "答案链路完整",
		},
	)
	if err != nil {
		t.Fatalf("ReviewManualReviewSubmission() error = %v", err)
	}
	if resp.ReviewStatus != model.SubmissionReviewStatusApproved || !resp.IsCorrect || resp.Score != 120 {
		t.Fatalf("unexpected review response: %+v", resp)
	}
	if updatedSubmission == nil {
		t.Fatal("expected submission to be updated")
	}
	if updatedSubmission.ReviewStatus != model.SubmissionReviewStatusApproved || !updatedSubmission.IsCorrect || updatedSubmission.Score != 120 {
		t.Fatalf("unexpected updated submission: %+v", updatedSubmission)
	}
	requireEventually(t, time.Second, func() bool {
		return scoreUpdateCalls.Load() == 1
	})
}

func TestPracticePublishesFlagAcceptedEvent(t *testing.T) {
	t.Parallel()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Submission{}); err != nil {
		t.Fatalf("migrate submissions: %v", err)
	}

	mr := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = redisClient.Close() })
	flagSalt := "static-salt"

	bus := events.NewBus()
	repo := &stubPracticeRepository{
		findCorrectSubmissionFn: func(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
			return nil, gorm.ErrRecordNotFound
		},
		createSubmissionFn: func(ctx context.Context, submission *model.Submission) error {
			return db.Create(submission).Error
		},
	}
	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:       id,
					Category: model.DimensionWeb,
					Points:   100,
					Status:   model.ChallengeStatusPublished,
					FlagType: model.FlagTypeStatic,
					FlagSalt: flagSalt,
					FlagHash: flagcrypto.HashStaticFlag("flag{correct}", flagSalt),
				}, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		redisClient,
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
		nil)

	service.SetEventBus(bus)

	received := make(chan practicecontracts.FlagAcceptedEvent, 1)
	bus.Subscribe(practicecontracts.EventFlagAccepted, func(_ context.Context, evt events.Event) error {
		payload, ok := evt.Payload.(practicecontracts.FlagAcceptedEvent)
		if !ok {
			t.Fatalf("unexpected payload type: %T", evt.Payload)
		}
		received <- payload
		return nil
	})

	resp, err := service.SubmitFlag(context.Background(), 7, 11, "flag{correct}")
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !resp.IsCorrect {
		t.Fatalf("expected correct submission response, got %+v", resp)
	}

	select {
	case evt := <-received:
		if evt.UserID != 7 || evt.ChallengeID != 11 || evt.Dimension != model.DimensionWeb {
			t.Fatalf("unexpected event payload: %+v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("expected practice.flag_accepted event to be published")
	}
}

func TestSubmitFlagWithSharedStaticChallengeUsesRegularFlagValidation(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	flagSalt := "shared-static-salt"

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	service := NewService(
		practiceinfra.NewRepository(db),
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:              id,
					Category:        model.DimensionWeb,
					Points:          100,
					Status:          model.ChallengeStatusPublished,
					FlagType:        model.FlagTypeStatic,
					FlagSalt:        flagSalt,
					FlagHash:        flagcrypto.HashStaticFlag("flag{shared-static}", flagSalt),
					InstanceSharing: model.InstanceSharingShared,
				}, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		redisClient,
		&config.Config{
			RateLimit: config.RateLimitConfig{
				RedisKeyPrefix: "practice:test",
				FlagSubmit: config.RateLimitPolicyConfig{
					Limit:  5,
					Window: time.Minute,
				},
			},
		},
		nil)

	resp, err := service.SubmitFlag(context.Background(), 7, 11, "flag{shared-static}")
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !resp.IsCorrect || resp.Status != dto.SubmissionStatusCorrect {
		t.Fatalf("expected shared static submission success, got %+v", resp)
	}
}

func TestSubmitFlagAllowsRepeatCorrectSubmissionWithoutExtraPoints(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	flagSalt := "repeat-submit-salt"

	if err := db.Create(&model.User{
		ID:        71,
		Username:  "student-repeat",
		Role:      model.RoleStudent,
		Status:    model.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	service := NewService(
		practiceinfra.NewRepository(db),
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:       id,
					Category: model.DimensionWeb,
					Points:   100,
					Status:   model.ChallengeStatusPublished,
					FlagType: model.FlagTypeStatic,
					FlagSalt: flagSalt,
					FlagHash: flagcrypto.HashStaticFlag("flag{repeatable}", flagSalt),
				}, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		redisClient,
		&config.Config{
			RateLimit: config.RateLimitConfig{
				RedisKeyPrefix: "practice:test",
				FlagSubmit: config.RateLimitPolicyConfig{
					Limit:  5,
					Window: time.Minute,
				},
			},
		},
		nil)

	first, err := service.SubmitFlag(context.Background(), 71, 11, "flag{repeatable}")
	if err != nil {
		t.Fatalf("SubmitFlag() first error = %v", err)
	}
	if !first.IsCorrect || first.Points != 100 {
		t.Fatalf("expected first correct submission to score once, got %+v", first)
	}

	repeat, err := service.SubmitFlag(context.Background(), 71, 11, "flag{repeatable}")
	if err != nil {
		t.Fatalf("SubmitFlag() repeat error = %v", err)
	}
	if !repeat.IsCorrect || repeat.Status != dto.SubmissionStatusCorrect {
		t.Fatalf("expected repeated correct submission to stay correct, got %+v", repeat)
	}
	if repeat.Points != 0 {
		t.Fatalf("expected repeated correct submission not to award points, got %+v", repeat)
	}

	var count int64
	if err := db.Model(&model.Submission{}).
		Where("user_id = ? AND challenge_id = ?", 71, 11).
		Count(&count).Error; err != nil {
		t.Fatalf("count submissions: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected repeated correct submission not to create extra record, got %d", count)
	}
}

func TestSubmitFlagShrinksOwnedInstanceExpiryAfterSolve(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	flagSalt := "solve-grace-salt"

	if err := db.Create(&model.User{
		ID:        7,
		Username:  "student7",
		Role:      model.RoleStudent,
		Status:    model.UserStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	originalExpiry := now.Add(2 * time.Hour)
	if err := db.Create(&model.Instance{
		ID:          1001,
		UserID:      7,
		ChallengeID: 11,
		Status:      model.InstanceStatusRunning,
		ShareScope:  model.InstanceSharingPerUser,
		ExpiresAt:   originalExpiry,
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	}).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	service := NewService(
		practiceinfra.NewRepository(db),
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:              id,
					Category:        model.DimensionWeb,
					Points:          100,
					Status:          model.ChallengeStatusPublished,
					FlagType:        model.FlagTypeStatic,
					FlagSalt:        flagSalt,
					FlagHash:        flagcrypto.HashStaticFlag("flag{correct}", flagSalt),
					InstanceSharing: model.InstanceSharingPerUser,
				}, nil
			},
		},
		nil,
		runtimeinfrarepo.NewRepository(db),
		nil,
		nil,
		redisClient,
		&config.Config{
			RateLimit: config.RateLimitConfig{
				RedisKeyPrefix: "practice:test",
				FlagSubmit: config.RateLimitPolicyConfig{
					Limit:  5,
					Window: time.Minute,
				},
			},
			Container: config.ContainerConfig{
				SolveGracePeriod: 10 * time.Minute,
			},
		},
		nil)

	beforeSubmit := time.Now()
	resp, err := service.SubmitFlag(context.Background(), 7, 11, "flag{correct}")
	if err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !resp.IsCorrect {
		t.Fatalf("expected correct submission response, got %+v", resp)
	}
	if resp.InstanceShutdownAt == nil {
		t.Fatalf("expected shutdown hint, got %+v", resp)
	}
	if resp.Message != "" {
		t.Fatalf("expected practice submit message to be omitted, got %q", resp.Message)
	}

	expectedMax := beforeSubmit.Add(10*time.Minute + 5*time.Second)
	expectedMin := beforeSubmit.Add(9*time.Minute + 50*time.Second)
	if resp.InstanceShutdownAt.Before(expectedMin) || resp.InstanceShutdownAt.After(expectedMax) {
		t.Fatalf("unexpected shutdown time: got %v, want around %v", resp.InstanceShutdownAt, beforeSubmit.Add(10*time.Minute))
	}

	var stored model.Instance
	if err := db.First(&stored, 1001).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if !stored.ExpiresAt.Equal(*resp.InstanceShutdownAt) {
		t.Fatalf("expected instance expiry to match response: stored=%v response=%v", stored.ExpiresAt, *resp.InstanceShutdownAt)
	}
	if !stored.ExpiresAt.Before(originalExpiry) {
		t.Fatalf("expected instance expiry to shrink: before=%v after=%v", originalExpiry, stored.ExpiresAt)
	}
}

func TestListMyChallengeSubmissionsMapsStoredHistory(t *testing.T) {
	t.Parallel()

	now := time.Now()
	service := NewService(
		&stubPracticeRepository{
			listChallengeSubmissionsFn: func(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error) {
				if userID != 7 || challengeID != 11 {
					t.Fatalf("unexpected query: user=%d challenge=%d", userID, challengeID)
				}
				if limit <= 0 {
					t.Fatalf("expected positive limit, got %d", limit)
				}
				return []model.Submission{
					{
						ID:           3,
						UserID:       7,
						ChallengeID:  11,
						IsCorrect:    true,
						ReviewStatus: model.SubmissionReviewStatusNotRequired,
						SubmittedAt:  now.Add(-time.Minute),
					},
					{
						ID:           2,
						UserID:       7,
						ChallengeID:  11,
						IsCorrect:    false,
						ReviewStatus: model.SubmissionReviewStatusPending,
						Flag:         "answer with reasoning",
						SubmittedAt:  now.Add(-2 * time.Minute),
					},
					{
						ID:           1,
						UserID:       7,
						ChallengeID:  11,
						IsCorrect:    false,
						ReviewStatus: model.SubmissionReviewStatusNotRequired,
						SubmittedAt:  now.Add(-3 * time.Minute),
					},
				}, nil
			},
		},
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:     id,
					Status: model.ChallengeStatusPublished,
				}, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
		nil)

	items, err := service.ListMyChallengeSubmissions(context.Background(), 7, 11)
	if err != nil {
		t.Fatalf("ListMyChallengeSubmissions() error = %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 records, got %d", len(items))
	}
	if items[0].Status != dto.SubmissionStatusCorrect {
		t.Fatalf("unexpected correct record: %+v", items[0])
	}
	if items[1].Status != dto.SubmissionStatusPendingReview {
		t.Fatalf("unexpected pending record: %+v", items[1])
	}
	if items[1].Answer != "answer with reasoning" {
		t.Fatalf("expected manual review answer to be preserved, got %+v", items[1])
	}
	if items[2].Status != dto.SubmissionStatusIncorrect {
		t.Fatalf("unexpected incorrect record: %+v", items[2])
	}
}

func TestSubmitFlagRejectsUnknownFlagType(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	service := NewService(
		practiceinfra.NewRepository(db),
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:       id,
					Category: model.DimensionWeb,
					Points:   100,
					Status:   model.ChallengeStatusPublished,
					FlagType: "shared_proof",
				}, nil
			},
		},
		nil,
		&stubPracticeInstanceStore{},
		nil,
		nil,
		redisClient,
		&config.Config{
			RateLimit: config.RateLimitConfig{
				RedisKeyPrefix: "practice:test",
				FlagSubmit: config.RateLimitPolicyConfig{
					Limit:  5,
					Window: time.Minute,
				},
			},
		},
		nil)

	_, err := service.SubmitFlag(context.Background(), 7, 11, "flag{legacy}")
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params for unknown flag type, got %v", err)
	}
}

func TestSubmitFlagPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("submit")
	expectedCtxValue := "ctx-submit-flag"
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()
	flagSalt := "context-submit-salt"

	findCorrectCalled := false
	createSubmissionCalled := false
	challengeLookupCalled := false
	repo := &stubPracticeRepository{
		findCorrectSubmissionFn: func(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
			findCorrectCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-correct ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil, gorm.ErrRecordNotFound
		},
		createSubmissionFn: func(ctx context.Context, submission *model.Submission) error {
			createSubmissionCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected create-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				challengeLookupCalled = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected challenge lookup ctx value %v, got %v", expectedCtxValue, got)
				}
				return &model.Challenge{
					ID:       id,
					Category: model.DimensionWeb,
					Points:   100,
					Status:   model.ChallengeStatusPublished,
					FlagType: model.FlagTypeStatic,
					FlagSalt: flagSalt,
					FlagHash: flagcrypto.HashStaticFlag("flag{ctx-submit}", flagSalt),
				}, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		redisClient,
		&config.Config{
			RateLimit: config.RateLimitConfig{
				RedisKeyPrefix: "practice:test",
				FlagSubmit: config.RateLimitPolicyConfig{
					Limit:  5,
					Window: time.Minute,
				},
			},
		},
		nil)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.SubmitFlag(ctx, 7, 11, "flag{ctx-submit}"); err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !challengeLookupCalled {
		t.Fatal("expected challenge lookup to be called")
	}
	if !findCorrectCalled {
		t.Fatal("expected find correct submission repository to be called")
	}
	if !createSubmissionCalled {
		t.Fatal("expected create submission repository to be called")
	}
}

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
				Submission: model.Submission{
					ID:           id,
					UserID:       88,
					ChallengeID:  11,
					Flag:         "answer",
					ReviewStatus: model.SubmissionReviewStatusPending,
					SubmittedAt:  now,
					UpdatedAt:    now,
				},
				StudentUsername: "student88",
				StudentName:     "Student 88",
				ClassName:       "Class A",
				ChallengeTitle:  "manual challenge",
			}, nil
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			findRequesterCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.User{ID: userID, Role: model.RoleTeacher, ClassName: "Class A"}, nil
		},
		updateSubmissionFn: func(ctx context.Context, submission *model.Submission) error {
			updatedCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update-submission ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				challengeLookupCalled = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected challenge lookup ctx value %v, got %v", expectedCtxValue, got)
				}
				return &model.Challenge{
					ID:       id,
					Category: model.DimensionWeb,
					Points:   120,
					Status:   model.ChallengeStatusPublished,
					FlagType: model.FlagTypeManualReview,
				}, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
		nil)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.ReviewManualReviewSubmission(
		ctx,
		91,
		1001,
		model.RoleTeacher,
		&dto.ReviewManualReviewSubmissionReq{ReviewStatus: model.SubmissionReviewStatusApproved},
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
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.User{ID: userID, Role: model.RoleTeacher, ClassName: "Class A"}, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
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
	service := NewService(repo, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.ListTeacherManualReviewSubmissions(ctx, 1001, model.RoleTeacher, &dto.TeacherManualReviewSubmissionQuery{}); err != nil {
		t.Fatalf("ListTeacherManualReviewSubmissions() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected list manual review repository to be called")
	}
}

func TestListTeacherManualReviewSubmissionsRejectsStudentRole(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			t.Fatal("did not expect requester lookup for student role")
			return nil, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			t.Fatal("did not expect list repository call for student role")
			return nil, 0, nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.ListTeacherManualReviewSubmissions(context.Background(), 1001, model.RoleStudent, &dto.TeacherManualReviewSubmissionQuery{})
	if err == nil {
		t.Fatal("expected student role to be rejected")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrForbidden.Code {
		t.Fatalf("expected forbidden error, got %v", err)
	}
}

func TestListTeacherManualReviewSubmissionsRejectsInvalidReviewStatus(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			t.Fatal("did not expect requester lookup for invalid review status")
			return nil, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			t.Fatal("did not expect list repository call for invalid review status")
			return nil, 0, nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.ListTeacherManualReviewSubmissions(
		context.Background(),
		1001,
		model.RoleTeacher,
		&dto.TeacherManualReviewSubmissionQuery{ReviewStatus: "archived"},
	)
	if err == nil {
		t.Fatal("expected invalid review status to be rejected")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestListTeacherManualReviewSubmissionsRejectsOversizedPageSize(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			t.Fatal("did not expect requester lookup for oversized page size")
			return nil, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			t.Fatal("did not expect list repository call for oversized page size")
			return nil, 0, nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.ListTeacherManualReviewSubmissions(
		context.Background(),
		1001,
		model.RoleTeacher,
		&dto.TeacherManualReviewSubmissionQuery{Size: 101},
	)
	if err == nil {
		t.Fatal("expected oversized page size to be rejected")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestListTeacherManualReviewSubmissionsRejectsNonPositiveStudentID(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			t.Fatal("did not expect requester lookup for non-positive student id")
			return nil, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			t.Fatal("did not expect list repository call for non-positive student id")
			return nil, 0, nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)
	studentID := int64(0)

	_, err := service.ListTeacherManualReviewSubmissions(
		context.Background(),
		1001,
		model.RoleTeacher,
		&dto.TeacherManualReviewSubmissionQuery{StudentID: &studentID},
	)
	if err == nil {
		t.Fatal("expected non-positive student id to be rejected")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestListTeacherManualReviewSubmissionsRejectsNonPositiveChallengeID(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			t.Fatal("did not expect requester lookup for non-positive challenge id")
			return nil, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			t.Fatal("did not expect list repository call for non-positive challenge id")
			return nil, 0, nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)
	challengeID := int64(0)

	_, err := service.ListTeacherManualReviewSubmissions(
		context.Background(),
		1001,
		model.RoleTeacher,
		&dto.TeacherManualReviewSubmissionQuery{ChallengeID: &challengeID},
	)
	if err == nil {
		t.Fatal("expected non-positive challenge id to be rejected")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestListTeacherManualReviewSubmissionsRejectsOversizedClassName(t *testing.T) {
	t.Parallel()

	repo := &stubPracticeRepository{
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			t.Fatal("did not expect requester lookup for oversized class name")
			return nil, nil
		},
		listTeacherManualReviewSubmissionsFn: func(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
			t.Fatal("did not expect list repository call for oversized class name")
			return nil, 0, nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.ListTeacherManualReviewSubmissions(
		context.Background(),
		1001,
		model.RoleAdmin,
		&dto.TeacherManualReviewSubmissionQuery{ClassName: strings.Repeat("A", 129)},
	)
	if err == nil {
		t.Fatal("expected oversized class name to be rejected")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrInvalidParams.Code {
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
				Submission:      model.Submission{ID: id, UserID: 88, ChallengeID: 11, ReviewStatus: model.SubmissionReviewStatusPending, SubmittedAt: now, UpdatedAt: now},
				StudentUsername: "student88",
				StudentName:     "Student 88",
				ClassName:       "Class A",
				ChallengeTitle:  "manual challenge",
			}, nil
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			findRequesterCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.User{ID: userID, Role: model.RoleTeacher, ClassName: "Class A"}, nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.GetTeacherManualReviewSubmission(ctx, 91, 1001, model.RoleTeacher); err != nil {
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
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			t.Fatal("did not expect requester lookup for student role")
			return nil, nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.GetTeacherManualReviewSubmission(context.Background(), 91, 1001, model.RoleStudent)
	if err == nil {
		t.Fatal("expected student role to be rejected")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrForbidden.Code {
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
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			t.Fatal("did not expect requester lookup for student role")
			return nil, nil
		},
		updateSubmissionFn: func(ctx context.Context, submission *model.Submission) error {
			t.Fatal("did not expect submission update for student role")
			return nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.ReviewManualReviewSubmission(
		context.Background(),
		91,
		1001,
		model.RoleStudent,
		&dto.ReviewManualReviewSubmissionReq{ReviewStatus: model.SubmissionReviewStatusApproved},
	)
	if err == nil {
		t.Fatal("expected student role to be rejected")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrForbidden.Code {
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
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			t.Fatal("did not expect requester lookup for invalid review status")
			return nil, nil
		},
		updateSubmissionFn: func(ctx context.Context, submission *model.Submission) error {
			t.Fatal("did not expect submission update for invalid review status")
			return nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.ReviewManualReviewSubmission(
		context.Background(),
		91,
		1001,
		model.RoleTeacher,
		&dto.ReviewManualReviewSubmissionReq{ReviewStatus: model.SubmissionReviewStatusPending},
	)
	if err == nil {
		t.Fatal("expected invalid review status to be rejected")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrInvalidParams.Code {
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
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			t.Fatal("did not expect requester lookup for oversized review comment")
			return nil, nil
		},
		updateSubmissionFn: func(ctx context.Context, submission *model.Submission) error {
			t.Fatal("did not expect submission update for oversized review comment")
			return nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	_, err := service.ReviewManualReviewSubmission(
		context.Background(),
		91,
		1001,
		model.RoleTeacher,
		&dto.ReviewManualReviewSubmissionReq{
			ReviewStatus:  model.SubmissionReviewStatusApproved,
			ReviewComment: strings.Repeat("a", 4001),
		},
	)
	if err == nil {
		t.Fatal("expected oversized review comment to be rejected")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrInvalidParams.Code {
		t.Fatalf("expected invalid params error, got %v", err)
	}
}

func TestReviewManualReviewSubmissionRejectsApprovalAfterChallengeAlreadySolved(t *testing.T) {
	t.Parallel()

	now := time.Now()
	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDFn: func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
			return &practiceports.TeacherManualReviewSubmissionRecord{
				Submission: model.Submission{
					ID:           id,
					UserID:       88,
					ChallengeID:  11,
					Flag:         "answer",
					ReviewStatus: model.SubmissionReviewStatusPending,
					SubmittedAt:  now,
					UpdatedAt:    now,
				},
				StudentUsername: "student88",
				StudentName:     "Student 88",
				ClassName:       "Class A",
				ChallengeTitle:  "manual challenge",
			}, nil
		},
		findUserByIDFn: func(ctx context.Context, userID int64) (*model.User, error) {
			return &model.User{ID: userID, Role: model.RoleTeacher, ClassName: "Class A"}, nil
		},
		findCorrectSubmissionFn: func(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
			return &model.Submission{
				ID:           99,
				UserID:       userID,
				ChallengeID:  challengeID,
				IsCorrect:    true,
				ReviewStatus: model.SubmissionReviewStatusApproved,
				SubmittedAt:  now.Add(-time.Minute),
				UpdatedAt:    now.Add(-time.Minute),
			}, nil
		},
		updateSubmissionFn: func(ctx context.Context, submission *model.Submission) error {
			t.Fatal("did not expect submission update when challenge already solved")
			return nil
		},
	}
	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:       id,
					Category: model.DimensionWeb,
					Points:   120,
					Status:   model.ChallengeStatusPublished,
					FlagType: model.FlagTypeManualReview,
				}, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
		nil)

	_, err := service.ReviewManualReviewSubmission(
		context.Background(),
		91,
		1001,
		model.RoleTeacher,
		&dto.ReviewManualReviewSubmissionReq{ReviewStatus: model.SubmissionReviewStatusApproved},
	)
	if err == nil {
		t.Fatal("expected already solved approval to be rejected")
	}
	var appErr *errcode.AppError
	if !errors.As(err, &appErr) || appErr.Code != errcode.ErrAlreadySolved.Code {
		t.Fatalf("expected already solved error, got %v", err)
	}
}

func TestListMyChallengeSubmissionsPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("list-submissions")
	expectedCtxValue := "ctx-list-submissions"
	challengeLookupCalled := false
	listCalled := false
	service := NewService(
		&stubPracticeRepository{
			listChallengeSubmissionsFn: func(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error) {
				listCalled = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected submission listing ctx value %v, got %v", expectedCtxValue, got)
				}
				return []model.Submission{{ID: 1, UserID: userID, ChallengeID: challengeID, SubmittedAt: time.Now()}}, nil
			},
		},
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				challengeLookupCalled = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected challenge lookup ctx value %v, got %v", expectedCtxValue, got)
				}
				return &model.Challenge{ID: id, Status: model.ChallengeStatusPublished}, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
		nil)

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

func TestSubmitFlagPropagatesContextToDynamicFlagInstanceLookup(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("dynamic-flag")
	expectedCtxValue := "ctx-dynamic-flag"
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()
	instanceLookupCalled := false
	instanceStore := &stubPracticeInstanceStore{
		findByUserAndChallengeWithContextFn: func(ctx context.Context, userID, challengeID int64) (*model.Instance, error) {
			instanceLookupCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected dynamic flag instance lookup ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Instance{ID: 301, UserID: userID, ChallengeID: challengeID, Nonce: "nonce-301"}, nil
		},
	}
	service := NewService(
		&stubPracticeRepository{
			findCorrectSubmissionFn: func(context.Context, int64, int64) (*model.Submission, error) {
				return nil, gorm.ErrRecordNotFound
			},
			createSubmissionFn: func(context.Context, *model.Submission) error {
				return nil
			},
		},
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:         id,
					Category:   model.DimensionWeb,
					Points:     100,
					Status:     model.ChallengeStatusPublished,
					FlagType:   model.FlagTypeDynamic,
					FlagPrefix: "flag",
				}, nil
			},
		},
		nil,
		instanceStore,
		nil,
		nil,
		redisClient,
		&config.Config{
			RateLimit: config.RateLimitConfig{
				RedisKeyPrefix: "practice:test",
				FlagSubmit:     config.RateLimitPolicyConfig{Limit: 5, Window: time.Minute},
			},
			Container: config.ContainerConfig{FlagGlobalSecret: "12345678901234567890123456789012"},
		},
		nil)

	flag := flagcrypto.GenerateDynamicFlag(7, 11, "12345678901234567890123456789012", "nonce-301", "flag")
	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.SubmitFlag(ctx, 7, 11, flag); err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !instanceLookupCalled {
		t.Fatal("expected dynamic flag instance lookup to be called")
	}
}

func TestSubmitFlagPropagatesContextToSolveGraceInstanceUpdates(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("solve-grace")
	expectedCtxValue := "ctx-solve-grace"
	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()
	lookupCalled := false
	refreshCalled := false
	instanceStore := &stubPracticeInstanceStore{
		findByUserAndChallengeWithContextFn: func(ctx context.Context, userID, challengeID int64) (*model.Instance, error) {
			lookupCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected solve grace lookup ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.Instance{ID: 401, UserID: userID, ChallengeID: challengeID, ShareScope: model.InstanceSharingPerUser, ExpiresAt: time.Now().Add(2 * time.Hour)}, nil
		},
		refreshInstanceExpiryWithContextFn: func(ctx context.Context, instanceID int64, expiresAt time.Time) error {
			refreshCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected solve grace refresh ctx value %v, got %v", expectedCtxValue, got)
			}
			if instanceID != 401 {
				t.Fatalf("unexpected instance id: %d", instanceID)
			}
			return nil
		},
	}
	flagSalt := "solve-grace-ctx"
	service := NewService(
		&stubPracticeRepository{
			findCorrectSubmissionFn: func(context.Context, int64, int64) (*model.Submission, error) {
				return nil, gorm.ErrRecordNotFound
			},
			createSubmissionFn: func(context.Context, *model.Submission) error {
				return nil
			},
		},
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:              id,
					Category:        model.DimensionWeb,
					Points:          100,
					Status:          model.ChallengeStatusPublished,
					FlagType:        model.FlagTypeStatic,
					FlagSalt:        flagSalt,
					FlagHash:        flagcrypto.HashStaticFlag("flag{solve-grace-ctx}", flagSalt),
					InstanceSharing: model.InstanceSharingPerUser,
				}, nil
			},
		},
		nil,
		instanceStore,
		nil,
		nil,
		redisClient,
		&config.Config{
			RateLimit: config.RateLimitConfig{
				RedisKeyPrefix: "practice:test",
				FlagSubmit:     config.RateLimitPolicyConfig{Limit: 5, Window: time.Minute},
			},
			Container: config.ContainerConfig{SolveGracePeriod: 10 * time.Minute},
		},
		nil)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.SubmitFlag(ctx, 7, 11, "flag{solve-grace-ctx}"); err != nil {
		t.Fatalf("SubmitFlag() error = %v", err)
	}
	if !lookupCalled {
		t.Fatal("expected solve grace instance lookup to be called")
	}
	if !refreshCalled {
		t.Fatal("expected solve grace refresh to be called")
	}
}
