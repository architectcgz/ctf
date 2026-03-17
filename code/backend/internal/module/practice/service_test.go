package practice

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	"ctf-platform/internal/module/challenge"
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

func TestBuildTopologyCreateRequestKeepsFineGrainedPolicies(t *testing.T) {
	db := setupPracticeTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	service := &Service{
		imageRepo: challenge.NewImageRepository(db),
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
