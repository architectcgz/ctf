package practice

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	practicecontracts "ctf-platform/internal/module/practice/contracts"
	"ctf-platform/internal/platform/events"
	flagcrypto "ctf-platform/pkg/crypto"
)

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

func (s *stubPracticeChallengeContract) FindHintByLevel(challengeID int64, level int) (*model.ChallengeHint, error) {
	return nil, nil
}

func (s *stubPracticeChallengeContract) CreateHintUnlock(unlock *model.ChallengeHintUnlock) error {
	return nil
}

func (s *stubPracticeChallengeContract) FindChallengeTopologyByChallengeID(challengeID int64) (*model.ChallengeTopology, error) {
	return nil, nil
}

type stubPracticeInstanceStore struct {
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

func TestBuildTopologyCreateRequestKeepsFineGrainedPolicies(t *testing.T) {
	db := setupPracticeTestDB(t)
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
	repo := NewRepository(db)
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

func setupPracticeTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Image{}); err != nil {
		t.Fatalf("migrate image: %v", err)
	}
	return db
}
