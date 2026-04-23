package commands

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
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

func reserveClosedLoopbackPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen loopback port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatalf("close loopback listener: %v", err)
	}
	return port
}

func parseHTTPServerEndpoint(t *testing.T, rawURL string) (string, int) {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}

	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		t.Fatalf("parse server port: %v", err)
	}
	return parsed.Hostname(), port
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
	findByIDFn                              func(id int64) (*model.Challenge, error)
	findByIDWithContextFn                   func(ctx context.Context, id int64) (*model.Challenge, error)
	findChallengeTopologyByChallengeIDFn    func(challengeID int64) (*model.ChallengeTopology, error)
	findChallengeTopologyByChallengeIDCtxFn func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
}

func (s *stubPracticeChallengeContract) FindByID(id int64) (*model.Challenge, error) {
	if s.findByIDFn == nil {
		return nil, nil
	}
	return s.findByIDFn(id)
}

func (s *stubPracticeChallengeContract) FindByIDWithContext(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return s.FindByID(id)
}

func (s *stubPracticeChallengeContract) FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDFn != nil {
		return s.findChallengeTopologyByChallengeIDFn(challengeID)
	}
	return nil, nil
}

func (s *stubPracticeChallengeContract) FindChallengeTopologyByChallengeIDWithContext(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDCtxFn != nil {
		return s.findChallengeTopologyByChallengeIDCtxFn(ctx, challengeID)
	}
	return s.FindChallengeTopologyByChallengeID(challengeID)
}

type stubPracticeImageStore struct {
	findByIDFn            func(id int64) (*model.Image, error)
	findByIDWithContextFn func(ctx context.Context, id int64) (*model.Image, error)
}

func (s *stubPracticeImageStore) FindByID(id int64) (*model.Image, error) {
	if s.findByIDFn == nil {
		return nil, nil
	}
	return s.findByIDFn(id)
}

func (s *stubPracticeImageStore) FindByIDWithContext(ctx context.Context, id int64) (*model.Image, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return s.FindByID(id)
}

type stubPracticeInstanceStore struct {
	findByIDWithContextFn                   func(ctx context.Context, id int64) (*model.Instance, error)
	refreshInstanceExpiryFn                 func(instanceID int64, expiresAt time.Time) error
	refreshInstanceExpiryWithContextFn      func(ctx context.Context, instanceID int64, expiresAt time.Time) error
	updateStatusAndReleasePortFn            func(id int64, status string) error
	updateStatusAndReleasePortWithContextFn func(ctx context.Context, id int64, status string) error
	findByUserAndChallengeFn                func(userID, challengeID int64) (*model.Instance, error)
	findByUserAndChallengeWithContextFn     func(ctx context.Context, userID, challengeID int64) (*model.Instance, error)
}

func (s *stubPracticeInstanceStore) FindByIDWithContext(ctx context.Context, id int64) (*model.Instance, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubPracticeInstanceStore) UpdateRuntime(instance *model.Instance) error {
	return nil
}

func (s *stubPracticeInstanceStore) RefreshInstanceExpiry(instanceID int64, expiresAt time.Time) error {
	if s.refreshInstanceExpiryFn != nil {
		return s.refreshInstanceExpiryFn(instanceID, expiresAt)
	}
	return nil
}

func (s *stubPracticeInstanceStore) RefreshInstanceExpiryWithContext(ctx context.Context, instanceID int64, expiresAt time.Time) error {
	if s.refreshInstanceExpiryWithContextFn != nil {
		return s.refreshInstanceExpiryWithContextFn(ctx, instanceID, expiresAt)
	}
	return s.RefreshInstanceExpiry(instanceID, expiresAt)
}

func (s *stubPracticeInstanceStore) UpdateStatusAndReleasePort(id int64, status string) error {
	if s.updateStatusAndReleasePortFn != nil {
		return s.updateStatusAndReleasePortFn(id, status)
	}
	return nil
}

func (s *stubPracticeInstanceStore) UpdateStatusAndReleasePortWithContext(ctx context.Context, id int64, status string) error {
	if s.updateStatusAndReleasePortWithContextFn != nil {
		return s.updateStatusAndReleasePortWithContextFn(ctx, id, status)
	}
	return s.UpdateStatusAndReleasePort(id, status)
}

func (s *stubPracticeInstanceStore) FindByUserAndChallenge(userID, challengeID int64) (*model.Instance, error) {
	if s.findByUserAndChallengeFn != nil {
		return s.findByUserAndChallengeFn(userID, challengeID)
	}
	return nil, nil
}

func (s *stubPracticeInstanceStore) FindByUserAndChallengeWithContext(ctx context.Context, userID, challengeID int64) (*model.Instance, error) {
	if s.findByUserAndChallengeWithContextFn != nil {
		return s.findByUserAndChallengeWithContextFn(ctx, userID, challengeID)
	}
	return s.FindByUserAndChallenge(userID, challengeID)
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

	request, err := service.buildTopologyCreateRequest(context.Background(), 30001, &model.Challenge{ImageID: 1}, "web", model.TopologySpec{
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

	_, err := service.buildTopologyCreateRequest(context.Background(), 30002, &model.Challenge{
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
		&stubPracticeRepository{
			findCorrectSubmissionWithContextFn: func(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
				return nil, gorm.ErrRecordNotFound
			},
			createSubmissionWithContextFn: func(ctx context.Context, submission *model.Submission) error {
				return nil
			},
		},
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
		&stubPracticeRepository{
			findCorrectSubmissionWithContextFn: func(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
				return nil, gorm.ErrRecordNotFound
			},
			createSubmissionWithContextFn: func(ctx context.Context, submission *model.Submission) error {
				return nil
			},
		},
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
			findByIDFn: func(id int64) (*model.Challenge, error) {
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

	resp, err := service.SubmitFlagWithContext(context.Background(), 7, 11, "flag{shared-static}")
	if err != nil {
		t.Fatalf("SubmitFlagWithContext() error = %v", err)
	}
	if !resp.IsCorrect || resp.Status != dto.SubmissionStatusCorrect {
		t.Fatalf("expected shared static submission success, got %+v", resp)
	}
}

func TestSubmitFlagWithContextAllowsRepeatCorrectSubmissionWithoutExtraPoints(t *testing.T) {
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
			findByIDFn: func(id int64) (*model.Challenge, error) {
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

	first, err := service.SubmitFlagWithContext(context.Background(), 71, 11, "flag{repeatable}")
	if err != nil {
		t.Fatalf("SubmitFlagWithContext() first error = %v", err)
	}
	if !first.IsCorrect || first.Points != 100 {
		t.Fatalf("expected first correct submission to score once, got %+v", first)
	}

	repeat, err := service.SubmitFlagWithContext(context.Background(), 71, 11, "flag{repeatable}")
	if err != nil {
		t.Fatalf("SubmitFlagWithContext() repeat error = %v", err)
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

func TestSubmitFlagWithContextShrinksOwnedInstanceExpiryAfterSolve(t *testing.T) {
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
			findByIDFn: func(id int64) (*model.Challenge, error) {
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
		nil,
	)

	beforeSubmit := time.Now()
	resp, err := service.SubmitFlagWithContext(context.Background(), 7, 11, "flag{correct}")
	if err != nil {
		t.Fatalf("SubmitFlagWithContext() error = %v", err)
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
			listChallengeSubmissionsFn: func(userID, challengeID int64, limit int) ([]model.Submission, error) {
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
			findByIDFn: func(id int64) (*model.Challenge, error) {
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
		nil,
		&config.Config{},
		nil,
	)

	items, err := service.ListMyChallengeSubmissions(7, 11)
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
			findByIDFn: func(id int64) (*model.Challenge, error) {
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

	_, err := service.SubmitFlagWithContext(context.Background(), 7, 11, "flag{legacy}")
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params for unknown flag type, got %v", err)
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

func TestStartContestAWDServiceDoesNotRequireContestChallengeLookup(t *testing.T) {
	t.Parallel()

	teamID := int64(4104)
	repo := &stubPracticeRepository{
		findContestByIDWithContextFn: func(ctx context.Context, contestID int64) (*model.Contest, error) {
			if contestID != 3104 {
				t.Fatalf("unexpected contest id: %d", contestID)
			}
			return &model.Contest{
				ID:     contestID,
				Mode:   model.ContestModeAWD,
				Status: model.ContestStatusRunning,
			}, nil
		},
		findContestAWDServiceWithContextFn: func(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error) {
			if contestID != 3104 || serviceID != 7104 {
				t.Fatalf("unexpected awd service lookup: contest=%d service=%d", contestID, serviceID)
			}
			return &model.ContestAWDService{
				ID:              serviceID,
				ContestID:       contestID,
				ChallengeID:     2104,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":104,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestChallengeWithContextFn: func(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error) {
			t.Fatalf("unexpected contest challenge lookup for awd start: contest=%d challenge=%d", contestID, challengeID)
			return nil, nil
		},
		findContestRegistrationWithContextFn: func(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
			if contestID != 3104 || userID != 5104 {
				t.Fatalf("unexpected registration lookup: contest=%d user=%d", contestID, userID)
			}
			return &model.ContestRegistration{
				ContestID: contestID,
				UserID:    userID,
				TeamID:    &teamID,
				Status:    model.ContestRegistrationStatusApproved,
			}, nil
		},
		createInstanceFn: func(instance *model.Instance) error {
			instance.ID = 9104
			return nil
		},
	}

	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDFn: func(id int64) (*model.Challenge, error) {
				if id != 2104 {
					t.Fatalf("unexpected challenge lookup: %d", id)
				}
				return &model.Challenge{
					ID:       id,
					Status:   model.ChallengeStatusPublished,
					ImageID:  104,
					FlagType: model.FlagTypeStatic,
					FlagHash: "flag{awd-static}",
				}, nil
			},
		},
		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
		nil,
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				PortRangeStart:       30000,
				PortRangeEnd:         30010,
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil,
	)

	resp, err := service.StartContestAWDService(context.Background(), 5104, 3104, 7104)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}
	if resp.ID != 9104 {
		t.Fatalf("expected created awd service instance id, got %+v", resp)
	}
	if resp.ChallengeID != 2104 {
		t.Fatalf("expected awd service challenge id 2104, got %+v", resp)
	}
	if resp.Status != model.InstanceStatusPending {
		t.Fatalf("expected pending awd service instance, got %+v", resp)
	}
}

func TestRunProvisioningLoopPromotesPendingInstanceToRunning(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(healthyServer.Close)
	publicHost, hostPort := parseHTTPServerEndpoint(t, healthyServer.URL)
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
				return "container-queued", "network-queued", hostPort, 8080, nil
			},
		},
		nil,
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				PortRangeStart:       hostPort,
				PortRangeEnd:         hostPort + 1,
				DefaultExposedPort:   8080,
				PublicHost:           publicHost,
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

func TestStartChallengeWithContextIgnoresExpiredRunningInstance(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{
		ID:        106,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&model.Challenge{
		ID:         206,
		Title:      "Expired Runtime",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		ImageID:    106,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		FlagHash:   "flag{static}",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&model.User{ID: 46, Username: "student-46", Role: model.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&model.Instance{
		ID:          9006,
		UserID:      46,
		ChallengeID: 206,
		HostPort:    30000,
		ContainerID: "expired-runtime",
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30000",
		ExpiresAt:   now.Add(-2 * time.Minute),
		MaxExtends:  2,
		CreatedAt:   now.Add(-time.Hour),
		UpdatedAt:   now.Add(-time.Hour),
	}).Error; err != nil {
		t.Fatalf("create expired instance: %v", err)
	}

	service := NewService(
		practiceinfra.NewRepository(db),
		challengeinfra.NewRepository(db),
		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{},
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
				MaxConcurrentPerUser: 1,
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

	resp, err := service.StartChallengeWithContext(context.Background(), 46, 206)
	if err != nil {
		t.Fatalf("StartChallengeWithContext() error = %v", err)
	}
	if resp.ID == 9006 {
		t.Fatalf("expected expired instance to be replaced, got reused instance %+v", resp)
	}
	if resp.Status != model.InstanceStatusPending {
		t.Fatalf("expected pending status for restarted instance, got %+v", resp)
	}

	var instances []model.Instance
	if err := db.Order("id asc").Find(&instances).Error; err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected expired instance and restarted instance, got %+v", instances)
	}
}

func TestProvisionInstanceMarksInstanceFailedWhenAccessURLIsNotReady(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{
		ID:        104,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	challenge := &model.Challenge{
		ID:         205,
		Title:      "Readiness Failure",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		Points:     100,
		ImageID:    104,
		Status:     model.ChallengeStatusPublished,
		FlagType:   model.FlagTypeStatic,
		FlagHash:   "flag{static}",
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	hostPort := reserveClosedLoopbackPort(t)
	instance := &model.Instance{
		UserID:      44,
		ChallengeID: challenge.ID,
		HostPort:    hostPort,
		Status:      model.InstanceStatusCreating,
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}

	var cleanupCalls atomic.Int32
	service := NewService(
		practiceinfra.NewRepository(db),
		challengeinfra.NewRepository(db),
		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(instance *model.Instance) error {
				cleanupCalls.Add(1)
				return nil
			},
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error) {
				return "container-readiness", "network-readiness", reservedHostPort, 8080, nil
			},
		},
		nil,
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				PublicHost:         "127.0.0.1",
				CreateTimeout:      time.Second,
				StartProbeTimeout:  50 * time.Millisecond,
				StartProbeInterval: 10 * time.Millisecond,
				StartProbeAttempts: 2,
			},
		},
		nil,
	)

	err := service.provisionInstance(context.Background(), instance, challenge, nil, "flag{static}")
	if err == nil || err.Error() != errcode.ErrContainerStartFailed.Error() {
		t.Fatalf("expected container start failed error, got %v", err)
	}

	var stored model.Instance
	if err := db.First(&stored, instance.ID).Error; err != nil {
		t.Fatalf("load failed instance: %v", err)
	}
	if stored.Status != model.InstanceStatusFailed {
		t.Fatalf("expected failed instance status, got %+v", stored)
	}
	if stored.AccessURL != "" {
		t.Fatalf("expected access url to stay empty after failed readiness, got %q", stored.AccessURL)
	}
	if cleanupCalls.Load() != 1 {
		t.Fatalf("expected cleanup to be called once, got %d", cleanupCalls.Load())
	}
}

func TestProvisionInstanceMarksInstanceFailedWithContext(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{
		ID:        105,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	ctxKey := practiceServiceContextKey("mark-failed")
	const expectedCtxValue = "practice-provision-failure"

	var markedFailed atomic.Int32
	service := NewService(
		nil,
		nil,
		challengeinfra.NewImageRepository(db),
		&stubPracticeInstanceStore{
			updateStatusAndReleasePortFn: func(id int64, status string) error {
				t.Fatalf("expected context-aware failed status update, got legacy call id=%d status=%s", id, status)
				return nil
			},
			updateStatusAndReleasePortWithContextFn: func(ctx context.Context, id int64, status string) error {
				markedFailed.Add(1)
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected failed status update ctx value %v, got %v", expectedCtxValue, got)
				}
				if id != 611 {
					t.Fatalf("expected failed instance id 611, got %d", id)
				}
				if status != model.InstanceStatusFailed {
					t.Fatalf("expected failed instance status %s, got %s", model.InstanceStatusFailed, status)
				}
				return nil
			},
		},
		&stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error) {
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected create container ctx value %v, got %v", expectedCtxValue, got)
				}
				return "ctr-ctx", "net-ctx", reservedHostPort, 8080, nil
			},
		},
		nil,
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				PublicHost:         "127.0.0.1",
				CreateTimeout:      time.Second,
				StartProbeTimeout:  20 * time.Millisecond,
				StartProbeInterval: 10 * time.Millisecond,
				StartProbeAttempts: 1,
			},
		},
		nil,
	)

	instance := &model.Instance{ID: 611, ChallengeID: 711, HostPort: reserveClosedLoopbackPort(t), Status: model.InstanceStatusCreating}
	challenge := &model.Challenge{ID: 711, ImageID: 105, Status: model.ChallengeStatusPublished}
	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)

	err := service.provisionInstance(ctx, instance, challenge, nil, "flag{ctx}")
	if err == nil || err.Error() != errcode.ErrContainerStartFailed.Error() {
		t.Fatalf("expected container start failed error, got %v", err)
	}
	if markedFailed.Load() != 1 {
		t.Fatalf("expected failed status update once, got %d", markedFailed.Load())
	}
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

type practiceServiceContextKey string

func TestLoadRuntimeSubjectWithScopePropagatesContextToChallengeContract(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("runtime-subject")
	expectedCtxValue := "ctx-runtime-subject"
	challengeLookupCalled := false
	topologyLookupCalled := false
	service := NewService(
		nil,
		&stubPracticeChallengeContract{
			findByIDFn: func(id int64) (*model.Challenge, error) {
				t.Fatalf("expected context-aware challenge lookup, got legacy call for %d", id)
				return nil, nil
			},
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				challengeLookupCalled = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected challenge lookup ctx value %v, got %v", expectedCtxValue, got)
				}
				return &model.Challenge{ID: id, Status: model.ChallengeStatusPublished}, nil
			},
			findChallengeTopologyByChallengeIDFn: func(challengeID int64) (*model.ChallengeTopology, error) {
				t.Fatalf("expected context-aware topology lookup, got legacy call for %d", challengeID)
				return nil, nil
			},
			findChallengeTopologyByChallengeIDCtxFn: func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
				topologyLookupCalled = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected topology lookup ctx value %v, got %v", expectedCtxValue, got)
				}
				return nil, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
		nil,
	)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	challenge, topology, err := service.loadRuntimeSubjectWithScope(ctx, practiceports.InstanceScope{}, 42)
	if err != nil {
		t.Fatalf("loadRuntimeSubjectWithScope() error = %v", err)
	}
	if challenge == nil || challenge.ID != 42 {
		t.Fatalf("expected challenge 42, got %+v", challenge)
	}
	if topology != nil {
		t.Fatalf("expected nil topology, got %+v", topology)
	}
	if !challengeLookupCalled {
		t.Fatal("expected challenge lookup to be called")
	}
	if !topologyLookupCalled {
		t.Fatal("expected topology lookup to be called")
	}
}

func TestBuildTopologyCreateRequestPropagatesContextToImageRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("topology-image")
	expectedCtxValue := "ctx-topology-image"
	lookups := make([]int64, 0, 2)
	service := &Service{
		imageRepo: &stubPracticeImageStore{
			findByIDFn: func(id int64) (*model.Image, error) {
				t.Fatalf("expected context-aware image lookup, got legacy call for %d", id)
				return nil, nil
			},
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Image, error) {
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected image lookup ctx value %v, got %v", expectedCtxValue, got)
				}
				lookups = append(lookups, id)
				return &model.Image{ID: id, Name: fmt.Sprintf("repo/%d", id), Tag: "latest", Status: model.ImageStatusAvailable}, nil
			},
		},
		config: &config.Config{},
	}

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	request, err := service.buildTopologyCreateRequest(ctx, 30001, &model.Challenge{ImageID: 1}, "web", model.TopologySpec{
		Nodes: []model.TopologyNode{
			{Key: "web", Name: "Web", ServicePort: 8080},
			{Key: "worker", Name: "Worker", ImageID: 2, ServicePort: 9000},
		},
	}, "flag{ctx-image}")
	if err != nil {
		t.Fatalf("buildTopologyCreateRequest() error = %v", err)
	}
	if len(request.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %+v", request.Nodes)
	}
	if len(lookups) != 2 || lookups[0] != 1 || lookups[1] != 2 {
		t.Fatalf("expected image lookups [1 2], got %v", lookups)
	}
}

func TestSubmitFlagWithContextPropagatesContextToRepository(t *testing.T) {
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
		findCorrectSubmissionWithContextFn: func(ctx context.Context, userID, challengeID int64) (*model.Submission, error) {
			findCorrectCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-correct ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil, gorm.ErrRecordNotFound
		},
		createSubmissionWithContextFn: func(ctx context.Context, submission *model.Submission) error {
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
			findByIDFn: func(id int64) (*model.Challenge, error) {
				t.Fatalf("expected context-aware challenge lookup, got legacy call for %d", id)
				return nil, nil
			},
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

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.SubmitFlagWithContext(ctx, 7, 11, "flag{ctx-submit}"); err != nil {
		t.Fatalf("SubmitFlagWithContext() error = %v", err)
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

func TestReviewManualReviewSubmissionWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("review")
	expectedCtxValue := "ctx-review-manual"
	now := time.Now()
	updatedCalled := false
	findRequesterCalled := false
	findRecordCalled := false
	challengeLookupCalled := false
	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDWithContextFn: func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
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
		findUserByIDWithContextFn: func(ctx context.Context, userID int64) (*model.User, error) {
			findRequesterCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.User{ID: userID, Role: model.RoleTeacher, ClassName: "Class A"}, nil
		},
		updateSubmissionWithContextFn: func(ctx context.Context, submission *model.Submission) error {
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
			findByIDFn: func(id int64) (*model.Challenge, error) {
				t.Fatalf("expected context-aware challenge lookup, got legacy call for %d", id)
				return nil, nil
			},
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
		nil,
		&config.Config{},
		nil,
	)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.ReviewManualReviewSubmissionWithContext(
		ctx,
		91,
		1001,
		model.RoleTeacher,
		&dto.ReviewManualReviewSubmissionReq{ReviewStatus: model.SubmissionReviewStatusApproved},
	); err != nil {
		t.Fatalf("ReviewManualReviewSubmissionWithContext() error = %v", err)
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

func TestListTeacherManualReviewSubmissionsWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("list-review")
	expectedCtxValue := "ctx-list-review"
	listCalled := false
	repo := &stubPracticeRepository{
		findUserByIDWithContextFn: func(ctx context.Context, userID int64) (*model.User, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.User{ID: userID, Role: model.RoleTeacher, ClassName: "Class A"}, nil
		},
		listTeacherManualReviewSubmissionsWithContextFn: func(ctx context.Context, query *dto.TeacherManualReviewSubmissionQuery) ([]practiceports.TeacherManualReviewSubmissionRecord, int64, error) {
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
	service := NewService(repo, nil, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.ListTeacherManualReviewSubmissionsWithContext(ctx, 1001, model.RoleTeacher, &dto.TeacherManualReviewSubmissionQuery{}); err != nil {
		t.Fatalf("ListTeacherManualReviewSubmissionsWithContext() error = %v", err)
	}
	if !listCalled {
		t.Fatal("expected list manual review repository to be called")
	}
}

func TestGetTeacherManualReviewSubmissionWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("get-review")
	expectedCtxValue := "ctx-get-review"
	now := time.Now()
	getCalled := false
	findRequesterCalled := false
	repo := &stubPracticeRepository{
		getTeacherManualReviewSubmissionByIDWithContextFn: func(ctx context.Context, id int64) (*practiceports.TeacherManualReviewSubmissionRecord, error) {
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
		findUserByIDWithContextFn: func(ctx context.Context, userID int64) (*model.User, error) {
			findRequesterCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-user ctx value %v, got %v", expectedCtxValue, got)
			}
			return &model.User{ID: userID, Role: model.RoleTeacher, ClassName: "Class A"}, nil
		},
	}
	service := NewService(repo, nil, nil, nil, nil, nil, nil, nil, &config.Config{}, nil)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.GetTeacherManualReviewSubmissionWithContext(ctx, 91, 1001, model.RoleTeacher); err != nil {
		t.Fatalf("GetTeacherManualReviewSubmissionWithContext() error = %v", err)
	}
	if !getCalled {
		t.Fatal("expected get manual review repository to be called")
	}
	if !findRequesterCalled {
		t.Fatal("expected requester repository to be called")
	}
}

func TestListMyChallengeSubmissionsWithContextPropagatesContextToRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("list-submissions")
	expectedCtxValue := "ctx-list-submissions"
	challengeLookupCalled := false
	listCalled := false
	service := NewService(
		&stubPracticeRepository{
			listChallengeSubmissionsFn: func(userID, challengeID int64, limit int) ([]model.Submission, error) {
				t.Fatalf("expected context-aware submission listing, got legacy call user=%d challenge=%d limit=%d", userID, challengeID, limit)
				return nil, nil
			},
			listChallengeSubmissionsWithContextFn: func(ctx context.Context, userID, challengeID int64, limit int) ([]model.Submission, error) {
				listCalled = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected submission listing ctx value %v, got %v", expectedCtxValue, got)
				}
				return []model.Submission{{ID: 1, UserID: userID, ChallengeID: challengeID, SubmittedAt: time.Now()}}, nil
			},
		},
		&stubPracticeChallengeContract{
			findByIDFn: func(id int64) (*model.Challenge, error) {
				t.Fatalf("expected context-aware challenge lookup, got legacy call for %d", id)
				return nil, nil
			},
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
		nil,
		&config.Config{},
		nil,
	)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	items, err := service.ListMyChallengeSubmissionsWithContext(ctx, 7, 11)
	if err != nil {
		t.Fatalf("ListMyChallengeSubmissionsWithContext() error = %v", err)
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

func TestSubmitFlagWithContextPropagatesContextToDynamicFlagInstanceLookup(t *testing.T) {
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
			findCorrectSubmissionWithContextFn: func(context.Context, int64, int64) (*model.Submission, error) {
				return nil, gorm.ErrRecordNotFound
			},
			createSubmissionWithContextFn: func(context.Context, *model.Submission) error {
				return nil
			},
		},
		&stubPracticeChallengeContract{
			findByIDFn: func(id int64) (*model.Challenge, error) {
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
		nil,
		redisClient,
		&config.Config{
			RateLimit: config.RateLimitConfig{
				RedisKeyPrefix: "practice:test",
				FlagSubmit:     config.RateLimitPolicyConfig{Limit: 5, Window: time.Minute},
			},
			Container: config.ContainerConfig{FlagGlobalSecret: "12345678901234567890123456789012"},
		},
		nil,
	)

	flag := flagcrypto.GenerateDynamicFlag(7, 11, "12345678901234567890123456789012", "nonce-301", "flag")
	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.SubmitFlagWithContext(ctx, 7, 11, flag); err != nil {
		t.Fatalf("SubmitFlagWithContext() error = %v", err)
	}
	if !instanceLookupCalled {
		t.Fatal("expected dynamic flag instance lookup to be called")
	}
}

func TestSubmitFlagWithContextPropagatesContextToSolveGraceInstanceUpdates(t *testing.T) {
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
			findCorrectSubmissionWithContextFn: func(context.Context, int64, int64) (*model.Submission, error) {
				return nil, gorm.ErrRecordNotFound
			},
			createSubmissionWithContextFn: func(context.Context, *model.Submission) error {
				return nil
			},
		},
		&stubPracticeChallengeContract{
			findByIDFn: func(id int64) (*model.Challenge, error) {
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
		nil,
		redisClient,
		&config.Config{
			RateLimit: config.RateLimitConfig{
				RedisKeyPrefix: "practice:test",
				FlagSubmit:     config.RateLimitPolicyConfig{Limit: 5, Window: time.Minute},
			},
			Container: config.ContainerConfig{SolveGracePeriod: 10 * time.Minute},
		},
		nil,
	)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	if _, err := service.SubmitFlagWithContext(ctx, 7, 11, "flag{solve-grace-ctx}"); err != nil {
		t.Fatalf("SubmitFlagWithContext() error = %v", err)
	}
	if !lookupCalled {
		t.Fatal("expected solve grace instance lookup to be called")
	}
	if !refreshCalled {
		t.Fatal("expected solve grace refresh to be called")
	}
}
