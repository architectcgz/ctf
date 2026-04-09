package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	"ctf-platform/internal/platform/events"
	flagcrypto "ctf-platform/pkg/crypto"
	"ctf-platform/pkg/errcode"
)

type stubPracticeRuntimeService struct {
	cleanupRuntimeFn  func(instance *model.Instance) error
	createTopologyFn  func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error)
	createContainerFn func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
}

func (s *stubPracticeRuntimeService) CleanupRuntime(instance *model.Instance) error {
	if s.cleanupRuntimeFn == nil {
		return nil
	}
	return s.cleanupRuntimeFn(instance)
}

func (s *stubPracticeRuntimeService) CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if s.createTopologyFn == nil {
		return nil, errors.New("unexpected CreateTopology call")
	}
	return s.createTopologyFn(ctx, req)
}

func (s *stubPracticeRuntimeService) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error) {
	if s.createContainerFn == nil {
		return "", "", 0, 0, errors.New("unexpected CreateContainer call")
	}
	return s.createContainerFn(ctx, imageName, env, reservedHostPort)
}

func requireEventually(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition was not satisfied before timeout")
}

func newPracticeCommandTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("%s/%s.sqlite", t.TempDir(), strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Image{},
		&model.Challenge{},
		&model.ChallengeTopology{},
		&model.User{},
		&model.Team{},
		&model.Instance{},
		&model.PortAllocation{},
		&model.Submission{},
	); err != nil {
		t.Fatalf("migrate practice command tables: %v", err)
	}
	return db
}

type stubAssessmentService struct {
	updateFn func(ctx context.Context, userID int64, dimension string) error
}

func (s *stubAssessmentService) UpdateSkillProfileForDimension(ctx context.Context, userID int64, dimension string) error {
	if s.updateFn == nil {
		return nil
	}
	return s.updateFn(ctx, userID, dimension)
}

type stubScoreUpdater struct {
	updateFn func(ctx context.Context, userID int64) error
	lockWait time.Duration
}

func (s *stubScoreUpdater) UpdateUserScoreWithContext(ctx context.Context, userID int64) error {
	if s.updateFn == nil {
		return nil
	}
	return s.updateFn(ctx, userID)
}

func (s *stubScoreUpdater) lockTimeout() time.Duration {
	return s.lockWait
}

type stubPracticeChallengeContract struct {
	findByIDFn func(id int64) (*model.Challenge, error)
}

func (s *stubPracticeChallengeContract) FindByID(id int64) (*model.Challenge, error) {
	if s.findByIDFn == nil {
		return nil, nil
	}
	return s.findByIDFn(id)
}

func (s *stubPracticeChallengeContract) FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error) {
	return nil, nil
}

type stubPracticeInstanceStore struct {
}

func (s *stubPracticeInstanceStore) FindByIDWithContext(ctx context.Context, id int64) (*model.Instance, error) {
	return nil, nil
}

func (s *stubPracticeInstanceStore) UpdateRuntime(instance *model.Instance) error {
	return nil
}

func (s *stubPracticeInstanceStore) UpdateStatusAndReleasePort(id int64, status string) error {
	return nil
}

func (s *stubPracticeInstanceStore) FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error) {
	return nil, nil
}

func (s *stubPracticeInstanceStore) ListPendingInstancesWithContext(ctx context.Context, limit int) ([]*model.Instance, error) {
	return []*model.Instance{}, nil
}

func (s *stubPracticeInstanceStore) TryTransitionStatusWithContext(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error) {
	return false, nil
}

func (s *stubPracticeInstanceStore) CountInstancesByStatusWithContext(ctx context.Context, statuses []string) (int64, error) {
	return 0, nil
}

func TestBuildTopologyCreateRequestKeepsFineGrainedPolicies(t *testing.T) {
	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	service := &Service{
		imageRepo: challengeinfra.NewImageRepository(db),
		config:    &config.Config{},
	}

	request, err := service.buildTopologyCreateRequest(30001, &model.Challenge{ImageID: 1}, "web", model.TopologySpec{
		Nodes: []model.TopologyNode{
			{Key: "web", ServicePort: 8080, InjectFlag: true},
		},
		Policies: []model.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "web", Action: model.TopologyPolicyActionAllow, Protocol: model.TopologyPolicyProtocolTCP, Ports: []int{8080}},
		},
	}, "flag{demo}")
	if err != nil {
		t.Fatalf("buildTopologyCreateRequest() error = %v", err)
	}
	if len(request.Policies) != 1 {
		t.Fatalf("expected fine-grained policy to be kept, got %+v", request.Policies)
	}
	if request.Policies[0].Protocol != model.TopologyPolicyProtocolTCP {
		t.Fatalf("unexpected policy protocol: %+v", request.Policies[0])
	}
}

func TestBuildTopologyCreateRequestRejectsSharedChallengeFlagInjection(t *testing.T) {
	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 2, Name: "ctf/web", Tag: "v2", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	service := &Service{
		imageRepo: challengeinfra.NewImageRepository(db),
		config:    &config.Config{},
	}

	_, err := service.buildTopologyCreateRequest(30002, &model.Challenge{
		ImageID:         2,
		InstanceSharing: model.InstanceSharingShared,
	}, "web", model.TopologySpec{
		Nodes: []model.TopologyNode{
			{Key: "web", ServicePort: 8080, InjectFlag: true},
		},
	}, "flag{demo}")
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params for shared topology flag injection, got %v", err)
	}
}

func TestPracticeServiceCloseCancelsAssessmentUpdate(t *testing.T) {
	t.Parallel()

	startedCh := make(chan struct{})
	var calls atomic.Int32
	service := NewService(
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&stubAssessmentService{
			updateFn: func(ctx context.Context, userID int64, dimension string) error {
				if userID != 42 {
					t.Fatalf("unexpected userID: %d", userID)
				}
				if dimension != model.DimensionWeb {
					t.Fatalf("unexpected dimension: %s", dimension)
				}
				calls.Add(1)
				close(startedCh)
				<-ctx.Done()
				return ctx.Err()
			},
		},
		nil,
		&config.Config{
			Assessment: config.AssessmentConfig{
				IncrementalUpdateDelay:   0,
				IncrementalUpdateTimeout: time.Minute,
			},
		},
		nil,
	)

	service.triggerAssessmentUpdate(42, model.DimensionWeb)

	select {
	case <-startedCh:
	case <-time.After(time.Second):
		t.Fatal("expected assessment update to start")
	}

	closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := service.Close(closeCtx); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one assessment update call, got %d", calls.Load())
	}
}

func TestPracticeServiceCloseCancelsAsyncScoreUpdate(t *testing.T) {
	t.Parallel()

	startedCh := make(chan struct{})
	var calls atomic.Int32
	service := NewService(
		nil,
		nil,
		nil,
		nil,
		nil,
		&stubScoreUpdater{
			lockWait: time.Minute,
			updateFn: func(ctx context.Context, userID int64) error {
				if userID != 7 {
					t.Fatalf("unexpected userID: %d", userID)
				}
				calls.Add(1)
				close(startedCh)
				<-ctx.Done()
				return ctx.Err()
			},
		},
		nil,
		nil,
		&config.Config{},
		nil,
	)

	service.triggerScoreUpdate(7)

	select {
	case <-startedCh:
	case <-time.After(time.Second):
		t.Fatal("expected score update to start")
	}

	closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := service.Close(closeCtx); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected one score update call, got %d", calls.Load())
	}
}

func TestSubmitFlagWithRegexChallengeMatchesPattern(t *testing.T) {
	t.Parallel()

	redisServer := miniredis.RunT(t)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	defer redisClient.Close()

	repo := &stubPracticeRepository{}
	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDFn: func(id int64) (*model.Challenge, error) {
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
		nil,
	)

	resp, err := service.SubmitFlagWithContext(context.Background(), 9, 19, "flag{regex-42}")
	if err != nil {
		t.Fatalf("SubmitFlagWithContext() error = %v", err)
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
		createSubmissionFn: func(submission *model.Submission) error {
			createdSubmission = submission
			return nil
		},
	}
	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDFn: func(id int64) (*model.Challenge, error) {
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
		nil,
	)

	resp, err := service.SubmitFlagWithContext(context.Background(), 8, 18, "answer with reasoning")
	if err != nil {
		t.Fatalf("SubmitFlagWithContext() error = %v", err)
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
		getTeacherManualReviewSubmissionByIDFn: func(id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
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
		updateSubmissionFn: func(submission *model.Submission) error {
			updatedSubmission = submission
			return nil
		},
		findUserByIDFn: func(userID int64) (*model.User, error) {
			return &model.User{ID: userID, Username: "teacher", Role: model.RoleTeacher, ClassName: "Class 1"}, nil
		},
	}
	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDFn: func(id int64) (*model.Challenge, error) {
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
		nil,
	)

	resp, err := service.ReviewManualReviewSubmissionWithContext(
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
		t.Fatalf("ReviewManualReviewSubmissionWithContext() error = %v", err)
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

func TestPracticeServiceRunAsyncTaskReturnsWhenClosed(t *testing.T) {
	t.Parallel()

	service := NewService(nil, nil, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)
	closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := service.Close(closeCtx); err != nil {
		t.Fatalf("initial Close() error = %v", err)
	}

	called := make(chan struct{}, 1)
	service.runAsyncTask(func(context.Context) {
		called <- struct{}{}
	})

	select {
	case <-called:
		t.Fatal("expected closed service to skip async task")
	case <-time.After(50 * time.Millisecond):
	}
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
		findCorrectSubmissionFn: func(userID, challengeID int64) (*model.Submission, error) {
			return nil, gorm.ErrRecordNotFound
		},
		createSubmissionFn: func(submission *model.Submission) error {
			return db.Create(submission).Error
		},
	}
	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDFn: func(id int64) (*model.Challenge, error) {
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
		nil,
	)
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

	resp, err := service.SubmitFlagWithContext(context.Background(), 7, 11, "flag{correct}")
	if err != nil {
		t.Fatalf("SubmitFlagWithContext() error = %v", err)
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

func TestStartChallengeQueuesProvisioningWithoutSynchronousContainerCreation(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{
		ID:        101,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&model.Challenge{
		ID:         201,
		Title:      "Queued Web",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		ImageID:    101,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		FlagHash:   "flag{static}",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&model.User{ID: 42, Username: "student-42", Role: model.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	var createCalls atomic.Int32
	service := NewService(
		practiceinfra.NewRepository(db),
		challengeinfra.NewRepository(db),
		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error) {
				createCalls.Add(1)
				return "container-sync", "network-sync", reservedHostPort, 8080, nil
			},
		},
		nil,
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				PortRangeStart:       30000,
				PortRangeEnd:         30010,
				DefaultExposedPort:   8080,
				PublicHost:           "127.0.0.1",
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				CreateTimeout:        time.Second,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled:             true,
					PollInterval:        10 * time.Millisecond,
					BatchSize:           1,
					MaxConcurrentStarts: 1,
					MaxActiveInstances:  10,
				},
			},
		},
		nil,
	)

	resp, err := service.StartChallengeWithContext(context.Background(), 42, 201)
	if err != nil {
		t.Fatalf("StartChallengeWithContext() error = %v", err)
	}
	if resp.Status != model.InstanceStatusPending {
		t.Fatalf("expected pending status, got %+v", resp)
	}
	if createCalls.Load() != 0 {
		t.Fatalf("expected no synchronous container creation, got %d calls", createCalls.Load())
	}

	var stored model.Instance
	if err := db.First(&stored, resp.ID).Error; err != nil {
		t.Fatalf("load pending instance: %v", err)
	}
	if stored.Status != model.InstanceStatusPending {
		t.Fatalf("expected stored pending instance, got %+v", stored)
	}
}

func TestRunProvisioningLoopPromotesPendingInstanceToRunning(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{
		ID:        102,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&model.Challenge{
		ID:         202,
		Title:      "Queued Runner",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		ImageID:    102,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		FlagHash:   "flag{static}",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&model.User{ID: 43, Username: "student-43", Role: model.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := NewService(
		practiceinfra.NewRepository(db),
		challengeinfra.NewRepository(db),
		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error) {
				return "container-queued", "network-queued", reservedHostPort, 8080, nil
			},
		},
		nil,
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				PortRangeStart:       30000,
				PortRangeEnd:         30010,
				DefaultExposedPort:   8080,
				PublicHost:           "127.0.0.1",
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				CreateTimeout:        time.Second,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled:             true,
					PollInterval:        10 * time.Millisecond,
					BatchSize:           1,
					MaxConcurrentStarts: 1,
					MaxActiveInstances:  10,
				},
			},
		},
		nil,
	)

	resp, err := service.StartChallengeWithContext(context.Background(), 43, 202)
	if err != nil {
		t.Fatalf("StartChallengeWithContext() error = %v", err)
	}
	if resp.Status != model.InstanceStatusPending {
		t.Fatalf("expected pending status, got %+v", resp)
	}

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.RunProvisioningLoop(runCtx)

	requireEventually(t, time.Second, func() bool {
		var instance model.Instance
		if err := db.First(&instance, resp.ID).Error; err != nil {
			return false
		}
		return instance.Status == model.InstanceStatusRunning && instance.ContainerID == "container-queued"
	})
}

func TestRunProvisioningLoopLeavesOverflowPendingWhenGlobalCapacityReached(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{
		ID:        103,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	for _, challengeID := range []int64{203, 204} {
		if err := db.Create(&model.Challenge{
			ID:         challengeID,
			Title:      "Queued Capacity",
			Category:   model.DimensionWeb,
			Difficulty: model.ChallengeDifficultyEasy,
			Points:     100,
			ImageID:    103,
			Status:     model.ChallengeStatusPublished,
			FlagType:   model.FlagTypeStatic,
			FlagHash:   "flag{static}",
			CreatedAt:  now,
			UpdatedAt:  now,
		}).Error; err != nil {
			t.Fatalf("create challenge %d: %v", challengeID, err)
		}
	}
	for _, userID := range []int64{51, 52} {
		if err := db.Create(&model.User{ID: userID, Username: fmt.Sprintf("student-%d", userID), Role: model.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
			t.Fatalf("create user %d: %v", userID, err)
		}
	}

	started := make(chan int, 2)
	release := make(chan struct{})
	service := NewService(
		practiceinfra.NewRepository(db),
		challengeinfra.NewRepository(db),
		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error) {
				started <- reservedHostPort
				<-release
				return "container-capacity", "network-capacity", reservedHostPort, 8080, nil
			},
		},
		nil,
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				PortRangeStart:       30000,
				PortRangeEnd:         30010,
				DefaultExposedPort:   8080,
				PublicHost:           "127.0.0.1",
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				CreateTimeout:        time.Second,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled:             true,
					PollInterval:        10 * time.Millisecond,
					BatchSize:           2,
					MaxConcurrentStarts: 2,
					MaxActiveInstances:  1,
				},
			},
		},
		nil,
	)

	first, err := service.StartChallengeWithContext(context.Background(), 51, 203)
	if err != nil {
		t.Fatalf("StartChallengeWithContext() first error = %v", err)
	}
	second, err := service.StartChallengeWithContext(context.Background(), 52, 204)
	if err != nil {
		t.Fatalf("StartChallengeWithContext() second error = %v", err)
	}

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.RunProvisioningLoop(runCtx)

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("expected one pending instance to start provisioning")
	}

	var firstInstance model.Instance
	if err := db.First(&firstInstance, first.ID).Error; err != nil {
		t.Fatalf("load first instance: %v", err)
	}
	var secondInstance model.Instance
	if err := db.First(&secondInstance, second.ID).Error; err != nil {
		t.Fatalf("load second instance: %v", err)
	}

	statuses := []string{firstInstance.Status, secondInstance.Status}
	pendingCount := 0
	creatingCount := 0
	for _, status := range statuses {
		if status == model.InstanceStatusPending {
			pendingCount++
		}
		if status == model.InstanceStatusCreating {
			creatingCount++
		}
	}
	if pendingCount != 1 || creatingCount != 1 {
		t.Fatalf("expected one creating and one pending instance, got %+v", statuses)
	}

	close(release)
}
