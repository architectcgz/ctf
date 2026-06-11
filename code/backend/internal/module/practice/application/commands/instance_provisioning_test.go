package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeentity "ctf-platform/internal/module/contest/entity"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
	flagcrypto "ctf-platform/internal/shared/flagcrypto"
	"ctf-platform/internal/shared/taxonomy"
	"fmt"
	"go.uber.org/zap"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

type stubPracticeSchedulerLockStore struct {
	acquired bool
	err      error
}

type stubPracticeSchedulerLockLease struct{}

func TestBuildProvisioningFlagUsesInstanceFlagKeyID(t *testing.T) {
	t.Parallel()

	service := newServiceCore(nil, nil, nil, nil, nil, nil, &config.Config{
		Container: config.ContainerConfig{
			FlagGlobalSecret:        "active-secret-12345678901234567890",
			ResolvedFlagSecretKeyID: "active",
			ResolvedFlagSecrets: map[string]string{
				"active":   "active-secret-12345678901234567890",
				"previous": "previous-secret-123456789012345678",
			},
		},
	}, nil)

	flag, err := service.buildProvisioningFlag(&instancecontracts.Instance{
		UserID:      7,
		ChallengeID: 11,
		Nonce:       "nonce-provision",
		FlagKeyID:   "previous",
	}, toPracticeChallenge(&challengecontracts.PracticeRuntimeChallenge{
		ID:         11,
		FlagType:   challengecontracts.FlagTypeDynamic,
		FlagPrefix: "flag",
	}))
	if err != nil {
		t.Fatalf("buildProvisioningFlag() error = %v", err)
	}

	expected := flagcrypto.GenerateDynamicFlag(7, 11, "previous-secret-123456789012345678", "nonce-provision", "flag")
	if flag != expected {
		t.Fatalf("flag = %q, want %q", flag, expected)
	}
}

func (s *stubPracticeSchedulerLockStore) AcquireProvisioningSchedulerLock(context.Context, time.Duration) (practiceports.PracticeSchedulerLockLease, bool, error) {
	if s.err != nil {
		return nil, false, s.err
	}
	if !s.acquired {
		return nil, false, nil
	}
	return stubPracticeSchedulerLockLease{}, true, nil
}

func (stubPracticeSchedulerLockLease) Key(context.Context) string { return "practice-scheduler-lock" }

func (stubPracticeSchedulerLockLease) Release(context.Context) (bool, error) { return true, nil }

func (stubPracticeSchedulerLockLease) Refresh(context.Context, time.Duration) (bool, error) {
	return true, nil
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
	if err := db.Create(&practiceCommandImageRow{
		ID:        102,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&practiceCommandChallengeRow{
		ID:         202,
		Title:      "Queued Runner",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		Points:     100,
		ImageID:    102,
		Status:     challengecontracts.ChallengeStatusPublished,
		FlagType:   challengecontracts.FlagTypeStatic,
		FlagHash:   "flag{static}",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&identitycontracts.User{ID: 43, Username: "student-43", Role: identitycontracts.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	service := wirePracticeScopeAdapters(newServiceCore(
		newPracticeRepositoryWithRuntimePortOwner(db),

		challengeinfra.NewImageRepository(db),
		newPracticeTestInstanceRepository(db),
		&stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, _ int64) (string, string, int, int, error) {
				return "container-queued", "network-queued", hostPort, 8080, nil
			},
		},
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
		nil),

		newPracticeRepositoryWithRuntimePortOwner(db), challengeinfra.NewRepository(db)).
		SetInstanceReadinessProbe(practiceinfra.NewInstanceReadinessProbe())

	service.StartBackgroundTasks(context.Background())

	resp, err := service.StartChallenge(context.Background(), 43, 202)
	if err != nil {
		t.Fatalf("StartChallenge() error = %v", err)
	}
	if resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected pending status, got %+v", resp)
	}

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.RunProvisioningLoop(runCtx)

	requireEventually(t, time.Second, func() bool {
		var instance instanceentity.Instance
		if err := db.First(&instance, resp.ID).Error; err != nil {
			return false
		}
		return instance.Status == instanceentity.InstanceStatusRunning && instance.ContainerID == "container-queued"
	})
}

func TestRunProvisioningLoopSkipsWorkWhenSchedulerLockHeldByOtherReplica(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	healthyServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	t.Cleanup(healthyServer.Close)
	publicHost, hostPort := parseHTTPServerEndpoint(t, healthyServer.URL)
	if err := db.Create(&practiceCommandImageRow{
		ID:        112,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&practiceCommandChallengeRow{
		ID:         212,
		Title:      "Queued Standby",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		Points:     100,
		ImageID:    112,
		Status:     challengecontracts.ChallengeStatusPublished,
		FlagType:   challengecontracts.FlagTypeStatic,
		FlagHash:   "flag{static}",
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&identitycontracts.User{ID: 53, Username: "student-53", Role: identitycontracts.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	var createCalls atomic.Int32
	service := wirePracticeScopeAdapters(newServiceCore(
		newPracticeRepositoryWithRuntimePortOwner(db),

		challengeinfra.NewImageRepository(db),
		newPracticeTestInstanceRepository(db),
		&stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, _ int64) (string, string, int, int, error) {
				createCalls.Add(1)
				return "container-standby", "network-standby", hostPort, 8080, nil
			},
		},
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
					LockTTL:             time.Second,
				},
			},
		},
		nil),

		newPracticeRepositoryWithRuntimePortOwner(db), challengeinfra.NewRepository(db)).
		SetInstanceReadinessProbe(practiceinfra.NewInstanceReadinessProbe()).
		SetSchedulerLockStore(&stubPracticeSchedulerLockStore{acquired: false})

	service.StartBackgroundTasks(context.Background())

	resp, err := service.StartChallenge(context.Background(), 53, 212)
	if err != nil {
		t.Fatalf("StartChallenge() error = %v", err)
	}
	if resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected pending status, got %+v", resp)
	}

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.RunProvisioningLoop(runCtx)

	time.Sleep(120 * time.Millisecond)

	var instance instanceentity.Instance
	if err := db.First(&instance, resp.ID).Error; err != nil {
		t.Fatalf("load instance: %v", err)
	}
	if instance.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected standby replica to leave instance pending, got %s", instance.Status)
	}
	if createCalls.Load() != 0 {
		t.Fatalf("expected standby replica to skip runtime provisioning, got %d calls", createCalls.Load())
	}
}

func TestProvisionInstanceMarksInstanceFailedWhenAccessURLIsNotReady(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{
		ID:        104,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	challenge := &challengecontracts.PracticeRuntimeChallenge{
		ID:         205,
		Title:      "Readiness Failure",
		Category:   taxonomy.DimensionWeb,
		Difficulty: taxonomy.DifficultyEasy,
		Points:     100,
		ImageID:    104,
		Status:     challengecontracts.ChallengeStatusPublished,
		FlagType:   challengecontracts.FlagTypeStatic,
		FlagHash:   "flag{static}",
	}
	if err := db.Create(&practiceCommandChallengeRow{
		ID:         challenge.ID,
		Title:      challenge.Title,
		Category:   challenge.Category,
		Difficulty: challenge.Difficulty,
		Points:     challenge.Points,
		ImageID:    challenge.ImageID,
		Status:     challenge.Status,
		FlagType:   challenge.FlagType,
		FlagHash:   challenge.FlagHash,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	hostPort := reserveClosedLoopbackPort(t)
	instance := &instanceentity.Instance{
		UserID:      44,
		ChallengeID: challenge.ID,
		HostPort:    hostPort,
		Status:      instanceentity.InstanceStatusCreating,
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}

	var cleanupCalls atomic.Int32
	service := newServiceCore(
		newPracticeRepositoryWithRuntimePortOwner(db),

		challengeinfra.NewImageRepository(db),
		newPracticeTestInstanceRepository(db),
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(context.Context, *instanceentity.Instance) error {
				cleanupCalls.Add(1)
				return nil
			},
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, _ int64) (string, string, int, int, error) {
				return "container-readiness", "network-readiness", reservedHostPort, 8080, nil
			},
		},
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
		nil).
		SetInstanceReadinessProbe(practiceinfra.NewInstanceReadinessProbe())

	err := service.provisionInstance(context.Background(), instance, toPracticeChallenge(challenge), nil, "flag{static}")
	if err == nil || err.Error() != instancecontracts.ErrContainerStartFailed.Error() {
		t.Fatalf("expected container start failed error, got %v", err)
	}

	var stored instanceentity.Instance
	if err := db.First(&stored, instance.ID).Error; err != nil {
		t.Fatalf("load failed instance: %v", err)
	}
	if stored.Status != instanceentity.InstanceStatusFailed {
		t.Fatalf("expected failed instance status, got %+v", stored)
	}
	if stored.AccessURL != "" {
		t.Fatalf("expected access url to stay empty after failed readiness, got %q", stored.AccessURL)
	}
	if cleanupCalls.Load() != 1 {
		t.Fatalf("expected cleanup to be called once, got %d", cleanupCalls.Load())
	}
}

func TestProvisionInstancePropagatesContextToUpdateRuntime(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("update-runtime")
	expectedCtxValue := "ctx-update-runtime"
	instanceStore := &stubPracticeInstanceStore{
		updateRuntimeWithContextFn: func(ctx context.Context, instance *instanceentity.Instance) error {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update runtime ctx value %v, got %v", expectedCtxValue, got)
			}
			if instance.Status != instanceentity.InstanceStatusRunning {
				t.Fatalf("expected running status before persistence, got %+v", instance)
			}
			return nil
		},
	}
	service := newServiceCore(
		nil,

		&stubPracticeImageStore{
			findByIDFn: func(ctx context.Context, id int64) (*challengecontracts.Image, error) {
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected image lookup ctx value %v, got %v", expectedCtxValue, got)
				}
				return &challengecontracts.Image{ID: id, Name: "ctf/web", Tag: "v1", Status: challengecontracts.ImageStatusAvailable}, nil
			},
		},
		instanceStore,
		&stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, _ int64) (string, string, int, int, error) {
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected runtime create ctx value %v, got %v", expectedCtxValue, got)
				}
				return "ctr-running", "net-running", reservedHostPort, 8080, nil
			},
		},
		nil,
		nil,
		&config.Config{Container: config.ContainerConfig{PublicHost: "127.0.0.1", CreateTimeout: time.Second, StartProbeTimeout: 50 * time.Millisecond, StartProbeInterval: 10 * time.Millisecond, StartProbeAttempts: 1}},
		nil).
		SetInstanceReadinessProbe(practiceinfra.NewInstanceReadinessProbe())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	host, port := parseHTTPServerEndpoint(t, server.URL)
	instance := &instanceentity.Instance{ID: 951, ChallengeID: 2051, HostPort: port, Status: instanceentity.InstanceStatusCreating}
	challenge := &challengecontracts.PracticeRuntimeChallenge{ID: 2051, ImageID: 301, Status: challengecontracts.ChallengeStatusPublished, FlagType: challengecontracts.FlagTypeStatic, FlagHash: "flag{ok}"}
	service.config.Container.PublicHost = host
	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)

	if err := service.provisionInstance(ctx, instance, toPracticeChallenge(challenge), nil, "flag{ok}"); err != nil {
		t.Fatalf("provisionInstance() error = %v", err)
	}
}

func TestProvisionInstanceCleansRuntimeWhenInstanceLeavesCreatingBeforePersist(t *testing.T) {
	t.Parallel()

	var cleanupCalls atomic.Int32
	instanceStore := &stubPracticeInstanceStore{
		persistProvisionedRuntimeWithContextFn: func(ctx context.Context, instance *instanceentity.Instance) (bool, error) {
			if instance.Status != instanceentity.InstanceStatusRunning {
				t.Fatalf("expected running status before conditional persistence, got %+v", instance)
			}
			return false, nil
		},
	}
	service := newServiceCore(
		nil,
		&stubPracticeImageStore{
			findByIDFn: func(context.Context, int64) (*challengecontracts.Image, error) {
				return &challengecontracts.Image{ID: 302, Name: "ctf/web", Tag: "v1", Status: challengecontracts.ImageStatusAvailable}, nil
			},
		},
		instanceStore,
		&stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, _ int64) (string, string, int, int, error) {
				return "ctr-cancelled", "net-cancelled", reservedHostPort, 8080, nil
			},
			cleanupRuntimeFn: func(ctx context.Context, instance *instanceentity.Instance) error {
				cleanupCalls.Add(1)
				if instance.ContainerID != "ctr-cancelled" || instance.NetworkID != "net-cancelled" {
					t.Fatalf("expected cleanup of created runtime, got %+v", instance)
				}
				return nil
			},
		},
		nil,
		nil,
		&config.Config{Container: config.ContainerConfig{PublicHost: "127.0.0.1", CreateTimeout: time.Second, StartProbeTimeout: 50 * time.Millisecond, StartProbeInterval: 10 * time.Millisecond, StartProbeAttempts: 1}},
		nil).
		SetInstanceReadinessProbe(practiceinfra.NewInstanceReadinessProbe())

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	host, port := parseHTTPServerEndpoint(t, server.URL)
	service.config.Container.PublicHost = host

	instance := &instanceentity.Instance{ID: 953, ChallengeID: 2053, HostPort: port, Status: instanceentity.InstanceStatusCreating}
	challenge := &challengecontracts.PracticeRuntimeChallenge{ID: 2053, ImageID: 302, Status: challengecontracts.ChallengeStatusPublished, FlagType: challengecontracts.FlagTypeStatic, FlagHash: "flag{cancelled}"}

	if err := service.provisionInstance(context.Background(), instance, toPracticeChallenge(challenge), nil, "flag{cancelled}"); err != nil {
		t.Fatalf("provisionInstance() error = %v", err)
	}
	if cleanupCalls.Load() != 1 {
		t.Fatalf("expected created runtime to be cleaned once after leaving creating, got %d", cleanupCalls.Load())
	}
}

func TestMarkInstanceFailedSkipsFailedTransitionWhenInstanceLeavesCreating(t *testing.T) {
	t.Parallel()

	var cleanupCalls atomic.Int32
	finishCalled := false
	instanceStore := &stubPracticeInstanceStore{
		failProvisioningWithContextFn: func(ctx context.Context, id int64) (bool, error) {
			return false, nil
		},
		finishActiveAWDServiceOperationFn: func(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error {
			finishCalled = true
			return nil
		},
	}
	service := newServiceCore(
		nil,
		nil,
		instanceStore,
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, instance *instanceentity.Instance) error {
				cleanupCalls.Add(1)
				return nil
			},
		},
		nil,
		nil,
		&config.Config{},
		nil,
	)

	service.markInstanceFailed(context.Background(), &instanceentity.Instance{
		ID:          954,
		ChallengeID: 2054,
		ContainerID: "ctr-failed-cancelled",
		NetworkID:   "net-failed-cancelled",
		Status:      instanceentity.InstanceStatusCreating,
	})

	if cleanupCalls.Load() != 1 {
		t.Fatalf("expected failed runtime cleanup once, got %d", cleanupCalls.Load())
	}
	if finishCalled {
		t.Fatal("expected skipped failed transition to avoid finishing active AWD operation as failed")
	}
}

func TestProvisionInstanceAcceptsTCPAccessURLReadiness(t *testing.T) {
	t.Parallel()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer listener.Close()

	accepted := make(chan struct{}, 1)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		_ = conn.Close()
		accepted <- struct{}{}
	}()

	instanceStore := &stubPracticeInstanceStore{
		updateRuntimeWithContextFn: func(ctx context.Context, instance *instanceentity.Instance) error {
			if instance.Status != instanceentity.InstanceStatusRunning {
				t.Fatalf("expected running status, got %+v", instance)
			}
			if !strings.HasPrefix(instance.AccessURL, "tcp://") {
				t.Fatalf("expected tcp access url, got %q", instance.AccessURL)
			}
			return nil
		},
	}
	service := newServiceCore(
		nil,

		&stubPracticeImageStore{
			findByIDFn: func(context.Context, int64) (*challengecontracts.Image, error) {
				return &challengecontracts.Image{ID: 301, Name: "ctf/pwn", Tag: "v1", Status: challengecontracts.ImageStatusAvailable}, nil
			},
		},
		instanceStore,
		&stubPracticeRuntimeService{
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				if len(req.Nodes) != 1 {
					t.Fatalf("unexpected topology request: %+v", req)
				}
				if req.Nodes[0].ServiceProtocol != challengecontracts.ChallengeTargetProtocolTCP {
					t.Fatalf("expected tcp topology node, got %+v", req.Nodes[0])
				}
				return &practiceports.TopologyCreateResult{
					PrimaryContainerID: "pwn-ctr",
					NetworkID:          "pwn-net",
					AccessURL:          fmt.Sprintf("tcp://%s", listener.Addr().String()),
					RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
						Containers: []runtimecontracts.InstanceRuntimeContainer{
							{
								NodeKey:         "default",
								ContainerID:     "pwn-ctr",
								ServicePort:     8080,
								ServiceProtocol: challengecontracts.ChallengeTargetProtocolTCP,
								IsEntryPoint:    true,
								NetworkKeys:     []string{runtimecontracts.TopologyDefaultNetworkKey},
							},
						},
					},
				}, nil
			},
		},
		nil,
		nil,
		&config.Config{Container: config.ContainerConfig{PublicHost: "127.0.0.1", CreateTimeout: time.Second, StartProbeTimeout: 50 * time.Millisecond, StartProbeInterval: 10 * time.Millisecond, StartProbeAttempts: 2}},
		nil).
		SetInstanceReadinessProbe(practiceinfra.NewInstanceReadinessProbe())

	instance := &instanceentity.Instance{ID: 952, ChallengeID: 2052, HostPort: 0, Status: instanceentity.InstanceStatusCreating}
	challenge := &challengecontracts.PracticeRuntimeChallenge{
		ID:             2052,
		ImageID:        301,
		Status:         challengecontracts.ChallengeStatusPublished,
		FlagType:       challengecontracts.FlagTypeStatic,
		FlagHash:       "flag{ok}",
		TargetProtocol: challengecontracts.ChallengeTargetProtocolTCP,
	}

	if err := service.provisionInstance(context.Background(), instance, toPracticeChallenge(challenge), nil, "flag{ok}"); err != nil {
		t.Fatalf("provisionInstance() error = %v", err)
	}
	select {
	case <-accepted:
	case <-time.After(time.Second):
		t.Fatal("expected tcp readiness probe to connect")
	}
}

func TestProvisionAWDStableAliasSkipsHostReadinessProbe(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{
		ID:        502,
		Name:      "ctf/awd-web",
		Tag:       "v2",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(7002)
	teamID := int64(7102)
	serviceID := int64(8002)
	serviceSnapshot, err := contestentity.EncodeContestAWDServiceSnapshot(contestentity.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":         502,
			"instance_sharing": string(challengecontracts.InstanceSharingPerTeam),
			"defense_workspace": map[string]any{
				"entry_mode":      "ssh",
				"seed_root":       "runtime/workspace",
				"workspace_roots": []string{"runtime/workspace/app"},
				"writable_roots":  []string{"runtime/workspace/app"},
				"readonly_roots":  []string{},
				"runtime_mounts": []map[string]any{
					{"source": "runtime/workspace/app", "target": "/workspace/app", "mode": "rw"},
				},
			},
		},
		FlagConfig: map[string]any{
			"flag_type":   challengecontracts.FlagTypeStatic,
			"flag_prefix": "flag",
		},
	})
	if err != nil {
		t.Fatalf("encode service snapshot: %v", err)
	}
	service := &serviceCore{
		repo: &stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
				if gotContestID != contestID || gotServiceID != serviceID {
					t.Fatalf("unexpected awd service lookup: contest=%d service=%d", gotContestID, gotServiceID)
				}
				return &practiceports.ContestAWDServiceRecord{
					ID:              serviceID,
					ContestID:       contestID,
					AWDChallengeID:  502,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
		},
		imageRepo:    challengeinfra.NewImageRepository(db),
		instanceRepo: newPracticeTestInstanceRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				return &practiceports.TopologyCreateResult{
					PrimaryContainerID: "awd-alias-ctr",
					NetworkID:          "net-awd-contest-7002",
					AccessURL:          "http://awd-c7002-t7102-s8002:8080",
					RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
						Networks: []runtimecontracts.InstanceRuntimeNetwork{
							{Key: runtimecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-7002", NetworkID: "net-awd-contest-7002", Shared: true},
						},
						Containers: []runtimecontracts.InstanceRuntimeContainer{
							{NodeKey: "default", ContainerID: "awd-alias-ctr", ServicePort: 8080, IsEntryPoint: true, NetworkAliases: []string{"awd-c7002-t7102-s8002"}},
						},
					},
				}, nil
			},
		},
		config: &config.Config{
			Container: config.ContainerConfig{
				CreateTimeout:      time.Second,
				StartProbeTimeout:  10 * time.Millisecond,
				StartProbeInterval: 10 * time.Millisecond,
				StartProbeAttempts: 1,
			},
		},
		logger: zap.NewNop(),
	}
	instance := &instanceentity.Instance{
		ID:          9002,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 502,
		Status:      instanceentity.InstanceStatusCreating,
	}
	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}
	challenge := &challengecontracts.PracticeRuntimeChallenge{
		ID:       502,
		ImageID:  502,
		FlagType: challengecontracts.FlagTypeStatic,
	}

	if err := service.provisionInstance(context.Background(), instance, toPracticeChallenge(challenge), nil, "flag{demo}"); err != nil {
		t.Fatalf("provisionInstance() should not host-probe AWD alias URL: %v", err)
	}
	var stored instanceentity.Instance
	if err := db.First(&stored, instance.ID).Error; err != nil {
		t.Fatalf("load stored instance: %v", err)
	}
	if stored.Status != instanceentity.InstanceStatusRunning {
		t.Fatalf("expected running status, got %+v", stored)
	}
	if stored.AccessURL != "http://awd-c7002-t7102-s8002:8080" {
		t.Fatalf("expected stable alias access url, got %+v", stored)
	}
}

func TestProvisionInstanceCleansPrimaryRuntimeWhenWorkspaceStatePersistenceFails(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{
		ID:        503,
		Name:      "ctf/awd-web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(7003)
	teamID := int64(7103)
	serviceID := int64(8003)
	serviceSnapshot, err := contestentity.EncodeContestAWDServiceSnapshot(contestentity.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":         503,
			"instance_sharing": string(challengecontracts.InstanceSharingPerTeam),
			"defense_workspace": map[string]any{
				"entry_mode":      "ssh",
				"seed_root":       "docker/workspace",
				"workspace_roots": []string{"docker/workspace/src"},
				"writable_roots":  []string{"docker/workspace/src"},
				"readonly_roots":  []string{},
				"runtime_mounts": []map[string]any{
					{"source": "docker/workspace/src", "target": "/workspace/src", "mode": "rw"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("encode service snapshot: %v", err)
	}

	instance := &instanceentity.Instance{
		ID:          9003,
		UserID:      45,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 503,
		HostPort:    30031,
		Status:      instanceentity.InstanceStatusCreating,
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}

	var cleanupPayload *instanceentity.Instance
	instanceRepo := &interceptAWDDefenseWorkspaceRepository{
		practiceTestInstanceRepository: newPracticeTestInstanceRepository(db),
		upsertFn: func(ctx context.Context, workspace *runtimeentity.AWDDefenseWorkspace) error {
			if workspace != nil && workspace.Status == runtimeentity.AWDDefenseWorkspaceStatusRunning {
				return fmt.Errorf("persist running workspace state failed")
			}
			return nil
		},
	}
	service := newServiceCore(
		&stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
				return &practiceports.ContestAWDServiceRecord{
					ID:              gotServiceID,
					ContestID:       gotContestID,
					AWDChallengeID:  503,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
		},

		challengeinfra.NewImageRepository(db),
		instanceRepo,
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, got *instanceentity.Instance) error {
				copied := *got
				cleanupPayload = &copied
				return nil
			},
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				if len(req.Nodes) == 1 && req.Nodes[0].Image == "python:3.12-alpine" {
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "workspace-ctr",
						NetworkID:          "net-awd-contest-7003",
						AccessURL:          "tcp://172.30.0.41:22",
						RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
							Containers: []runtimecontracts.InstanceRuntimeContainer{
								{NodeKey: "workspace", ContainerID: "workspace-ctr", ServicePort: 22, ServiceProtocol: challengecontracts.ChallengeTargetProtocolTCP, IsEntryPoint: true},
							},
						},
					}, nil
				}
				return &practiceports.TopologyCreateResult{
					PrimaryContainerID: "runtime-ctr",
					NetworkID:          "net-awd-contest-7003",
					AccessURL:          "http://host-gateway.internal:30031",
					RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
						Networks: []runtimecontracts.InstanceRuntimeNetwork{
							{Key: runtimecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-7003", NetworkID: "net-awd-contest-7003", Shared: true},
						},
						Containers: []runtimecontracts.InstanceRuntimeContainer{
							{NodeKey: "default", ContainerID: "runtime-ctr", ServicePort: 8080, HostPort: 30031, IsEntryPoint: true, NetworkAliases: []string{"awd-c7003-t7103-s8003"}},
						},
					},
				}, nil
			},
		},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				AccessHost:         "host-gateway.internal",
				PublicHost:         "127.0.0.1",
				CreateTimeout:      time.Second,
				StartProbeTimeout:  20 * time.Millisecond,
				StartProbeInterval: 10 * time.Millisecond,
				StartProbeAttempts: 1,
			},
		},
		nil)

	err = service.provisionInstance(context.Background(), instance, toPracticeChallenge(&challengecontracts.PracticeRuntimeChallenge{
		ID:             503,
		ImageID:        503,
		FlagType:       challengecontracts.FlagTypeStatic,
		FlagHash:       "flag{demo}",
		TargetPort:     8080,
		TargetProtocol: challengecontracts.ChallengeTargetProtocolHTTP,
	}), nil, "flag{demo}")
	if err == nil || err.Error() != instancecontracts.ErrContainerCreateFailed.Error() {
		t.Fatalf("expected container create failed error, got %v", err)
	}
	if cleanupPayload == nil {
		t.Fatal("expected cleanup to be triggered")
	}
	details, err := runtimecontracts.DecodeInstanceRuntimeDetails(cleanupPayload.RuntimeDetails)
	if err != nil {
		t.Fatalf("expected cleanup payload to carry runtime details, got %+v err=%v", cleanupPayload, err)
	}
	if len(details.Containers) != 1 || details.Containers[0].ContainerID != "runtime-ctr" {
		t.Fatalf("expected cleanup payload to include primary runtime details, got %+v", details)
	}
	if cleanupPayload.ContainerID != "runtime-ctr" || cleanupPayload.NetworkID != "net-awd-contest-7003" {
		t.Fatalf("expected cleanup payload to retain primary runtime identity, got %+v", cleanupPayload)
	}
}

func TestProvisionInstanceMarksInstanceFailedWithContext(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{
		ID:        105,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	ctxKey := practiceServiceContextKey("mark-failed")
	const expectedCtxValue = "practice-provision-failure"

	var markedFailed atomic.Int32
	service := newServiceCore(
		nil,

		challengeinfra.NewImageRepository(db),
		&stubPracticeInstanceStore{
			updateStatusAndReleasePortWithContextFn: func(ctx context.Context, id int64, status string) error {
				markedFailed.Add(1)
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected failed status update ctx value %v, got %v", expectedCtxValue, got)
				}
				if id != 611 {
					t.Fatalf("expected failed instance id 611, got %d", id)
				}
				if status != instanceentity.InstanceStatusFailed {
					t.Fatalf("expected failed instance status %s, got %s", instanceentity.InstanceStatusFailed, status)
				}
				return nil
			},
		},
		&stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, _ int64) (string, string, int, int, error) {
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected create container ctx value %v, got %v", expectedCtxValue, got)
				}
				return "ctr-ctx", "net-ctx", reservedHostPort, 8080, nil
			},
		},
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
		nil).
		SetInstanceReadinessProbe(practiceinfra.NewInstanceReadinessProbe())

	instance := &instanceentity.Instance{ID: 611, ChallengeID: 711, HostPort: reserveClosedLoopbackPort(t), Status: instanceentity.InstanceStatusCreating}
	challenge := &challengecontracts.PracticeRuntimeChallenge{ID: 711, ImageID: 105, Status: challengecontracts.ChallengeStatusPublished}
	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)

	err := service.provisionInstance(ctx, instance, toPracticeChallenge(challenge), nil, "flag{ctx}")
	if err == nil || err.Error() != instancecontracts.ErrContainerStartFailed.Error() {
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
	if err := db.Create(&practiceCommandImageRow{
		ID:        103,
		Name:      "ctf/web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	for _, challengeID := range []int64{203, 204} {
		if err := db.Create(&practiceCommandChallengeRow{
			ID:         challengeID,
			Title:      "Queued Capacity",
			Category:   taxonomy.DimensionWeb,
			Difficulty: taxonomy.DifficultyEasy,
			Points:     100,
			ImageID:    103,
			Status:     challengecontracts.ChallengeStatusPublished,
			FlagType:   challengecontracts.FlagTypeStatic,
			FlagHash:   "flag{static}",
			CreatedAt:  now,
			UpdatedAt:  now,
		}).Error; err != nil {
			t.Fatalf("create challenge %d: %v", challengeID, err)
		}
	}
	for _, userID := range []int64{51, 52} {
		if err := db.Create(&identitycontracts.User{ID: userID, Username: fmt.Sprintf("student-%d", userID), Role: identitycontracts.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
			t.Fatalf("create user %d: %v", userID, err)
		}
	}

	started := make(chan int, 2)
	release := make(chan struct{})
	service := wirePracticeScopeAdapters(newServiceCore(
		newPracticeRepositoryWithRuntimePortOwner(db),

		challengeinfra.NewImageRepository(db),
		newPracticeTestInstanceRepository(db),
		&stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, _ int64) (string, string, int, int, error) {
				started <- reservedHostPort
				<-release
				return "container-capacity", "network-capacity", reservedHostPort, 8080, nil
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
					BatchSize:           2,
					MaxConcurrentStarts: 2,
					MaxActiveInstances:  1,
				},
			},
		},
		nil),

		newPracticeRepositoryWithRuntimePortOwner(db), challengeinfra.NewRepository(db))

	service.StartBackgroundTasks(context.Background())

	first, err := service.StartChallenge(context.Background(), 51, 203)
	if err != nil {
		t.Fatalf("StartChallenge() first error = %v", err)
	}
	second, err := service.StartChallenge(context.Background(), 52, 204)
	if err != nil {
		t.Fatalf("StartChallenge() second error = %v", err)
	}

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go service.RunProvisioningLoop(runCtx)

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("expected one pending instance to start provisioning")
	}

	var firstInstance instanceentity.Instance
	if err := db.First(&firstInstance, first.ID).Error; err != nil {
		t.Fatalf("load first instance: %v", err)
	}
	var secondInstance instanceentity.Instance
	if err := db.First(&secondInstance, second.ID).Error; err != nil {
		t.Fatalf("load second instance: %v", err)
	}

	statuses := []string{firstInstance.Status, secondInstance.Status}
	pendingCount := 0
	creatingCount := 0
	for _, status := range statuses {
		if status == instanceentity.InstanceStatusPending {
			pendingCount++
		}
		if status == instanceentity.InstanceStatusCreating {
			creatingCount++
		}
	}
	if pendingCount != 1 || creatingCount != 1 {
		t.Fatalf("expected one creating and one pending instance, got %+v", statuses)
	}

	close(release)
}
