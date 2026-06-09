package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	"ctf-platform/internal/shared/taxonomy"
	"sync/atomic"
	"testing"
	"time"
)

func TestStartChallengeQueuesProvisioningWithoutSynchronousContainerCreation(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{
		ID:        101,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&practiceCommandChallengeRow{
		ID:         201,
		Title:      "Queued Web",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		Points:     100,
		ImageID:    101,
		Status:     challengecontracts.ChallengeStatusPublished,
		FlagType:   challengecontracts.FlagTypeStatic,
		FlagHash:   "flag{static}",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&identitycontracts.User{ID: 42, Username: "student-42", Role: identitycontracts.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	var createCalls atomic.Int32
	service := wirePracticeScopeAdapters(NewService(
		newPracticeRepositoryWithRuntimePortOwner(db),

		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, _ int64) (string, string, int, int, error) {
				createCalls.Add(1)
				return "container-sync", "network-sync", reservedHostPort, 8080, nil
			},
		},
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
		nil),

		newPracticeRepositoryWithRuntimePortOwner(db), challengeinfra.NewRepository(db))

	resp, err := service.StartChallenge(context.Background(), 42, 201)
	if err != nil {
		t.Fatalf("StartChallenge() error = %v", err)
	}
	if resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected pending status, got %+v", resp)
	}
	if createCalls.Load() != 0 {
		t.Fatalf("expected no synchronous container creation, got %d calls", createCalls.Load())
	}

	var stored instanceentity.Instance
	if err := db.First(&stored, resp.ID).Error; err != nil {
		t.Fatalf("load pending instance: %v", err)
	}
	if stored.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected stored pending instance, got %+v", stored)
	}
}

func TestStartChallengePersistsSelectedRuntimeNodeID(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{
		ID:        111,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&practiceCommandChallengeRow{
		ID:         211,
		Title:      "Node Bound Web",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		Points:     100,
		ImageID:    111,
		Status:     challengecontracts.ChallengeStatusPublished,
		FlagType:   challengecontracts.FlagTypeStatic,
		FlagHash:   "flag{static}",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&identitycontracts.User{ID: 52, Username: "student-52", Role: identitycontracts.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := wirePracticeScopeAdapters(NewService(
		newPracticeRepositoryWithRuntimePortOwner(db),
		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{},
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
		nil).
		SetRuntimeNodeSelector(&stubPracticeRuntimeNodeSelector{
			selectRuntimeNodeFn: func(ctx context.Context, scope practiceports.InstanceScope) (*practiceports.RuntimeNodeBinding, error) {
				return &practiceports.RuntimeNodeBinding{NodeID: 901, NodeName: "node-901"}, nil
			},
		}),
		newPracticeRepositoryWithRuntimePortOwner(db), challengeinfra.NewRepository(db))

	resp, err := service.StartChallenge(context.Background(), 52, 211)
	if err != nil {
		t.Fatalf("StartChallenge() error = %v", err)
	}

	var stored instanceentity.Instance
	if err := db.First(&stored, resp.ID).Error; err != nil {
		t.Fatalf("load pending instance: %v", err)
	}
	if stored.NodeID == nil || *stored.NodeID != 901 {
		t.Fatalf("expected persisted runtime node id 901, got %+v", stored.NodeID)
	}
}

func TestStartChallengeIgnoresExpiredRunningInstance(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now().UTC()
	if err := db.Create(&practiceCommandImageRow{
		ID:        106,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&practiceCommandChallengeRow{
		ID:         206,
		Title:      "Expired Runtime",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		Points:     100,
		ImageID:    106,
		Status:     challengecontracts.ChallengeStatusPublished,
		FlagType:   challengecontracts.FlagTypeStatic,
		FlagHash:   "flag{static}",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&identitycontracts.User{ID: 46, Username: "student-46", Role: identitycontracts.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&instanceentity.Instance{
		ID:          9006,
		UserID:      46,
		ChallengeID: 206,
		HostPort:    30000,
		ContainerID: "expired-runtime",
		Status:      instanceentity.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30000",
		ExpiresAt:   now.Add(-2 * time.Minute),
		MaxExtends:  2,
		CreatedAt:   now.Add(-time.Hour),
		UpdatedAt:   now.Add(-time.Hour),
	}).Error; err != nil {
		t.Fatalf("create expired instance: %v", err)
	}

	service := wirePracticeScopeAdapters(NewService(
		newPracticeRepositoryWithRuntimePortOwner(db),

		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{},
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
		nil),

		newPracticeRepositoryWithRuntimePortOwner(db), challengeinfra.NewRepository(db))

	resp, err := service.StartChallenge(context.Background(), 46, 206)
	if err != nil {
		t.Fatalf("StartChallenge() error = %v", err)
	}
	if resp.ID == 9006 {
		t.Fatalf("expected expired instance to be replaced, got reused instance %+v", resp)
	}
	if resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected pending status for restarted instance, got %+v", resp)
	}

	var instances []instanceentity.Instance
	if err := db.Order("id asc").Find(&instances).Error; err != nil {
		t.Fatalf("list instances: %v", err)
	}
	if len(instances) != 2 {
		t.Fatalf("expected expired instance and restarted instance, got %+v", instances)
	}
}

func TestStartChallengePropagatesContextToTransactionalRepositoryWhenReusingSharedInstance(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("tx-reuse")
	expectedCtxValue := "ctx-tx-reuse"
	lockCalled := false
	findExistingCalled := false
	refreshCalled := false
	repo := &stubPracticeRepository{
		lockInstanceScopeFn: func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) error {
			lockCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected lock ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
		findScopedExistingInstanceFn: func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			findExistingCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-existing ctx value %v, got %v", expectedCtxValue, got)
			}
			return &instanceentity.Instance{ID: 901, UserID: 7, ChallengeID: challengeID, ShareScope: instanceentity.ShareScopeShared, Status: instanceentity.InstanceStatusRunning, ExpiresAt: time.Now().Add(5 * time.Minute), MaxExtends: 2}, nil
		},
		refreshInstanceExpiryFn: func(instanceID int64, expiresAt time.Time) error {
			t.Fatalf("expected context-aware expiry refresh, got legacy call")
			return nil
		},
		refreshInstanceExpiryWithContextFn: func(ctx context.Context, instanceID int64, expiresAt time.Time) error {
			refreshCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected refresh ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
	}
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{ID: id, ImageID: 1, Status: challengecontracts.ChallengeStatusPublished, FlagType: challengecontracts.FlagTypeStatic, FlagHash: "flag{shared}", InstanceSharing: challengecontracts.InstanceSharingShared}, nil
		},
		findChallengeTopologyByChallengeIDFn: func(context.Context, int64) (*challengecontracts.PracticeRuntimeChallengeTopology, error) {
			return nil, nil
		},
	}
	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{Container: config.ContainerConfig{DefaultTTL: time.Hour, MaxConcurrentPerUser: 3}},
		nil),

		repo, challengeRepo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.StartChallenge(ctx, 7, 11)
	if err != nil {
		t.Fatalf("StartChallenge() error = %v", err)
	}
	if resp == nil || resp.ID != 901 {
		t.Fatalf("expected reused instance 901, got %+v", resp)
	}
	if !lockCalled || !findExistingCalled || !refreshCalled {
		t.Fatalf("expected lock/find/refresh to be called, got lock=%v find=%v refresh=%v", lockCalled, findExistingCalled, refreshCalled)
	}
}

func TestStartChallengeReusesStoppingInstanceInsteadOfCreatingNewOne(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	stoppingExpiresAt := now.Add(time.Hour)
	stoppingUpdatedAt := now.Add(-time.Minute)
	if err := db.Create(&practiceCommandImageRow{
		ID:        107,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&practiceCommandChallengeRow{
		ID:         207,
		Title:      "Stopping Reuse",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		Points:     100,
		ImageID:    107,
		Status:     challengecontracts.ChallengeStatusPublished,
		FlagType:   challengecontracts.FlagTypeStatic,
		FlagHash:   "flag{static}",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&identitycontracts.User{ID: 47, Username: "student-47", Role: identitycontracts.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&instanceentity.Instance{
		ID:          9007,
		UserID:      47,
		ChallengeID: 207,
		Status:      instanceentity.InstanceStatusStopping,
		ExpiresAt:   stoppingExpiresAt,
		MaxExtends:  2,
		CreatedAt:   now.Add(-time.Minute),
		UpdatedAt:   stoppingUpdatedAt,
	}).Error; err != nil {
		t.Fatalf("create stopping instance: %v", err)
	}

	service := wirePracticeScopeAdapters(NewService(
		newPracticeRepositoryWithRuntimePortOwner(db),

		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{},
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
		nil),

		newPracticeRepositoryWithRuntimePortOwner(db), challengeinfra.NewRepository(db))

	resp, err := service.StartChallenge(context.Background(), 47, 207)
	if err != nil {
		t.Fatalf("StartChallenge() error = %v", err)
	}
	if resp.ID != 9007 {
		t.Fatalf("expected stopping instance to be reused, got %+v", resp)
	}
	if resp.Status != "destroying" {
		t.Fatalf("expected stopping instance to be exposed as destroying, got %+v", resp)
	}
	if resp.AccessURL != "" || resp.Access != nil {
		t.Fatalf("expected destroying response access to be cleared, got %+v", resp)
	}

	var count int64
	if err := db.Model(&instanceentity.Instance{}).Where("user_id = ? AND challenge_id = ?", 47, 207).Count(&count).Error; err != nil {
		t.Fatalf("count instances: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected no replacement instance to be created, got %d rows", count)
	}

	stored, err := runtimeinfrarepo.NewRepository(db).FindByID(context.Background(), 9007)
	if err != nil {
		t.Fatalf("load stored stopping instance: %v", err)
	}
	if !stored.ExpiresAt.Equal(stoppingExpiresAt) {
		t.Fatalf("expected stopping instance expiry to remain unchanged, got %s want %s", stored.ExpiresAt, stoppingExpiresAt)
	}
	if !stored.UpdatedAt.Equal(stoppingUpdatedAt) {
		t.Fatalf("expected stopping instance updated_at to remain unchanged, got %s want %s", stored.UpdatedAt, stoppingUpdatedAt)
	}
}

func TestStartChallengePropagatesContextToTransactionalRepositoryWhenCreatingInstance(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("tx-create")
	expectedCtxValue := "ctx-tx-create"
	countCalled := false
	reserveCalled := false
	createCalled := false
	bindCalled := false
	repo := &stubPracticeRepository{
		lockInstanceScopeFn: func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) error {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected lock ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil
		},
		findScopedExistingInstanceFn: func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected find-existing ctx value %v, got %v", expectedCtxValue, got)
			}
			return nil, nil
		},
		countScopedRunningInstancesFn: func(ctx context.Context, userID int64, scope practiceports.InstanceScope) (int, error) {
			countCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected count ctx value %v, got %v", expectedCtxValue, got)
			}
			return 0, nil
		},
		reserveAvailablePortFn: func(ctx context.Context, start, end int) (int, error) {
			reserveCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected reserve-port ctx value %v, got %v", expectedCtxValue, got)
			}
			return 30007, nil
		},
		createInstanceFn: func(ctx context.Context, instance *instanceentity.Instance) error {
			createCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected create-instance ctx value %v, got %v", expectedCtxValue, got)
			}
			instance.ID = 902
			return nil
		},
		bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
			bindCalled = true
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected bind-port ctx value %v, got %v", expectedCtxValue, got)
			}
			if port != 30007 || instanceID != 902 {
				t.Fatalf("unexpected bind args port=%d instanceID=%d", port, instanceID)
			}
			return nil
		},
	}
	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{ID: id, ImageID: 1, Status: challengecontracts.ChallengeStatusPublished, FlagType: challengecontracts.FlagTypeStatic, FlagHash: "flag{new}"}, nil
		},
		findChallengeTopologyByChallengeIDFn: func(context.Context, int64) (*challengecontracts.PracticeRuntimeChallengeTopology, error) {
			return nil, nil
		},
	}
	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{Container: config.ContainerConfig{DefaultTTL: time.Hour, MaxConcurrentPerUser: 3, MaxExtends: 2, Scheduler: config.ContainerSchedulerConfig{Enabled: true}}},
		nil),

		repo, challengeRepo)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	resp, err := service.StartChallenge(ctx, 7, 11)
	if err != nil {
		t.Fatalf("StartChallenge() error = %v", err)
	}
	if resp == nil || resp.ID != 902 {
		t.Fatalf("expected created instance 902, got %+v", resp)
	}
	if !countCalled || !reserveCalled || !createCalled || !bindCalled {
		t.Fatalf("expected count/reserve/create/bind to be called, got count=%v reserve=%v create=%v bind=%v", countCalled, reserveCalled, createCalled, bindCalled)
	}
}
