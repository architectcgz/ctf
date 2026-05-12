package commands

import (
	"context"
	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	"ctf-platform/pkg/errcode"
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
		nil)

	service.StartBackgroundTasks(context.Background())

	resp, err := service.StartChallenge(context.Background(), 43, 202)
	if err != nil {
		t.Fatalf("StartChallenge() error = %v", err)
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
			cleanupRuntimeFn: func(context.Context, *model.Instance) error {
				cleanupCalls.Add(1)
				return nil
			},
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error) {
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
		nil)

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

func TestProvisionInstancePropagatesContextToUpdateRuntime(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("update-runtime")
	expectedCtxValue := "ctx-update-runtime"
	instanceStore := &stubPracticeInstanceStore{
		updateRuntimeWithContextFn: func(ctx context.Context, instance *model.Instance) error {
			if got := ctx.Value(ctxKey); got != expectedCtxValue {
				t.Fatalf("expected update runtime ctx value %v, got %v", expectedCtxValue, got)
			}
			if instance.Status != model.InstanceStatusRunning {
				t.Fatalf("expected running status before persistence, got %+v", instance)
			}
			return nil
		},
	}
	service := NewService(
		nil,
		nil,
		&stubPracticeImageStore{
			findByIDFn: func(ctx context.Context, id int64) (*model.Image, error) {
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected image lookup ctx value %v, got %v", expectedCtxValue, got)
				}
				return &model.Image{ID: id, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable}, nil
			},
		},
		instanceStore,
		&stubPracticeRuntimeService{
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error) {
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected runtime create ctx value %v, got %v", expectedCtxValue, got)
				}
				return "ctr-running", "net-running", reservedHostPort, 8080, nil
			},
		},
		nil,
		nil,
		&config.Config{Container: config.ContainerConfig{PublicHost: "127.0.0.1", CreateTimeout: time.Second, StartProbeTimeout: 50 * time.Millisecond, StartProbeInterval: 10 * time.Millisecond, StartProbeAttempts: 1}},
		nil)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	host, port := parseHTTPServerEndpoint(t, server.URL)
	instance := &model.Instance{ID: 951, ChallengeID: 2051, HostPort: port, Status: model.InstanceStatusCreating}
	challenge := &model.Challenge{ID: 2051, ImageID: 301, Status: model.ChallengeStatusPublished, FlagType: model.FlagTypeStatic, FlagHash: "flag{ok}"}
	service.config.Container.PublicHost = host
	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)

	if err := service.provisionInstance(ctx, instance, challenge, nil, "flag{ok}"); err != nil {
		t.Fatalf("provisionInstance() error = %v", err)
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
		updateRuntimeWithContextFn: func(ctx context.Context, instance *model.Instance) error {
			if instance.Status != model.InstanceStatusRunning {
				t.Fatalf("expected running status, got %+v", instance)
			}
			if !strings.HasPrefix(instance.AccessURL, "tcp://") {
				t.Fatalf("expected tcp access url, got %q", instance.AccessURL)
			}
			return nil
		},
	}
	service := NewService(
		nil,
		nil,
		&stubPracticeImageStore{
			findByIDFn: func(context.Context, int64) (*model.Image, error) {
				return &model.Image{ID: 301, Name: "ctf/pwn", Tag: "v1", Status: model.ImageStatusAvailable}, nil
			},
		},
		instanceStore,
		&stubPracticeRuntimeService{
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				if len(req.Nodes) != 1 {
					t.Fatalf("unexpected topology request: %+v", req)
				}
				if req.Nodes[0].ServiceProtocol != model.ChallengeTargetProtocolTCP {
					t.Fatalf("expected tcp topology node, got %+v", req.Nodes[0])
				}
				return &practiceports.TopologyCreateResult{
					PrimaryContainerID: "pwn-ctr",
					NetworkID:          "pwn-net",
					AccessURL:          fmt.Sprintf("tcp://%s", listener.Addr().String()),
					RuntimeDetails: model.InstanceRuntimeDetails{
						Containers: []model.InstanceRuntimeContainer{
							{
								NodeKey:         "default",
								ContainerID:     "pwn-ctr",
								ServicePort:     8080,
								ServiceProtocol: model.ChallengeTargetProtocolTCP,
								IsEntryPoint:    true,
								NetworkKeys:     []string{model.TopologyDefaultNetworkKey},
							},
						},
					},
				}, nil
			},
		},
		nil,
		nil,
		&config.Config{Container: config.ContainerConfig{PublicHost: "127.0.0.1", CreateTimeout: time.Second, StartProbeTimeout: 50 * time.Millisecond, StartProbeInterval: 10 * time.Millisecond, StartProbeAttempts: 2}},
		nil)

	instance := &model.Instance{ID: 952, ChallengeID: 2052, HostPort: 0, Status: model.InstanceStatusCreating}
	challenge := &model.Challenge{
		ID:             2052,
		ImageID:        301,
		Status:         model.ChallengeStatusPublished,
		FlagType:       model.FlagTypeStatic,
		FlagHash:       "flag{ok}",
		TargetProtocol: model.ChallengeTargetProtocolTCP,
	}

	if err := service.provisionInstance(context.Background(), instance, challenge, nil, "flag{ok}"); err != nil {
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
	if err := db.Create(&model.Image{
		ID:        502,
		Name:      "ctf/awd-web",
		Tag:       "v2",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(7002)
	teamID := int64(7102)
	serviceID := int64(8002)
	serviceSnapshot, err := model.EncodeContestAWDServiceSnapshot(model.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":         502,
			"instance_sharing": string(model.InstanceSharingPerTeam),
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
			"flag_type":   model.FlagTypeStatic,
			"flag_prefix": "flag",
		},
	})
	if err != nil {
		t.Fatalf("encode service snapshot: %v", err)
	}
	service := &Service{
		repo: &stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*model.ContestAWDService, error) {
				if gotContestID != contestID || gotServiceID != serviceID {
					t.Fatalf("unexpected awd service lookup: contest=%d service=%d", gotContestID, gotServiceID)
				}
				return &model.ContestAWDService{
					ID:              serviceID,
					ContestID:       contestID,
					AWDChallengeID:  502,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
		},
		imageRepo:    challengeinfra.NewImageRepository(db),
		instanceRepo: runtimeinfrarepo.NewRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				return &practiceports.TopologyCreateResult{
					PrimaryContainerID: "awd-alias-ctr",
					NetworkID:          "net-awd-contest-7002",
					AccessURL:          "http://awd-c7002-t7102-s8002:8080",
					RuntimeDetails: model.InstanceRuntimeDetails{
						Networks: []model.InstanceRuntimeNetwork{
							{Key: model.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-7002", NetworkID: "net-awd-contest-7002", Shared: true},
						},
						Containers: []model.InstanceRuntimeContainer{
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
	instance := &model.Instance{
		ID:          9002,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 502,
		Status:      model.InstanceStatusCreating,
	}
	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}
	challenge := &model.Challenge{
		ID:       502,
		ImageID:  502,
		FlagType: model.FlagTypeStatic,
	}

	if err := service.provisionInstance(context.Background(), instance, challenge, nil, "flag{demo}"); err != nil {
		t.Fatalf("provisionInstance() should not host-probe AWD alias URL: %v", err)
	}
	var stored model.Instance
	if err := db.First(&stored, instance.ID).Error; err != nil {
		t.Fatalf("load stored instance: %v", err)
	}
	if stored.Status != model.InstanceStatusRunning {
		t.Fatalf("expected running status, got %+v", stored)
	}
	if stored.AccessURL != "http://awd-c7002-t7102-s8002:8080" {
		t.Fatalf("expected stable alias access url, got %+v", stored)
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
		&config.Config{
			Container: config.ContainerConfig{
				PublicHost:         "127.0.0.1",
				CreateTimeout:      time.Second,
				StartProbeTimeout:  20 * time.Millisecond,
				StartProbeInterval: 10 * time.Millisecond,
				StartProbeAttempts: 1,
			},
		},
		nil)

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
		nil)

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
