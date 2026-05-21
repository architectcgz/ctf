package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
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
		practiceinfra.NewRepository(db),

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

		practiceinfra.NewRepository(db), challengeinfra.NewRepository(db))

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

func TestStartContestAWDServiceDoesNotRequireContestChallengeLookup(t *testing.T) {
	t.Parallel()

	teamID := int64(4104)
	contestEnd := time.Date(2026, 5, 15, 12, 0, 0, 0, time.UTC)
	var createdInstance *instanceentity.Instance
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, contestID int64) (*practiceports.ContestRecord, error) {
			if contestID != 3104 {
				t.Fatalf("unexpected contest id: %d", contestID)
			}
			return &practiceports.ContestRecord{
				ID:      contestID,
				Mode:    practiceports.ContestModeAWD,
				Status:  practiceports.ContestStatusRunning,
				EndTime: contestEnd,
			}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, contestID, serviceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			if contestID != 3104 || serviceID != 7104 {
				t.Fatalf("unexpected awd service lookup: contest=%d service=%d", contestID, serviceID)
			}
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2104,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":104,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestChallengeFn: func(ctx context.Context, contestID, challengeID int64) (*practiceports.ContestChallengeRecord, error) {
			t.Fatalf("unexpected contest challenge lookup for awd start: contest=%d challenge=%d", contestID, challengeID)
			return nil, nil
		},
		findContestRegistrationFn: func(ctx context.Context, contestID, userID int64) (*practiceports.ContestParticipation, error) {
			if contestID != 3104 || userID != 5104 {
				t.Fatalf("unexpected registration lookup: contest=%d user=%d", contestID, userID)
			}
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		createInstanceFn: func(ctx context.Context, instance *instanceentity.Instance) error {
			copied := *instance
			createdInstance = &copied
			instance.ID = 9104
			return nil
		},
	}

	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			if id != 2104 {
				t.Fatalf("unexpected challenge lookup: %d", id)
			}
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:       id,
				Status:   challengecontracts.ChallengeStatusPublished,
				ImageID:  104,
				FlagType: challengecontracts.FlagTypeStatic,
				FlagHash: "flag{awd-static}",
			}, nil
		},
	}
	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
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
		nil),

		repo, challengeRepo)

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
	if resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected pending awd service instance, got %+v", resp)
	}
	if createdInstance == nil || !createdInstance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("expected awd instance expiry to follow contest end time %s, got %+v", contestEnd, createdInstance)
	}
}

func TestStartContestAWDServiceDoesNotReserveHostPort(t *testing.T) {
	t.Parallel()

	teamID := int64(4105)
	contestEnd := time.Date(2026, 5, 15, 13, 0, 0, 0, time.UTC)
	var createdInstance *instanceentity.Instance
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, contestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{
				ID:      contestID,
				Mode:    practiceports.ContestModeAWD,
				Status:  practiceports.ContestStatusRunning,
				EndTime: contestEnd,
			}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, contestID, serviceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2105,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":105,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, contestID, userID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		reserveAvailablePortFn: func(ctx context.Context, start, end int) (int, error) {
			t.Fatal("AWD service instances must not reserve a host port")
			return 0, nil
		},
		bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
			t.Fatalf("AWD service instances must not bind a reserved host port: port=%d instance_id=%d", port, instanceID)
			return nil
		},
		createInstanceFn: func(ctx context.Context, instance *instanceentity.Instance) error {
			instance.ID = 9105
			copied := *instance
			createdInstance = &copied
			return nil
		},
	}

	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:       id,
				Status:   challengecontracts.ChallengeStatusPublished,
				ImageID:  105,
				FlagType: challengecontracts.FlagTypeStatic,
				FlagHash: "flag{awd-static}",
			}, nil
		},
	}
	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
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
		nil),

		repo, challengeRepo)

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
	if !createdInstance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("expected awd instance expiry to follow contest end time %s, got %s", contestEnd, createdInstance.ExpiresAt)
	}
}

func TestStartContestAWDServiceReservesHostPortWhenAccessHostConfigured(t *testing.T) {
	t.Parallel()

	teamID := int64(4115)
	contestEnd := time.Date(2026, 5, 15, 14, 0, 0, 0, time.UTC)
	var createdInstance *instanceentity.Instance
	reserved := false
	bound := false
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, contestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{
				ID:      contestID,
				Mode:    practiceports.ContestModeAWD,
				Status:  practiceports.ContestStatusRunning,
				EndTime: contestEnd,
			}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, contestID, serviceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2115,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":115,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, contestID, userID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		reserveAvailablePortFn: func(ctx context.Context, start, end int) (int, error) {
			reserved = true
			return 30015, nil
		},
		bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
			bound = true
			if port != 30015 || instanceID != 9115 {
				t.Fatalf("unexpected host port bind: port=%d instance_id=%d", port, instanceID)
			}
			return nil
		},
		createInstanceFn: func(ctx context.Context, instance *instanceentity.Instance) error {
			instance.ID = 9115
			copied := *instance
			createdInstance = &copied
			return nil
		},
	}

	challengeRepo := &stubPracticeChallengeContract{
		findByIDWithContextFn: func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
			return &challengecontracts.PracticeRuntimeChallenge{
				ID:       id,
				Status:   challengecontracts.ChallengeStatusPublished,
				ImageID:  115,
				FlagType: challengecontracts.FlagTypeStatic,
				FlagHash: "flag{awd-static}",
			}, nil
		},
	}
	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				AccessHost:           "host-gateway.internal",
				PortRangeStart:       30000,
				PortRangeEnd:         30020,
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil),

		repo, challengeRepo)

	resp, err := service.StartContestAWDService(context.Background(), 5115, 3115, 7115)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}
	if resp.ID != 9115 {
		t.Fatalf("expected created awd service instance, got %+v", resp)
	}
	if resp.AccessURL != "" {
		t.Fatalf("expected student-facing awd access url to stay hidden, got %+v", resp)
	}
	if !reserved || !bound {
		t.Fatalf("expected awd instance to reserve and bind host port, reserved=%v bound=%v", reserved, bound)
	}
	if createdInstance == nil {
		t.Fatal("expected instance to be created")
	}
	if createdInstance.HostPort != 30015 {
		t.Fatalf("expected reserved AWD host port, got %d", createdInstance.HostPort)
	}
	if !createdInstance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("expected awd instance expiry to follow contest end time %s, got %s", contestEnd, createdInstance.ExpiresAt)
	}
}

func TestStartContestAWDServiceRefreshesExistingInstanceExpiryToContestEnd(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4118)
	serviceID := int64(7118)
	contestID := int64(3118)
	userID := int64(5118)
	contestEnd := now.Add(6 * time.Hour).UTC()
	existingInstance := &instanceentity.Instance{
		ID:          9118,
		UserID:      userID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 2118,
		ServiceID:   &serviceID,
		ShareScope:  instanceentity.ShareScopePerTeam,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(5 * time.Minute),
		MaxExtends:  2,
	}
	refreshCalled := false
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{ID: gotContestID, Mode: practiceports.ContestModeAWD, Status: practiceports.ContestStatusRunning, EndTime: contestEnd}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2118,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":118,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		findScopedExistingInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			if gotUserID != userID || gotChallengeID != 2118 {
				t.Fatalf("unexpected scoped lookup: user=%d challenge=%d", gotUserID, gotChallengeID)
			}
			return existingInstance, nil
		},
		refreshInstanceExpiryWithContextFn: func(ctx context.Context, instanceID int64, expiresAt time.Time) error {
			refreshCalled = true
			if instanceID != existingInstance.ID {
				t.Fatalf("unexpected refresh instance id: %d", instanceID)
			}
			if !expiresAt.Equal(contestEnd) {
				t.Fatalf("expected expiry refresh to contest end time %s, got %s", contestEnd, expiresAt)
			}
			return nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
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
		nil),

		repo, &stubPracticeChallengeContract{})

	resp, err := service.StartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}
	if resp.ID != existingInstance.ID {
		t.Fatalf("expected reused awd instance, got %+v", resp)
	}
	if !refreshCalled {
		t.Fatal("expected awd instance expiry to be refreshed")
	}
	if !resp.ExpiresAt.Equal(contestEnd) || !existingInstance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("expected awd instance expiry to refresh to contest end time %s, resp=%s stored=%s", contestEnd, resp.ExpiresAt, existingInstance.ExpiresAt)
	}
}

func TestStartContestAWDServiceDoesNotRefreshStoppingInstanceExpiry(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4218)
	serviceID := int64(7218)
	contestID := int64(3218)
	userID := int64(5218)
	contestEnd := now.Add(6 * time.Hour).UTC()
	existingExpiry := now.Add(5 * time.Minute)
	existingInstance := &instanceentity.Instance{
		ID:          9218,
		UserID:      userID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 2218,
		ServiceID:   &serviceID,
		ShareScope:  instanceentity.ShareScopePerTeam,
		Status:      instanceentity.InstanceStatusStopping,
		ExpiresAt:   existingExpiry,
		MaxExtends:  2,
	}
	refreshCalled := false
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{ID: gotContestID, Mode: practiceports.ContestModeAWD, Status: practiceports.ContestStatusRunning, EndTime: contestEnd}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2218,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":118,"instance_sharing":"per_team"},"flag_config":{"flag_type":"static","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		findScopedExistingInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			if gotUserID != userID || gotChallengeID != 2218 {
				t.Fatalf("unexpected scoped lookup: user=%d challenge=%d", gotUserID, gotChallengeID)
			}
			return existingInstance, nil
		},
		refreshInstanceExpiryWithContextFn: func(ctx context.Context, instanceID int64, expiresAt time.Time) error {
			refreshCalled = true
			return nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{},
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
		nil),

		repo, &stubPracticeChallengeContract{})

	resp, err := service.StartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("StartContestAWDService() error = %v", err)
	}
	if resp.ID != existingInstance.ID {
		t.Fatalf("expected reused awd stopping instance, got %+v", resp)
	}
	if refreshCalled {
		t.Fatal("stopping awd instance should not refresh expiry")
	}
	if !resp.ExpiresAt.Equal(existingExpiry) || !existingInstance.ExpiresAt.Equal(existingExpiry) {
		t.Fatalf("expected stopping awd expiry to remain unchanged, resp=%s stored=%s want=%s", resp.ExpiresAt, existingInstance.ExpiresAt, existingExpiry)
	}
	if resp.Status != "destroying" {
		t.Fatalf("expected stopping awd instance to be exposed as destroying, got %+v", resp)
	}
	if resp.AccessURL != "" || resp.Access != nil {
		t.Fatalf("expected destroying awd response access to be cleared, got %+v", resp)
	}
}

func TestRestartOrStartScopedAWDServiceRecreatesActiveInstanceWhenCheckerTokenMetadataMissing(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	contestID := int64(3319)
	teamID := int64(4319)
	serviceID := int64(7319)
	challengeID := int64(2319)
	userID := int64(5319)
	existingInstance := &instanceentity.Instance{
		ID:          9319,
		UserID:      userID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
		ShareScope:  instanceentity.ShareScopePerTeam,
		Status:      instanceentity.InstanceStatusRunning,
		ExpiresAt:   now.Add(10 * time.Minute),
		RuntimeDetails: `{"networks":[{"name":"ctf-awd-contest-3319","shared":true}],
			"containers":[{"container_id":"ctr-3319","is_entry_point":true,"service_port":8080}]}`,
	}
	scope := practiceports.InstanceScope{
		ContestMode: practiceports.ContestModeAWD,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ShareScope:  instanceentity.ShareScopePerTeam,
	}

	cleanupCalled := false
	resetCalled := false
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			if gotContestID != contestID {
				t.Fatalf("unexpected contest lookup: %d", gotContestID)
			}
			return &practiceports.ContestRecord{
				ID:      contestID,
				Mode:    practiceports.ContestModeAWD,
				Status:  practiceports.ContestStatusRunning,
				EndTime: now.Add(time.Hour),
			}, nil
		},
		findScopedRestartableInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, gotScope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			if gotUserID != userID || gotChallengeID != challengeID {
				t.Fatalf("unexpected restartable lookup: user=%d challenge=%d", gotUserID, gotChallengeID)
			}
			if gotScope.ServiceID == nil || *gotScope.ServiceID != serviceID {
				t.Fatalf("unexpected scope: %+v", gotScope)
			}
			return existingInstance, nil
		},
		findContestAWDServiceRuntimeSubjectFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRuntimeSubject, error) {
			if gotContestID != contestID || gotServiceID != serviceID {
				t.Fatalf("unexpected runtime subject lookup: contest=%d service=%d", gotContestID, gotServiceID)
			}
			return &practiceports.ContestAWDServiceRuntimeSubject{
				ServiceID:   serviceID,
				ChallengeID: challengeID,
				Visible:     true,
				WorkspaceConfig: &practiceports.ContestAWDDefenseWorkspaceConfig{
					CheckerTokenEnv: "CHECKER_TOKEN",
				},
			}, nil
		},
		resetInstanceRuntimeForRestartFn: func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
			resetCalled = true
			if instanceID != existingInstance.ID {
				t.Fatalf("unexpected reset instance: %d", instanceID)
			}
			if status != instanceentity.InstanceStatusPending {
				t.Fatalf("unexpected reset status: %s", status)
			}
			if preserveHostPort {
				t.Fatal("expected awd restart to avoid preserving host port")
			}
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, operation *runtimeentity.AWDServiceOperation) error {
			return nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, instance *instanceentity.Instance) error {
				cleanupCalled = true
				if instance.ID != existingInstance.ID {
					t.Fatalf("unexpected cleanup instance: %+v", instance)
				}
				return nil
			},
		},
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
		nil),

		repo, nil)

	resp, err := service.restartOrStartScopedAWDService(context.Background(), awdScopedRuntimeRequest{
		OwnerUserID:  userID,
		ContestID:    contestID,
		ChallengeID:  challengeID,
		Scope:        scope,
		NoopIfActive: true,
		Audit: awdScopedRuntimeAudit{
			RestartOperationType: runtimecontracts.AWDServiceOperationTypeRecreate,
			RequestedBy:          runtimecontracts.AWDServiceOperationRequestedBySystem,
			Reason:               "desired_runtime_reconcile",
		},
	})
	if err != nil {
		t.Fatalf("restartOrStartScopedAWDService() error = %v", err)
	}
	if resp == nil || resp.ID != existingInstance.ID {
		t.Fatalf("expected restarted instance response, got %+v", resp)
	}
	if !cleanupCalled || !resetCalled {
		t.Fatalf("expected active instance to be recreated, cleanup=%v reset=%v", cleanupCalled, resetCalled)
	}
	if resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected pending status after restart scheduling, got %+v", resp)
	}
}

func TestRestartContestAWDServiceRequeuesExistingTeamInstance(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4106)
	serviceID := int64(7106)
	contestID := int64(3106)
	userID := int64(5106)
	contestEnd := now.Add(4 * time.Hour).UTC()
	instance := &instanceentity.Instance{
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
		ShareScope:     instanceentity.ShareScopePerTeam,
		Status:         instanceentity.InstanceStatusRunning,
		AccessURL:      "http://127.0.0.1:32106",
		Nonce:          "nonce-keep",
		ExpiresAt:      now.Add(-time.Minute),
		MaxExtends:     2,
	}
	var cleanupInstanceID int64
	var resetStatus string
	var operation *runtimeentity.AWDServiceOperation
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{ID: gotContestID, Mode: practiceports.ContestModeAWD, Status: practiceports.ContestStatusRunning, EndTime: contestEnd}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			if gotContestID != contestID || gotServiceID != serviceID {
				t.Fatalf("unexpected awd service lookup: contest=%d service=%d", gotContestID, gotServiceID)
			}
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2106,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":106,"instance_sharing":"per_team"},"flag_config":{"flag_type":"dynamic","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		findScopedRestartableInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			if gotUserID != userID || gotChallengeID != 2106 {
				t.Fatalf("unexpected scoped lookup: user=%d challenge=%d", gotUserID, gotChallengeID)
			}
			if scope.ServiceID == nil || *scope.ServiceID != serviceID || scope.TeamID == nil || *scope.TeamID != teamID || scope.ShareScope != instanceentity.ShareScopePerTeam {
				t.Fatalf("unexpected restart scope: %+v", scope)
			}
			return instance, nil
		},
		resetInstanceRuntimeForRestartFn: func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
			if instanceID != instance.ID {
				t.Fatalf("unexpected reset instance id: %d", instanceID)
			}
			if !expiresAt.Equal(contestEnd) {
				t.Fatalf("expected restart expiry to follow contest end time %s, got %s", contestEnd, expiresAt)
			}
			if preserveHostPort {
				t.Fatal("AWD restart must clear historical host port instead of preserving it")
			}
			resetStatus = status
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, got *runtimeentity.AWDServiceOperation) error {
			operation = got
			return nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, got *instanceentity.Instance) error {
				cleanupInstanceID = got.ID
				if got.HostPort != 0 {
					t.Fatalf("restart cleanup should preserve port allocation, got host_port=%d", got.HostPort)
				}
				return nil
			},
		},
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
		nil),

		repo, &stubPracticeChallengeContract{})

	resp, err := service.RestartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("RestartContestAWDService() error = %v", err)
	}
	if resp.ID != instance.ID || resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected same pending instance, got %+v", resp)
	}
	if cleanupInstanceID != instance.ID {
		t.Fatalf("expected cleanup for instance %d, got %d", instance.ID, cleanupInstanceID)
	}
	if resetStatus != instanceentity.InstanceStatusPending {
		t.Fatalf("expected reset to pending, got %q", resetStatus)
	}
	if !resp.ExpiresAt.Equal(contestEnd) || !instance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("restart should refresh awd instance expiry to contest end time, resp=%s instance=%s expected=%s", resp.ExpiresAt, instance.ExpiresAt, contestEnd)
	}
	if instance.ServiceID == nil || *instance.ServiceID != serviceID || instance.Nonce != "nonce-keep" || instance.HostPort != 0 {
		t.Fatalf("restart should preserve identity fields, got %+v", instance)
	}
	if instance.ContainerID != "" || instance.NetworkID != "" || instance.RuntimeDetails != "" || instance.AccessURL != "" {
		t.Fatalf("restart should clear runtime fields, got %+v", instance)
	}
	if operation == nil || operation.OperationType != runtimeentity.AWDServiceOperationTypeRestart || operation.RequestedBy != runtimeentity.AWDServiceOperationRequestedByUser || !operation.SLABillable {
		t.Fatalf("expected billable user restart operation, got %+v", operation)
	}
}

func TestRestartContestAWDServicePreservesHostPortWhenAccessHostConfigured(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4116)
	serviceID := int64(7116)
	contestID := int64(3116)
	userID := int64(5116)
	contestEnd := now.Add(3 * time.Hour).UTC()
	instance := &instanceentity.Instance{
		ID:             9116,
		UserID:         userID,
		ContestID:      &contestID,
		TeamID:         &teamID,
		ChallengeID:    2116,
		ServiceID:      &serviceID,
		HostPort:       32116,
		ContainerID:    "old-container",
		NetworkID:      "old-network",
		RuntimeDetails: `{"containers":[{"id":"old-container","host_port":32116}]}`,
		ShareScope:     instanceentity.ShareScopePerTeam,
		Status:         instanceentity.InstanceStatusRunning,
		AccessURL:      "http://host-gateway.internal:32116",
		Nonce:          "nonce-keep",
		ExpiresAt:      now.Add(-time.Minute),
		MaxExtends:     2,
	}
	var cleanupInstanceID int64
	var resetStatus string
	var preserveHostPortArg bool
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{ID: gotContestID, Mode: practiceports.ContestModeAWD, Status: practiceports.ContestStatusRunning, EndTime: contestEnd}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2116,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":116,"instance_sharing":"per_team"},"flag_config":{"flag_type":"dynamic","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		findScopedRestartableInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			return instance, nil
		},
		resetInstanceRuntimeForRestartFn: func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
			if !expiresAt.Equal(contestEnd) {
				t.Fatalf("expected restart expiry to follow contest end time %s, got %s", contestEnd, expiresAt)
			}
			preserveHostPortArg = preserveHostPort
			resetStatus = status
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, got *runtimeentity.AWDServiceOperation) error {
			return nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, got *instanceentity.Instance) error {
				cleanupInstanceID = got.ID
				if got.HostPort != 0 {
					t.Fatalf("restart cleanup must not release preserved host port allocation, got host_port=%d", got.HostPort)
				}
				return nil
			},
		},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				AccessHost:           "host-gateway.internal",
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil),

		repo, &stubPracticeChallengeContract{})

	resp, err := service.RestartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("RestartContestAWDService() error = %v", err)
	}
	if resp.ID != instance.ID || resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected same pending instance, got %+v", resp)
	}
	if cleanupInstanceID != instance.ID {
		t.Fatalf("expected cleanup for instance %d, got %d", instance.ID, cleanupInstanceID)
	}
	if resetStatus != instanceentity.InstanceStatusPending {
		t.Fatalf("expected reset to pending, got %q", resetStatus)
	}
	if !preserveHostPortArg {
		t.Fatal("expected restart to preserve awd host port when access host is configured")
	}
	if !resp.ExpiresAt.Equal(contestEnd) || !instance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("restart should refresh awd instance expiry to contest end time, resp=%s instance=%s expected=%s", resp.ExpiresAt, instance.ExpiresAt, contestEnd)
	}
	if instance.HostPort != 32116 {
		t.Fatalf("expected preserved awd host port, got %+v", instance)
	}
	if instance.ContainerID != "" || instance.NetworkID != "" || instance.RuntimeDetails != "" || instance.AccessURL != "" {
		t.Fatalf("restart should clear runtime fields, got %+v", instance)
	}
}

func TestRestartContestAWDServiceAllocatesHostPortWhenAccessHostConfiguredAndInstanceHasNone(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4117)
	serviceID := int64(7117)
	contestID := int64(3117)
	userID := int64(5117)
	contestEnd := now.Add(2 * time.Hour).UTC()
	instance := &instanceentity.Instance{
		ID:             9117,
		UserID:         userID,
		ContestID:      &contestID,
		TeamID:         &teamID,
		ChallengeID:    2117,
		ServiceID:      &serviceID,
		HostPort:       0,
		ContainerID:    "old-container",
		NetworkID:      "old-network",
		RuntimeDetails: `{"containers":[{"id":"old-container"}]}`,
		ShareScope:     instanceentity.ShareScopePerTeam,
		Status:         instanceentity.InstanceStatusRunning,
		AccessURL:      "http://awd-c3117-t4117-s7117:8080",
		Nonce:          "nonce-keep",
		ExpiresAt:      now.Add(-time.Minute),
		MaxExtends:     2,
	}
	var reserved bool
	var bound bool
	var preserveHostPortArg bool
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{ID: gotContestID, Mode: practiceports.ContestModeAWD, Status: practiceports.ContestStatusRunning, EndTime: contestEnd}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2117,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":117,"instance_sharing":"per_team"},"flag_config":{"flag_type":"dynamic","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		findScopedRestartableInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			return instance, nil
		},
		reserveAvailablePortFn: func(ctx context.Context, start, end int) (int, error) {
			reserved = true
			return 32117, nil
		},
		bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
			bound = true
			if port != 32117 || instanceID != instance.ID {
				t.Fatalf("unexpected reserved host port bind: port=%d instance=%d", port, instanceID)
			}
			return nil
		},
		resetInstanceRuntimeForRestartFn: func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
			if !expiresAt.Equal(contestEnd) {
				t.Fatalf("expected restart expiry to follow contest end time %s, got %s", contestEnd, expiresAt)
			}
			preserveHostPortArg = preserveHostPort
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, got *runtimeentity.AWDServiceOperation) error {
			return nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, got *instanceentity.Instance) error {
				if got.HostPort != 0 {
					t.Fatalf("restart cleanup must not release newly reserved host port allocation, got host_port=%d", got.HostPort)
				}
				return nil
			},
		},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				AccessHost:           "host-gateway.internal",
				PortRangeStart:       32000,
				PortRangeEnd:         32150,
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil),

		repo, &stubPracticeChallengeContract{})

	resp, err := service.RestartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("RestartContestAWDService() error = %v", err)
	}
	if resp.ID != instance.ID || resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected same pending instance, got %+v", resp)
	}
	if !reserved || !bound {
		t.Fatalf("expected restart to reserve and bind a host port, reserved=%v bound=%v", reserved, bound)
	}
	if !preserveHostPortArg {
		t.Fatal("expected reset path to preserve the newly reserved host port")
	}
	if !resp.ExpiresAt.Equal(contestEnd) || !instance.ExpiresAt.Equal(contestEnd) {
		t.Fatalf("restart should refresh awd instance expiry to contest end time, resp=%s instance=%s expected=%s", resp.ExpiresAt, instance.ExpiresAt, contestEnd)
	}
	if instance.HostPort != 32117 {
		t.Fatalf("expected instance host port to be backfilled before reprovision, got %+v", instance)
	}
}

func TestRestartContestAWDServiceReallocatesStaleHostPortWhenOwnedByAnotherInstance(t *testing.T) {
	t.Parallel()

	now := time.Now()
	teamID := int64(4118)
	serviceID := int64(7118)
	contestID := int64(3118)
	userID := int64(5118)
	contestEnd := now.Add(2 * time.Hour).UTC()
	instance := &instanceentity.Instance{
		ID:             9118,
		UserID:         userID,
		ContestID:      &contestID,
		TeamID:         &teamID,
		ChallengeID:    2118,
		ServiceID:      &serviceID,
		HostPort:       32118,
		ContainerID:    "old-container",
		NetworkID:      "old-network",
		RuntimeDetails: `{"containers":[{"id":"old-container","host_port":32118}]}`,
		ShareScope:     instanceentity.ShareScopePerTeam,
		Status:         instanceentity.InstanceStatusRunning,
		AccessURL:      "http://host-gateway.internal:32118",
		Nonce:          "nonce-keep",
		ExpiresAt:      now.Add(-time.Minute),
		MaxExtends:     2,
	}
	var reusableChecked bool
	var reserved bool
	var bound bool
	var preserveHostPortArg bool
	repo := &stubPracticeRepository{
		findContestByIDFn: func(ctx context.Context, gotContestID int64) (*practiceports.ContestRecord, error) {
			return &practiceports.ContestRecord{ID: gotContestID, Mode: practiceports.ContestModeAWD, Status: practiceports.ContestStatusRunning, EndTime: contestEnd}, nil
		},
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				AWDChallengeID:  2118,
				IsVisible:       true,
				ServiceSnapshot: `{"name":"awd-service","category":"web","difficulty":"medium","runtime_config":{"image_id":118,"instance_sharing":"per_team"},"flag_config":{"flag_type":"dynamic","flag_prefix":"flag"}}`,
			}, nil
		},
		findContestRegistrationFn: func(ctx context.Context, gotContestID, gotUserID int64) (*practiceports.ContestParticipation, error) {
			return &practiceports.ContestParticipation{TeamID: &teamID, Status: contestentity.ContestRegistrationStatusApproved}, nil
		},
		findScopedRestartableInstanceFn: func(ctx context.Context, gotUserID, gotChallengeID int64, scope practiceports.InstanceScope) (*instanceentity.Instance, error) {
			return instance, nil
		},
		isHostPortReusableForRestartFn: func(ctx context.Context, instanceID int64, hostPort int) (bool, error) {
			reusableChecked = true
			if instanceID != instance.ID || hostPort != 32118 {
				t.Fatalf("unexpected host port reuse check: instance=%d host_port=%d", instanceID, hostPort)
			}
			return false, nil
		},
		reserveAvailablePortExcludingFn: func(ctx context.Context, start, end, excludedPort int) (int, error) {
			reserved = true
			if excludedPort != 32118 {
				t.Fatalf("expected stale host port 32118 to be excluded, got %d", excludedPort)
			}
			return 32119, nil
		},
		bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
			bound = true
			if port != 32119 || instanceID != instance.ID {
				t.Fatalf("unexpected rebound host port bind: port=%d instance=%d", port, instanceID)
			}
			return nil
		},
		resetInstanceRuntimeForRestartFn: func(ctx context.Context, instanceID int64, status string, expiresAt time.Time, preserveHostPort bool) error {
			if !expiresAt.Equal(contestEnd) {
				t.Fatalf("expected restart expiry to follow contest end time %s, got %s", contestEnd, expiresAt)
			}
			preserveHostPortArg = preserveHostPort
			return nil
		},
		createAWDServiceOperationFn: func(ctx context.Context, got *runtimeentity.AWDServiceOperation) error {
			return nil
		},
	}

	service := wirePracticeScopeAdapters(NewService(
		repo,

		nil,
		&stubPracticeInstanceStore{},
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, got *instanceentity.Instance) error {
				if got.HostPort != 0 {
					t.Fatalf("restart cleanup must not release stale host port allocation, got host_port=%d", got.HostPort)
				}
				return nil
			},
		},
		nil,
		nil,
		&config.Config{
			Container: config.ContainerConfig{
				AccessHost:           "host-gateway.internal",
				PortRangeStart:       32000,
				PortRangeEnd:         32150,
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				Scheduler: config.ContainerSchedulerConfig{
					Enabled: true,
				},
			},
		},
		nil),

		repo, &stubPracticeChallengeContract{})

	resp, err := service.RestartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("RestartContestAWDService() error = %v", err)
	}
	if resp.ID != instance.ID || resp.Status != instanceentity.InstanceStatusPending {
		t.Fatalf("expected same pending instance, got %+v", resp)
	}
	if !reusableChecked || !reserved || !bound {
		t.Fatalf("expected stale port to be checked and reallocated, checked=%v reserved=%v bound=%v", reusableChecked, reserved, bound)
	}
	if !preserveHostPortArg {
		t.Fatal("expected reset path to preserve the rebound host port")
	}
	if instance.HostPort != 32119 {
		t.Fatalf("expected stale host port to be replaced before reprovision, got %+v", instance)
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

	if err := db.Create(&practiceCommandImageRow{
		ID:        imageID,
		Name:      "ctf/awd-runtime",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&contestentity.Contest{
		ID:        contestID,
		Title:     "AWD Restart",
		Mode:      contestentity.ContestModeAWD,
		Status:    contestentity.ContestStatusRunning,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&identitycontracts.User{ID: userID, Username: "restart-student", Role: identitycontracts.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&contestentity.ContestRegistration{
		ContestID: contestID,
		UserID:    userID,
		TeamID:    &teamID,
		Status:    contestentity.ContestRegistrationStatusApproved,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}
	serviceSnapshot, err := contestentity.EncodeContestAWDServiceSnapshot(contestentity.ContestAWDServiceSnapshot{
		Name: "Restart Service",
		RuntimeConfig: map[string]any{
			"image_id":         imageID,
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
	if err := db.Create(&contestentity.ContestAWDService{
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

	instance := &instanceentity.Instance{
		ID:          9201,
		UserID:      userID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
		Status:      instanceentity.InstanceStatusRunning,
		ShareScope:  instanceentity.ShareScopePerTeam,
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
	if err := db.Create(&runtimeentity.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        instance.ID,
		WorkspaceRevision: 4,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-existing",
		SeedSignature:     "seed:v1",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create workspace row: %v", err)
	}

	var createTopologyCalls atomic.Int32
	service := wirePracticeScopeAdapters(NewService(
		practiceinfra.NewRepository(db),

		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(context.Context, *instanceentity.Instance) error { return nil },
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
					RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
						Networks: []runtimecontracts.InstanceRuntimeNetwork{
							{Key: runtimecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-9101", NetworkID: "net-awd-contest-9101", Shared: true},
						},
						Containers: []runtimecontracts.InstanceRuntimeContainer{
							{NodeKey: "default", ContainerID: "runtime-new", ServicePort: 8080, IsEntryPoint: true, NetworkAliases: []string{"awd-c9101-t9102-s9103"}},
						},
					},
				}, nil
			},
		},
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
		nil),

		practiceinfra.NewRepository(db), challengeinfra.NewRepository(db))

	resp, err := service.RestartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("RestartContestAWDService() error = %v", err)
	}
	if resp.Status != instanceentity.InstanceStatusRunning {
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

	if err := db.Create(&practiceCommandImageRow{
		ID:        imageID,
		Name:      "ctf/awd-runtime",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}
	if err := db.Create(&practiceCommandChallengeRow{
		ID:        challengeID,
		Title:     "Restart Service",
		ImageID:   imageID,
		Status:    challengecontracts.ChallengeStatusPublished,
		FlagType:  challengecontracts.FlagTypeStatic,
		FlagHash:  "flag{restart}",
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := db.Create(&contestentity.Contest{
		ID:        contestID,
		Title:     "AWD Restart Missing Workspace",
		Mode:      contestentity.ContestModeAWD,
		Status:    contestentity.ContestStatusRunning,
		StartTime: now.Add(-time.Hour),
		EndTime:   now.Add(time.Hour),
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create contest: %v", err)
	}
	if err := db.Create(&identitycontracts.User{ID: userID, Username: "restart-student-missing", Role: identitycontracts.RoleStudent, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := db.Create(&contestentity.ContestRegistration{
		ContestID: contestID,
		UserID:    userID,
		TeamID:    &teamID,
		Status:    contestentity.ContestRegistrationStatusApproved,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create registration: %v", err)
	}
	serviceSnapshot, err := contestentity.EncodeContestAWDServiceSnapshot(contestentity.ContestAWDServiceSnapshot{
		Name: "Restart Service Missing Workspace",
		RuntimeConfig: map[string]any{
			"image_id":         imageID,
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
	if err := db.Create(&contestentity.ContestAWDService{
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

	instance := &instanceentity.Instance{
		ID:          9211,
		UserID:      userID,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
		Status:      instanceentity.InstanceStatusRunning,
		ShareScope:  instanceentity.ShareScopePerTeam,
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
	if err := db.Create(&runtimeentity.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        instance.ID,
		WorkspaceRevision: 4,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-missing",
		SeedSignature:     "seed:v1",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create workspace row: %v", err)
	}

	var createTopologyCalls atomic.Int32
	service := wirePracticeScopeAdapters(NewService(
		practiceinfra.NewRepository(db),

		challengeinfra.NewImageRepository(db),
		runtimeinfrarepo.NewRepository(db),
		&stubPracticeRuntimeService{
			cleanupRuntimeFn: func(context.Context, *instanceentity.Instance) error { return nil },
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
						RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
							Networks: []runtimecontracts.InstanceRuntimeNetwork{
								{Key: runtimecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-9111", NetworkID: "net-awd-contest-9111", Shared: true},
							},
							Containers: []runtimecontracts.InstanceRuntimeContainer{
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
					if len(req.Nodes) != 1 || len(req.Nodes[0].NetworkAliases) != 1 || req.Nodes[0].NetworkAliases[0] != "awd-ws-c9111-t9112-s9113-r4" {
						t.Fatalf("expected recreated workspace alias in second topology create, got %+v", req)
					}
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "workspace-recreated",
						NetworkID:          "net-awd-contest-9111",
						AccessURL:          "tcp://172.30.0.55:22",
						RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
							Containers: []runtimecontracts.InstanceRuntimeContainer{
								{NodeKey: "workspace", ContainerID: "workspace-recreated", ServicePort: 22, ServiceProtocol: challengecontracts.ChallengeTargetProtocolTCP, IsEntryPoint: true},
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
		&config.Config{
			Container: config.ContainerConfig{
				FlagGlobalSecret:     "restart-secret",
				DefaultTTL:           time.Hour,
				MaxConcurrentPerUser: 3,
				CreateTimeout:        time.Second,
				Scheduler:            config.ContainerSchedulerConfig{Enabled: false},
			},
		},
		nil),

		practiceinfra.NewRepository(db), challengeinfra.NewRepository(db))

	resp, err := service.RestartContestAWDService(context.Background(), userID, contestID, serviceID)
	if err != nil {
		t.Fatalf("RestartContestAWDService() error = %v", err)
	}
	if resp.Status != instanceentity.InstanceStatusRunning {
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
		practiceinfra.NewRepository(db),

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

		practiceinfra.NewRepository(db), challengeinfra.NewRepository(db))

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
		practiceinfra.NewRepository(db),

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

		practiceinfra.NewRepository(db), challengeinfra.NewRepository(db))

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
