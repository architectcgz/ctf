package commands

import (
	"context"
	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	"sync/atomic"
	"testing"
	"time"
)

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

	resp, err := service.StartChallenge(context.Background(), 42, 201)
	if err != nil {
		t.Fatalf("StartChallenge() error = %v", err)
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
		findContestByIDFn: func(ctx context.Context, contestID int64) (*model.Contest, error) {
			if contestID != 3104 {
				t.Fatalf("unexpected contest id: %d", contestID)
			}
			return &model.Contest{
				ID:     contestID,
				Mode:   model.ContestModeAWD,
				Status: model.ContestStatusRunning,
			}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error) {
			if contestID != 3104 || serviceID != 7104 {
				t.Fatalf("unexpected awd service lookup: contest=%d service=%d", contestID, serviceID)
			}
			return &model.ContestAWDService{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2104,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":104,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestChallengeFn: func(ctx context.Context, contestID, challengeID int64) (*model.ContestChallenge, error) {
			t.Fatalf("unexpected contest challenge lookup for awd start: contest=%d challenge=%d", contestID, challengeID)
			return nil, nil
		},
		findContestRegistrationFn: func(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
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
		createInstanceFn: func(ctx context.Context, instance *model.Instance) error {
			instance.ID = 9104
			return nil
		},
	}

	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
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

func TestStartContestAWDServiceDoesNotReserveHostPort(t *testing.T) {
	t.Parallel()

	teamID := int64(4105)
	var createdInstance *model.Instance
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, contestID int64) (*model.Contest, error) {
			return &model.Contest{
				ID:     contestID,
				Mode:   model.ContestModeAWD,
				Status: model.ContestStatusRunning,
			}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, contestID, serviceID int64) (*model.ContestAWDService, error) {
			return &model.ContestAWDService{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2105,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":105,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, contestID, userID int64) (*model.ContestRegistration, error) {
			return &model.ContestRegistration{
				ContestID: contestID,
				UserID:    userID,
				TeamID:    &teamID,
				Status:    model.ContestRegistrationStatusApproved,
			}, nil
		},
		reserveAvailablePortFn: func(ctx context.Context, start, end int) (int, error) {
			t.Fatal("AWD service instances must not reserve a host port")
			return 0, nil
		},
		bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
			t.Fatalf("AWD service instances must not bind a reserved host port: port=%d instance_id=%d", port, instanceID)
			return nil
		},
		createInstanceFn: func(ctx context.Context, instance *model.Instance) error {
			instance.ID = 9105
			copied := *instance
			createdInstance = &copied
			return nil
		},
	}

	service := NewService(
		repo,
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{
					ID:       id,
					Status:   model.ChallengeStatusPublished,
					ImageID:  105,
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
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil,
	)

	resp, err := service.StartContestAWDService(context.Background(), 5105, 3105, 7105)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}
	if resp.ID != 9105 {
		t.Fatalf("expected created awd service instance, got %+v", resp)
	}
	if createdInstance == nil {
		t.Fatal("expected instance to be created")
	}
	if createdInstance.HostPort != 0 {
		t.Fatalf("expected no AWD host port reservation, got %d", createdInstance.HostPort)
	}
}

func TestRestartContestAWDServiceRequeuesExistingTeamInstance(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4106)
	serviceID := int64(7106)
	contestID := int64(3106)
	userID := int64(5106)
	instance := &model.Instance{
		ID:             9106,
		UserID:         userID,
		ContestID:      &contestID,
		TeamID:         &teamID,
		ChallengeID:    2106,
		ServiceID:      &serviceID,
		HostPort:       32106,
		ContainerID:    "old-container",
		NetworkID:      "old-network",
		RuntimeDetails: `{"containers":[{"id":"old-container"}]}`,
		ShareScope:     model.InstanceSharingPerTeam,
		Status:         model.InstanceStatusRunning,
		AccessURL:      "http://127.0.0.1:32106",
		Nonce:          "nonce-keep",
		ExpiresAt:      now.Add(-time.Minute),
		MaxExtends:     2,
	}
	var cleanupInstanceID int64
	var resetStatus string
	var operation *model.AWDServiceOperation
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*model.Contest, error) {
			return &model.Contest{ID: gotContestID, Mode: model.ContestModeAWD, Status: model.ContestStatusRunning}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*model.ContestAWDService, error) {
			if gotContestID != contestID || gotServiceID != serviceID {
				t.Fatalf("unexpected awd service lookup: contest=%d service=%d", gotContestID, gotServiceID)
			}
			return &model.ContestAWDService{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2106,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":106,"instance_sharing":"per_team"},"flag_config":{"flag_type":"dynamic","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*model.ContestRegistration, error) {
			return &model.ContestRegistration{
				ContestID: gotContestID,
				UserID:    gotUserID,
				TeamID:    &teamID,
				Status:    model.ContestRegistrationStatusApproved,
			}, nil
		},
		findScopedRestartableInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*model.Instance, error) {
			if gotUserID != userID || gotChallengeID != 2106 {
				t.Fatalf("unexpected scoped lookup: user=%d challenge=%d", gotUserID, gotChallengeID)
			}
			if scope.ServiceID == nil || *scope.ServiceID != serviceID || scope.TeamID == nil || *scope.TeamID != teamID || scope.ShareScope != model.InstanceSharingPerTeam {
				t.Fatalf("unexpected restart scope: %+v", scope)
			}
			return instance, nil
		},
		resetInstanceRuntimeForRestartFn: func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
			if instanceID != instance.ID {
				t.Fatalf("unexpected reset instance id: %d", instanceID)
			}
			if !expiresAt.After(now) {
				t.Fatalf("expected refreshed restart expiry after now, got %s now=%s", expiresAt, now)
			}
			if preserveHostPort {
				t.Fatal("AWD restart must clear historical host port instead of preserving it")
			}
			resetStatus = status
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, got *model.AWDServiceOperation) error {
			operation = got
			return nil
		},
	}

	service := NewService(
		repo,
		&stubPracticeChallengeContract{},
		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, got *model.Instance) error {
				cleanupInstanceID = got.ID
				if got.HostPort != 0 {
					t.Fatalf("restart cleanup should preserve port allocation, got host_port=%d", got.HostPort)
				}
				return nil
			},
		},
		nil,
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil,
	)

	resp, err := service.RestartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("RestartContestAWDService() error = %v", err)
	}
	if resp.ID != instance.ID || resp.Status != model.InstanceStatusPending {
		t.Fatalf("expected same pending instance, got %+v", resp)
	}
	if cleanupInstanceID != instance.ID {
		t.Fatalf("expected cleanup for instance %d, got %d", instance.ID, cleanupInstanceID)
	}
	if resetStatus != model.InstanceStatusPending {
		t.Fatalf("expected reset to pending, got %q", resetStatus)
	}
	if !resp.ExpiresAt.After(now) || !instance.ExpiresAt.After(now) {
		t.Fatalf("restart should refresh expired instance ttl, resp=%s instance=%s now=%s", resp.ExpiresAt, instance.ExpiresAt, now)
	}
	if instance.ServiceID == nil || *instance.ServiceID != serviceID || instance.Nonce != "nonce-keep" || instance.HostPort != 0 {
		t.Fatalf("restart should preserve identity fields, got %+v", instance)
	}
	if instance.ContainerID != "" || instance.NetworkID != "" || instance.RuntimeDetails != "" || instance.AccessURL != "" {
		t.Fatalf("restart should clear runtime fields, got %+v", instance)
	}
	if operation == nil || operation.OperationType != model.AWDServiceOperationTypeRestart || operation.RequestedBy != model.AWDServiceOperationRequestedByUser || !operation.SLABillable {
		t.Fatalf("expected billable user restart operation, got %+v", operation)
	}
}

func TestRestartContestAWDServicePreservesExistingDefenseWorkspaceRevision(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	contestID := int64(9101)
	teamID := int64(9102)
	serviceID := int64(9103)
	userID := int64(9104)
	imageID := int64(9105)
	challengeID := int64(9106)

	if err := db.Create(&model.Image{
		ID:        imageID,
		Name:      "ctf/awd-runtime",
		Tag:       "v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "AWD Restart",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.User{ID: userID, Username: "restart-student", Role: model.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&model.ContestRegistration{
		ContestID: contestID,
		UserID:    userID,
		TeamID:    &teamID,
		Status:    model.ContestRegistrationStatusApproved,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}
	serviceSnapshot, err := model.EncodeContestAWDServiceSnapshot(model.ContestAWDServiceSnapshot{
		Name: "Restart Service",
		RuntimeConfig: map[string]any{
			"image_id":         imageID,
			"instance_sharing": string(model.InstanceSharingPerTeam),
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
	if err := db.Create(&model.ContestAWDService{
		ID:              serviceID,
		ContestID:       contestID,
		AWDChallengeID:  challengeID,
		DisplayName:     "Restart Service",
		ServiceSnapshot: serviceSnapshot,
		IsVisible:       true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create awd service: %v", err)
	}

	instance := &model.Instance{
		ID:          9201,
		UserID:      userID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
		Status:      model.InstanceStatusRunning,
		ShareScope:  model.InstanceSharingPerTeam,
		ContainerID: "runtime-old",
		NetworkID:   "net-old",
		AccessURL:   "http://awd-c9101-t9102-s9103:8080",
		Nonce:       "nonce",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}
	if err := db.Create(&model.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        instance.ID,
		WorkspaceRevision: 4,
		Status:            model.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-existing",
		SeedSignature:     "seed:v1",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create workspace row: %v", err)
	}

	var createTopologyCalls atomic.Int32
	service := NewService(
		practiceinfra.NewRepository(db),
		challengeinfra.NewRepository(db),
		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(context.Context, *model.Instance) error { return nil },
			inspectManagedContainerFn: func(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error) {
				if containerID != "workspace-existing" {
					t.Fatalf("unexpected workspace inspect: %s", containerID)
				}
				return &practiceports.ManagedContainerState{
					ID:      containerID,
					Exists:  true,
					Running: true,
					Status:  "running",
				}, nil
			},
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				createTopologyCalls.Add(1)
				return &practiceports.TopologyCreateResult{
					PrimaryContainerID: "runtime-new",
					NetworkID:          "net-awd-contest-9101",
					AccessURL:          "http://awd-c9101-t9102-s9103:8080",
					RuntimeDetails: model.InstanceRuntimeDetails{
						Networks: []model.InstanceRuntimeNetwork{
							{Key: model.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-9101", NetworkID: "net-awd-contest-9101", Shared: true},
						},
						Containers: []model.InstanceRuntimeContainer{
							{NodeKey: "default", ContainerID: "runtime-new", ServicePort: 8080, IsEntryPoint: true, NetworkAliases: []string{"awd-c9101-t9102-s9103"}},
						},
					},
				}, nil
			},
		},
		nil,
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				FlagGlobalSecret:     "restart-secret",
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				CreateTimeout:        time.Second,
				Scheduler:            config.ContainerSchedulerConfig{Enabled: false},
			},
		},
		nil,
	)

	resp, err := service.RestartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("RestartContestAWDService() error = %v", err)
	}
	if resp.Status != model.InstanceStatusRunning {
		t.Fatalf("expected restarted instance to be running, got %+v", resp)
	}
	if createTopologyCalls.Load() != 1 {
		t.Fatalf("expected only runtime container recreation, got %d topology calls", createTopologyCalls.Load())
	}

	workspace, err := runtimeinfrarepo.NewRepository(db).FindAWDDefenseWorkspace(context.Background(), contestID, teamID, serviceID)
	if err != nil {
		t.Fatalf("FindAWDDefenseWorkspace() error = %v", err)
	}
	if workspace == nil {
		t.Fatal("expected workspace row to remain")
	}
	if workspace.WorkspaceRevision != 4 {
		t.Fatalf("expected workspace revision 4 to be preserved, got %+v", workspace)
	}
	if workspace.ContainerID != "workspace-existing" {
		t.Fatalf("expected workspace container to be reused, got %+v", workspace)
	}
}

func TestRestartContestAWDServiceRecreatesMissingDefenseWorkspaceContainer(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	contestID := int64(9111)
	teamID := int64(9112)
	serviceID := int64(9113)
	userID := int64(9114)
	imageID := int64(9115)
	challengeID := int64(9116)

	if err := db.Create(&model.Image{
		ID:        imageID,
		Name:      "ctf/awd-runtime",
		Tag:       "v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&model.Challenge{
		ID:        challengeID,
		Title:     "Restart Service",
		ImageID:   imageID,
		Status:    model.ChallengeStatusPublished,
		FlagType:  model.FlagTypeStatic,
		FlagHash:  "flag{restart}",
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&model.Contest{
		ID:        contestID,
		Title:     "AWD Restart Missing Workspace",
		Mode:      model.ContestModeAWD,
		Status:    model.ContestStatusRunning,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&model.User{ID: userID, Username: "restart-student-missing", Role: model.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&model.ContestRegistration{
		ContestID: contestID,
		UserID:    userID,
		TeamID:    &teamID,
		Status:    model.ContestRegistrationStatusApproved,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}
	serviceSnapshot, err := model.EncodeContestAWDServiceSnapshot(model.ContestAWDServiceSnapshot{
		Name: "Restart Service Missing Workspace",
		RuntimeConfig: map[string]any{
			"image_id":         imageID,
			"instance_sharing": string(model.InstanceSharingPerTeam),
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
	if err := db.Create(&model.ContestAWDService{
		ID:              serviceID,
		ContestID:       contestID,
		AWDChallengeID:  challengeID,
		DisplayName:     "Restart Service Missing Workspace",
		ServiceSnapshot: serviceSnapshot,
		IsVisible:       true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create awd service: %v", err)
	}

	instance := &model.Instance{
		ID:          9211,
		UserID:      userID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
		Status:      model.InstanceStatusRunning,
		ShareScope:  model.InstanceSharingPerTeam,
		ContainerID: "runtime-old",
		NetworkID:   "net-old",
		AccessURL:   "http://awd-c9111-t9112-s9113:8080",
		Nonce:       "nonce",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("create instance: %v", err)
	}
	if err := db.Create(&model.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        instance.ID,
		WorkspaceRevision: 4,
		Status:            model.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-missing",
		SeedSignature:     "seed:v1",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create workspace row: %v", err)
	}

	var createTopologyCalls atomic.Int32
	service := NewService(
		practiceinfra.NewRepository(db),
		challengeinfra.NewRepository(db),
		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(context.Context, *model.Instance) error { return nil },
			inspectManagedContainerFn: func(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error) {
				if containerID != "workspace-missing" {
					t.Fatalf("unexpected workspace inspect: %s", containerID)
				}
				return &practiceports.ManagedContainerState{
					ID:      containerID,
					Exists:  false,
					Running: false,
					Status:  "missing",
				}, nil
			},
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				switch createTopologyCalls.Add(1) {
				case 1:
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "runtime-new",
						NetworkID:          "net-awd-contest-9111",
						AccessURL:          "http://awd-c9111-t9112-s9113:8080",
						RuntimeDetails: model.InstanceRuntimeDetails{
							Networks: []model.InstanceRuntimeNetwork{
								{Key: model.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-9111", NetworkID: "net-awd-contest-9111", Shared: true},
							},
							Containers: []model.InstanceRuntimeContainer{
								{NodeKey: "default", ContainerID: "runtime-new", ServicePort: 8080, IsEntryPoint: true, NetworkAliases: []string{"awd-c9111-t9112-s9113"}},
							},
						},
					}, nil
				case 2:
					if len(req.Nodes) != 1 {
						t.Fatalf("expected one workspace node, got %+v", req.Nodes)
					}
					assertAWDDefenseWorkspaceShellNode(t, req.Nodes[0])
					if req.Nodes[0].WorkingDir != "/workspace" {
						t.Fatalf("expected workspace working dir, got %+v", req.Nodes[0])
					}
					if len(req.Nodes[0].Mounts) != 1 || req.Nodes[0].Mounts[0].Target != "/workspace/src" {
						t.Fatalf("unexpected workspace mounts: %+v", req.Nodes[0].Mounts)
					}
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "workspace-recreated",
						NetworkID:          "net-awd-contest-9111",
						AccessURL:          "tcp://172.30.0.55:22",
						RuntimeDetails: model.InstanceRuntimeDetails{
							Containers: []model.InstanceRuntimeContainer{
								{NodeKey: "workspace", ContainerID: "workspace-recreated", ServicePort: 22, ServiceProtocol: model.ChallengeTargetProtocolTCP, IsEntryPoint: true},
							},
						},
					}, nil
				default:
					t.Fatalf("unexpected topology create call #%d", createTopologyCalls.Load())
					return nil, nil
				}
			},
		},
		nil,
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				FlagGlobalSecret:     "restart-secret",
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				CreateTimeout:        time.Second,
				Scheduler:            config.ContainerSchedulerConfig{Enabled: false},
			},
		},
		nil,
	)

	resp, err := service.RestartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("RestartContestAWDService() error = %v", err)
	}
	if resp.Status != model.InstanceStatusRunning {
		t.Fatalf("expected restarted instance to be running, got %+v", resp)
	}
	if createTopologyCalls.Load() != 2 {
		t.Fatalf("expected runtime and workspace recreation, got %d topology calls", createTopologyCalls.Load())
	}

	workspace, err := runtimeinfrarepo.NewRepository(db).FindAWDDefenseWorkspace(context.Background(), contestID, teamID, serviceID)
	if err != nil {
		t.Fatalf("FindAWDDefenseWorkspace() error = %v", err)
	}
	if workspace == nil {
		t.Fatal("expected workspace row to remain")
	}
	if workspace.WorkspaceRevision != 4 {
		t.Fatalf("expected workspace revision 4 to be preserved, got %+v", workspace)
	}
	if workspace.ContainerID != "workspace-recreated" {
		t.Fatalf("expected workspace container to be recreated, got %+v", workspace)
	}
}

func TestStartChallengeIgnoresExpiredRunningInstance(t *testing.T) {
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

	resp, err := service.StartChallenge(context.Background(), 46, 206)
	if err != nil {
		t.Fatalf("StartChallenge() error = %v", err)
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

func TestStartChallengePropagatesContextToTransactionalRepositoryWhenReusingSharedInstance(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("tx-reuse")
	expectedCtxValue := "ctx-tx-reuse"
	lockCalled := false
	findExistingCalled := false
	refreshCalled := false
	service := NewService(
		&stubPracticeRepository{
			lockInstanceScopeFn: func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) error {
				lockCalled = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected lock ctx value %v, got %v", expectedCtxValue, got)
				}
				return nil
			},
			findScopedExistingInstanceFn: func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error) {
				findExistingCalled = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected find-existing ctx value %v, got %v", expectedCtxValue, got)
				}
				return &model.Instance{ID: 901, UserID: 7, ChallengeID: challengeID, ShareScope: model.InstanceSharingShared, Status: model.InstanceStatusRunning, ExpiresAt: time.Now().Add(5 * time.Minute), MaxExtends: 2}, nil
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
		},
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{ID: id, ImageID: 1, Status: model.ChallengeStatusPublished, FlagType: model.FlagTypeStatic, FlagHash: "flag{shared}", InstanceSharing: model.InstanceSharingShared}, nil
			},
			findChallengeTopologyByChallengeIDFn: func(context.Context, int64) (*model.ChallengeTopology, error) {
				return nil, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{Container: config.ContainerConfig{DefaultTTL: time.Hour, MaxConcurrentPerUser: 3}},
		nil,
	)

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

func TestStartChallengePropagatesContextToTransactionalRepositoryWhenCreatingInstance(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("tx-create")
	expectedCtxValue := "ctx-tx-create"
	countCalled := false
	reserveCalled := false
	createCalled := false
	bindCalled := false
	service := NewService(
		&stubPracticeRepository{
			lockInstanceScopeFn: func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) error {
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected lock ctx value %v, got %v", expectedCtxValue, got)
				}
				return nil
			},
			findScopedExistingInstanceFn: func(ctx context.Context, userID, challengeID int64, scope practiceports.InstanceScope) (*model.Instance, error) {
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
			createInstanceFn: func(ctx context.Context, instance *model.Instance) error {
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
		},
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				return &model.Challenge{ID: id, ImageID: 1, Status: model.ChallengeStatusPublished, FlagType: model.FlagTypeStatic, FlagHash: "flag{new}"}, nil
			},
			findChallengeTopologyByChallengeIDFn: func(context.Context, int64) (*model.ChallengeTopology, error) {
				return nil, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{Container: config.ContainerConfig{DefaultTTL: time.Hour, MaxConcurrentPerUser: 3, MaxExtends: 2, Scheduler: config.ContainerSchedulerConfig{Enabled: true}}},
		nil,
	)

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
