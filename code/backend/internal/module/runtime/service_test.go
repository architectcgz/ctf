package runtime_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeqry "ctf-platform/internal/module/runtime/application/queries"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"ctf-platform/pkg/errcode"
)

func TestRepositoryListActiveContainerIDs(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	seedInstance(t, repo.db, &model.Instance{
		UserID:      1,
		ChallengeID: 101,
		ContainerID: "running-container",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedInstance(t, repo.db, &model.Instance{
		UserID:         1,
		ChallengeID:    102,
		ContainerID:    "creating-container",
		RuntimeDetails: `{"containers":[{"container_id":"sidecar-1"},{"container_id":"creating-container"}]}`,
		Status:         model.InstanceStatusCreating,
		ExpiresAt:      time.Now().Add(time.Hour),
	})
	seedInstance(t, repo.db, &model.Instance{
		UserID:      1,
		ChallengeID: 103,
		ContainerID: "stopped-container",
		Status:      model.InstanceStatusStopped,
		ExpiresAt:   time.Now().Add(time.Hour),
	})
	seedInstance(t, repo.db, &model.Instance{
		UserID:      1,
		ChallengeID: 104,
		ContainerID: "",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   time.Now().Add(time.Hour),
	})

	containerIDs, err := repo.ListActiveContainerIDs()
	if err != nil {
		t.Fatalf("ListActiveContainerIDs() error = %v", err)
	}
	if len(containerIDs) != 3 {
		t.Fatalf("expected 3 active container ids, got %d (%v)", len(containerIDs), containerIDs)
	}

	got := make(map[string]struct{}, len(containerIDs))
	for _, containerID := range containerIDs {
		got[containerID] = struct{}{}
	}
	if _, exists := got["running-container"]; !exists {
		t.Fatalf("running container not returned: %v", containerIDs)
	}
	if _, exists := got["creating-container"]; !exists {
		t.Fatalf("creating container not returned: %v", containerIDs)
	}
	if _, exists := got["sidecar-1"]; !exists {
		t.Fatalf("sidecar container not returned: %v", containerIDs)
	}
}

func TestRepositoryUpdateStatusAndReleasePortRemovesAllocation(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now()
	instance := &model.Instance{
		ID:          2001,
		UserID:      1,
		ChallengeID: 101,
		HostPort:    30001,
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	seedInstance(t, repo.db, instance)
	if err := repo.db.Create(&model.PortAllocation{
		Port:       30001,
		InstanceID: &instance.ID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}

	if err := repo.UpdateStatusAndReleasePort(instance.ID, model.InstanceStatusFailed); err != nil {
		t.Fatalf("UpdateStatusAndReleasePort() error = %v", err)
	}

	updated, err := repo.FindByID(context.Background(), instance.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if updated.Status != model.InstanceStatusFailed {
		t.Fatalf("expected failed status, got %+v", updated)
	}

	var count int64
	if err := repo.db.Model(&model.PortAllocation{}).Where("port = ?", 30001).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected port allocation to be removed, count=%d", count)
	}
}

func TestServiceCreateContainerCreatesIsolatedNetwork(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:           "net-123",
		containerID:         "ctr-123",
		resolvedServicePort: 80,
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart:     30000,
		PortRangeEnd:       30010,
		DefaultExposedPort: 8080,
	}, nil)

	containerID, networkID, hostPort, servicePort, err := service.CreateContainer(context.Background(), "ctf/web:v1", map[string]string{"FLAG": "flag{1}"}, 0)
	if err != nil {
		t.Fatalf("CreateContainer() error = %v", err)
	}
	if containerID != "ctr-123" {
		t.Fatalf("unexpected container id: %s", containerID)
	}
	if networkID != "net-123" {
		t.Fatalf("unexpected network id: %s", networkID)
	}
	if hostPort != 30000 {
		t.Fatalf("unexpected host port: %d", hostPort)
	}
	if servicePort != 80 {
		t.Fatalf("unexpected service port: %d", servicePort)
	}
	if engine.createdNetworkName == "" {
		t.Fatalf("expected isolated network to be created")
	}
	if engine.createdContainerCfg == nil || engine.createdContainerCfg.Network != engine.createdNetworkName {
		t.Fatalf("expected container to join created network, cfg=%+v network=%s", engine.createdContainerCfg, engine.createdNetworkName)
	}
	if _, exists := engine.createdContainerCfg.Ports["80"]; !exists {
		t.Fatalf("expected container to publish resolved service port 80, got %+v", engine.createdContainerCfg.Ports)
	}
}

func TestServiceCreateContainerFailsWhenRuntimeEngineUnavailable(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := runtimecmd.NewProvisioningService(repo, nil, &config.ContainerConfig{
		PortRangeStart:     30000,
		PortRangeEnd:       30010,
		DefaultExposedPort: 8080,
		PublicHost:         "127.0.0.1",
	}, nil)

	containerID, networkID, hostPort, servicePort, err := service.CreateContainer(context.Background(), "ctf/web:v1", nil, 0)
	if err == nil {
		t.Fatal("expected CreateContainer() to fail when runtime engine is unavailable")
	}
	if !strings.Contains(err.Error(), "runtime engine is not configured") {
		t.Fatalf("expected runtime engine unavailable error, got %v", err)
	}
	if containerID != "" || networkID != "" || hostPort != 0 || servicePort != 0 {
		t.Fatalf("expected zero runtime result on failure, got container=%q network=%q hostPort=%d servicePort=%d", containerID, networkID, hostPort, servicePort)
	}
}

func TestServiceCreateContainerRemovesNetworkWhenStartFails(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:   "net-456",
		containerID: "ctr-456",
		startErr:    errors.New("start failed"),
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart:     30000,
		PortRangeEnd:       30010,
		DefaultExposedPort: 8080,
	}, nil)

	_, _, _, _, err := service.CreateContainer(context.Background(), "ctf/web:v1", nil, 0)
	if err == nil {
		t.Fatal("expected CreateContainer() to fail")
	}
	if engine.removedContainerID != "ctr-456" {
		t.Fatalf("expected container cleanup, got %s", engine.removedContainerID)
	}
	if engine.removedNetworkID != "net-456" {
		t.Fatalf("expected network cleanup, got %s", engine.removedNetworkID)
	}
}

func TestServiceRemoveContainerFailsWhenRuntimeEngineUnavailable(t *testing.T) {
	t.Parallel()

	cleanupService := runtimecmd.NewRuntimeCleanupService(nil, nil)

	err := cleanupService.RemoveContainer(context.Background(), "ctr-missing-engine")
	if err == nil {
		t.Fatal("expected RemoveContainer() to fail when runtime engine is unavailable")
	}
	if !strings.Contains(err.Error(), "runtime engine is not configured") {
		t.Fatalf("expected runtime engine unavailable error, got %v", err)
	}
}

func TestServiceRemoveContainerHonorsCancellation(t *testing.T) {
	t.Parallel()

	engine := &fakeRuntimeEngine{
		removeContainerFn: func(ctx context.Context, containerID string, force bool) error {
			if containerID != "ctr-ctx" || !force {
				t.Fatalf("unexpected remove container args: id=%s force=%v", containerID, force)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := cleanupService.RemoveContainer(ctx, "ctr-ctx"); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestServiceCleanupRuntimeFailsWhenRuntimeEngineUnavailable(t *testing.T) {
	t.Parallel()

	cleanupService := runtimecmd.NewRuntimeCleanupService(nil, nil)
	instance := &model.Instance{
		ID:          3002,
		ContainerID: "ctr-missing-engine",
		NetworkID:   "net-missing-engine",
	}

	err := cleanupService.CleanupRuntime(context.Background(), instance)
	if err == nil {
		t.Fatal("expected CleanupRuntime() to fail when runtime engine is unavailable")
	}
	if !strings.Contains(err.Error(), "runtime engine is not configured") {
		t.Fatalf("expected runtime engine unavailable error, got %v", err)
	}
}

func TestServiceCleanupRuntimeHonorsCancellation(t *testing.T) {
	t.Parallel()

	engine := &fakeRuntimeEngine{
		removeContainerFn: func(ctx context.Context, containerID string, force bool) error {
			if containerID != "ctr-3001" || !force {
				t.Fatalf("unexpected remove container args: id=%s force=%v", containerID, force)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil)

	instance := &model.Instance{
		ID:          3001,
		ContainerID: "ctr-3001",
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := cleanupService.CleanupRuntime(ctx, instance); !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestServiceDestroyInstanceAllowsContestTeamMember(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(301)
	teamID := int64(401)

	if err := repo.db.Create(&model.Team{ID: teamID, ContestID: contestID, Name: "Alpha", CaptainID: 1, InviteCode: "alpha", MaxMembers: 4, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&model.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 2, JoinedAt: now, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
	seedInstance(t, repo.db, &model.Instance{
		ID:          901,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 101,
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	if err := service.DestroyInstance(context.Background(), 901, 2); err != nil {
		t.Fatalf("DestroyInstance() error = %v", err)
	}

	instance, err := repo.FindByID(context.Background(), 901)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if instance.Status != model.InstanceStatusStopped {
		t.Fatalf("expected stopped status, got %s", instance.Status)
	}
}

func TestServiceExtendInstanceAllowsContestTeamMember(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(302)
	teamID := int64(402)

	if err := repo.db.Create(&model.Team{ID: teamID, ContestID: contestID, Name: "Beta", CaptainID: 1, InviteCode: "beta", MaxMembers: 4, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&model.TeamMember{ContestID: contestID, TeamID: teamID, UserID: 2, JoinedAt: now, CreatedAt: now}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}
	initialExpiry := now.Add(time.Hour)
	seedInstance(t, repo.db, &model.Instance{
		ID:          902,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 102,
		ContainerID: "contest-shared-extend",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   initialExpiry,
	})

	resp, err := service.ExtendInstance(context.Background(), 902, 2)
	if err != nil {
		t.Fatalf("ExtendInstance() error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected extend response")
	}
	if resp.RemainingExtends != 1 {
		t.Fatalf("expected remaining extends 1, got %d", resp.RemainingExtends)
	}

	instance, err := repo.FindByID(context.Background(), 902)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if !instance.ExpiresAt.After(initialExpiry) {
		t.Fatalf("expected expiry to be extended, got %s", instance.ExpiresAt)
	}
	if instance.ExtendCount != 1 {
		t.Fatalf("expected extend count 1, got %d", instance.ExtendCount)
	}
}

func TestServiceDestroyInstanceRejectsSharedInstance(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	if err := repo.db.Create(&model.Challenge{
		ID:              903,
		Title:           "Shared Practice",
		Category:        model.DimensionWeb,
		Difficulty:      model.ChallengeDifficultyEasy,
		FlagType:        model.FlagTypeStatic,
		Status:          model.ChallengeStatusPublished,
		InstanceSharing: model.InstanceSharingShared,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	seedInstance(t, repo.db, &model.Instance{
		ID:          903,
		UserID:      1,
		ChallengeID: 903,
		ShareScope:  model.InstanceSharingShared,
		ContainerID: "shared-ctr",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	err := service.DestroyInstance(context.Background(), 903, 2)
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden for shared destroy, got %v", err)
	}
}

func TestServiceExtendInstanceRejectsSharedInstance(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	if err := repo.db.Create(&model.Challenge{
		ID:              904,
		Title:           "Shared Practice",
		Category:        model.DimensionWeb,
		Difficulty:      model.ChallengeDifficultyEasy,
		FlagType:        model.FlagTypeStatic,
		Status:          model.ChallengeStatusPublished,
		InstanceSharing: model.InstanceSharingShared,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	seedInstance(t, repo.db, &model.Instance{
		ID:          904,
		UserID:      1,
		ChallengeID: 904,
		ShareScope:  model.InstanceSharingShared,
		ContainerID: "shared-ctr",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	_, err := service.ExtendInstance(context.Background(), 904, 2)
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden for shared extend, got %v", err)
	}
}

func TestServiceGetUserInstancesIncludesChallengeMetadata(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	if err := repo.db.Create(&model.Challenge{
		ID:         101,
		Title:      "Matrix Web Challenge",
		Category:   model.DimensionWeb,
		Difficulty: model.ChallengeDifficultyEasy,
		FlagType:   model.FlagTypeStatic,
		Status:     model.ChallengeStatusPublished,
		Points:     100,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	seedInstance(t, repo.db, &model.Instance{
		ID:          1001,
		UserID:      1,
		ChallengeID: 101,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30001",
		ExpiresAt:   now.Add(time.Hour),
		ExtendCount: 1,
		MaxExtends:  3,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	items, err := service.GetUserInstances(context.Background(), 1)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 instance, got %+v", items)
	}
	item := items[0]
	if item.ChallengeTitle != "Matrix Web Challenge" {
		t.Fatalf("expected challenge title, got %+v", item)
	}
	if item.Category != model.DimensionWeb {
		t.Fatalf("expected category %q, got %+v", model.DimensionWeb, item)
	}
	if item.Difficulty != model.ChallengeDifficultyEasy {
		t.Fatalf("expected difficulty %q, got %+v", model.ChallengeDifficultyEasy, item)
	}
	if item.FlagType != model.FlagTypeStatic {
		t.Fatalf("expected flag type %q, got %+v", model.FlagTypeStatic, item)
	}
	if item.RemainingExtends != 2 {
		t.Fatalf("expected remaining extends 2, got %+v", item)
	}
}

func TestServiceGetUserInstancesShowsContestSharedInstanceToTeamMember(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()
	contestID := int64(501)
	teamID := int64(601)

	if err := repo.db.Create(&model.Challenge{
		ID:         102,
		Title:      "Shared AWD Challenge",
		Category:   model.DimensionPwn,
		Difficulty: model.ChallengeDifficultyMedium,
		FlagType:   model.FlagTypeDynamic,
		Status:     model.ChallengeStatusPublished,
		Points:     150,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}
	if err := repo.db.Create(&model.Team{
		ID:         teamID,
		ContestID:  contestID,
		Name:       "Runtime Team",
		CaptainID:  1,
		InviteCode: "runtime",
		MaxMembers: 4,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create team: %v", err)
	}
	if err := repo.db.Create(&model.TeamMember{
		ContestID: contestID,
		TeamID:    teamID,
		UserID:    2,
		JoinedAt:  now,
		CreatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create team member: %v", err)
	}

	seedInstance(t, repo.db, &model.Instance{
		ID:          1002,
		UserID:      1,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: 102,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30002",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	items, err := service.GetUserInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1002 {
		t.Fatalf("expected teammate visible shared instance, got %+v", items)
	}
}

func TestServiceGetUserInstancesShowsPracticeSharedInstanceToAnyUser(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	if err := repo.db.Create(&model.Challenge{
		ID:              103,
		Title:           "Shared Practice",
		Category:        model.DimensionWeb,
		Difficulty:      model.ChallengeDifficultyEasy,
		FlagType:        model.FlagTypeStatic,
		Status:          model.ChallengeStatusPublished,
		InstanceSharing: model.InstanceSharingShared,
		Points:          80,
		CreatedAt:       now,
		UpdatedAt:       now,
	}).Error; err != nil {
		t.Fatalf("create challenge: %v", err)
	}

	seedInstance(t, repo.db, &model.Instance{
		ID:          1003,
		UserID:      1,
		ChallengeID: 103,
		ShareScope:  model.InstanceSharingShared,
		Status:      model.InstanceStatusRunning,
		AccessURL:   "http://127.0.0.1:30003",
		ExpiresAt:   now.Add(time.Hour),
		MaxExtends:  2,
		CreatedAt:   now,
		UpdatedAt:   now,
	})

	items, err := service.GetUserInstances(context.Background(), 2)
	if err != nil {
		t.Fatalf("GetUserInstances() error = %v", err)
	}
	if len(items) != 1 || items[0].ID != 1003 {
		t.Fatalf("expected global shared instance visible to another user, got %+v", items)
	}
	if items[0].ShareScope != model.InstanceSharingShared {
		t.Fatalf("expected share scope to be returned, got %+v", items[0])
	}
}

func TestServiceCreateTopologyCreatesMultipleContainersOnSharedNetwork(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-789",
		containerIDs: []string{"web-ctr", "db-ctr"},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{model.TopologyDefaultNetworkKey}},
			{Key: "db", Image: "ctf/db:v1", NetworkKeys: []string{model.TopologyDefaultNetworkKey}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if result.PrimaryContainerID != "web-ctr" || result.NetworkID != "net-789" {
		t.Fatalf("unexpected topology result: %+v", result)
	}
	if len(result.RuntimeDetails.Containers) != 2 {
		t.Fatalf("unexpected runtime details: %+v", result.RuntimeDetails)
	}
	if len(engine.createdContainerCfgs) != 2 {
		t.Fatalf("expected two create container calls, got %d", len(engine.createdContainerCfgs))
	}
	if engine.createdContainerCfgs[0].Network != engine.createdNetworkName || engine.createdContainerCfgs[1].Network != engine.createdNetworkName {
		t.Fatalf("expected all containers to join shared network")
	}
	if _, exists := engine.createdContainerCfgs[1].Ports["8080"]; exists {
		t.Fatalf("non-entry container should not expose host port")
	}
}

func TestServiceDestroyManagedInstanceRemovesAllRuntimeContainers(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{}
	service := newTestRuntimeModule(repo, engine)

	instance := &model.Instance{
		ID:             1,
		UserID:         1,
		ChallengeID:    1,
		HostPort:       30001,
		ContainerID:    "web-ctr",
		NetworkID:      "net-1",
		RuntimeDetails: `{"containers":[{"container_id":"web-ctr"},{"container_id":"db-ctr"}],"acl_rules":[{"comment":"ctf:acl:test","source_ip":"172.30.0.2","target_ip":"172.30.0.3","action":"allow","protocol":"tcp","ports":[3306]}]}`,
		Status:         model.InstanceStatusRunning,
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	seedInstance(t, repo.db, instance)
	if err := repo.db.Create(&model.PortAllocation{Port: 30001, InstanceID: &instance.ID}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}

	if err := service.DestroyInstance(context.Background(), instance.ID, instance.UserID); err != nil {
		t.Fatalf("DestroyInstance() error = %v", err)
	}
	if len(engine.removedContainerIDs) != 2 {
		t.Fatalf("expected 2 removed containers, got %v", engine.removedContainerIDs)
	}
	if len(engine.removedNetworkIDs) != 1 || engine.removedNetworkIDs[0] != "net-1" {
		t.Fatalf("expected 1 removed network, got %v", engine.removedNetworkIDs)
	}
	if len(engine.removedACLRules) != 1 || engine.removedACLRules[0].Comment != "ctf:acl:test" {
		t.Fatalf("expected acl rules to be removed, got %+v", engine.removedACLRules)
	}

	updated, err := repo.FindByID(context.Background(), instance.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if updated.Status != model.InstanceStatusStopped {
		t.Fatalf("expected stopped status, got %+v", updated)
	}

	var count int64
	if err := repo.db.Model(&model.PortAllocation{}).Where("port = ?", 30001).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected port allocation to be removed, count=%d", count)
	}
}

func TestServiceCleanExpiredInstancesKeepsRunningStateWhenRuntimeCleanupFails(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now()
	instance := &model.Instance{
		ID:          2101,
		UserID:      1,
		ChallengeID: 1,
		HostPort:    30002,
		ContainerID: "web-ctr",
		NetworkID:   "net-2",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   now.Add(-time.Minute),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	seedInstance(t, repo.db, instance)
	if err := repo.db.Create(&model.PortAllocation{
		Port:       30002,
		InstanceID: &instance.ID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}

	engine := &fakeRuntimeEngine{removeContainerErr: errors.New("remove failed")}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil)
	service := runtimecmd.NewRuntimeMaintenanceService(repo, nil, cleanupService, &config.ContainerConfig{
		MaxExtends:        2,
		ExtendDuration:    30 * time.Minute,
		OrphanGracePeriod: 5 * time.Minute,
	}, nil)

	if err := service.CleanExpiredInstances(context.Background()); err != nil {
		t.Fatalf("CleanExpiredInstances() error = %v", err)
	}

	updated, err := repo.FindByID(context.Background(), instance.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if updated.Status != model.InstanceStatusRunning {
		t.Fatalf("expected instance to remain running for retry, got %+v", updated)
	}

	var count int64
	if err := repo.db.Model(&model.PortAllocation{}).Where("port = ?", 30002).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected port allocation to remain reserved, count=%d", count)
	}
}

func TestServiceCleanExpiredInstancesMarksExpiredWhenContainerAlreadyRemoved(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	now := time.Now()
	instance := &model.Instance{
		ID:          2102,
		UserID:      1,
		ChallengeID: 1,
		HostPort:    30003,
		ContainerID: "missing-ctr",
		NetworkID:   "net-3",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   now.Add(-time.Minute),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	seedInstance(t, repo.db, instance)
	if err := repo.db.Create(&model.PortAllocation{
		Port:       30003,
		InstanceID: &instance.ID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}).Error; err != nil {
		t.Fatalf("create port allocation: %v", err)
	}

	engine := &fakeRuntimeEngine{
		removeContainerErr: errors.New("Error response from daemon: No such container: missing-ctr"),
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil)
	service := runtimecmd.NewRuntimeMaintenanceService(repo, nil, cleanupService, &config.ContainerConfig{
		MaxExtends:        2,
		ExtendDuration:    30 * time.Minute,
		OrphanGracePeriod: 5 * time.Minute,
	}, nil)

	if err := service.CleanExpiredInstances(context.Background()); err != nil {
		t.Fatalf("CleanExpiredInstances() error = %v", err)
	}

	updated, err := repo.FindByID(context.Background(), instance.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if updated.Status != model.InstanceStatusExpired {
		t.Fatalf("expected instance to be marked expired, got %+v", updated)
	}

	var count int64
	if err := repo.db.Model(&model.PortAllocation{}).Where("port = ?", 30003).Count(&count).Error; err != nil {
		t.Fatalf("count port allocations: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected port allocation to be removed, count=%d", count)
	}
}

func TestServiceCreateTopologyCreatesAndConnectsMultipleNetworks(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkIDs:   []string{"net-public", "net-backend"},
		containerIDs: []string{"web-ctr", "db-ctr"},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: "public"},
			{Key: "backend", Internal: true},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{"public", "backend"}},
			{Key: "db", Image: "ctf/db:v1", NetworkKeys: []string{"backend"}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if result.NetworkID != "net-public" {
		t.Fatalf("unexpected primary network id: %+v", result)
	}
	if len(result.RuntimeDetails.Networks) != 2 {
		t.Fatalf("unexpected runtime networks: %+v", result.RuntimeDetails)
	}
	if len(engine.createdNetworkNames) != 2 {
		t.Fatalf("expected two created networks, got %v", engine.createdNetworkNames)
	}
	if len(engine.connectedNetworks["web-ctr"]) != 1 || engine.connectedNetworks["web-ctr"][0] != engine.createdNetworkNames[1] {
		t.Fatalf("expected web container to connect to backend network, got %+v", engine.connectedNetworks)
	}
	if len(engine.connectedNetworks["db-ctr"]) != 0 {
		t.Fatalf("db container should not need extra network connect, got %+v", engine.connectedNetworks)
	}
}

func TestServiceCreateTopologyAppliesFineGrainedACLRules(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-acl",
		containerIDs: []string{"web-ctr", "db-ctr"},
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			if len(engine.createdNetworkNames) == 0 {
				return nil
			}
			switch containerID {
			case "web-ctr":
				return map[string]string{engine.createdNetworkNames[0]: "172.30.0.2"}
			case "db-ctr":
				return map[string]string{engine.createdNetworkNames[0]: "172.30.0.3"}
			default:
				return nil
			}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{model.TopologyDefaultNetworkKey}},
			{Key: "db", Image: "ctf/db:v1", NetworkKeys: []string{model.TopologyDefaultNetworkKey}},
		},
		Policies: []model.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "db", Action: model.TopologyPolicyActionAllow, Protocol: model.TopologyPolicyProtocolTCP, Ports: []int{3306}},
		},
	})
	if err != nil {
		t.Fatalf("CreateTopology() error = %v", err)
	}
	if len(engine.appliedACLRules) != 2 {
		t.Fatalf("expected 2 acl rules, got %+v", engine.appliedACLRules)
	}
	if len(result.RuntimeDetails.ACLRules) != 2 {
		t.Fatalf("expected runtime acl rules, got %+v", result.RuntimeDetails.ACLRules)
	}
	if engine.appliedACLRules[0].Action != model.TopologyPolicyActionAllow || engine.appliedACLRules[0].Protocol != model.TopologyPolicyProtocolTCP {
		t.Fatalf("unexpected allow acl rule: %+v", engine.appliedACLRules[0])
	}
	if engine.appliedACLRules[1].Action != model.TopologyPolicyActionDeny || len(engine.appliedACLRules[1].Ports) != 0 {
		t.Fatalf("unexpected fallback deny rule: %+v", engine.appliedACLRules[1])
	}
}

func TestServiceCreateTopologyRollsBackWhenACLApplyFails(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-rollback",
		containerIDs: []string{"web-ctr", "db-ctr"},
		applyACLErr:  errors.New("acl apply failed"),
		inspectContainerNetworkIPsFunc: func(containerID string, engine *fakeRuntimeEngine) map[string]string {
			if len(engine.createdNetworkNames) == 0 {
				return nil
			}
			switch containerID {
			case "web-ctr":
				return map[string]string{engine.createdNetworkNames[0]: "172.31.0.2"}
			case "db-ctr":
				return map[string]string{engine.createdNetworkNames[0]: "172.31.0.3"}
			default:
				return nil
			}
		},
	}
	service := runtimecmd.NewProvisioningService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	_, err := service.CreateTopology(context.Background(), &runtimeports.TopologyCreateRequest{
		Networks: []runtimeports.TopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey},
		},
		Nodes: []runtimeports.TopologyCreateNode{
			{Key: "web", Image: "ctf/web:v1", ServicePort: 8080, IsEntryPoint: true, NetworkKeys: []string{model.TopologyDefaultNetworkKey}},
			{Key: "db", Image: "ctf/db:v1", NetworkKeys: []string{model.TopologyDefaultNetworkKey}},
		},
		Policies: []model.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "db", Action: model.TopologyPolicyActionAllow, Protocol: model.TopologyPolicyProtocolTCP, Ports: []int{3306}},
		},
	})
	if err == nil {
		t.Fatal("expected CreateTopology() to fail")
	}
	if len(engine.removedContainerIDs) != 2 {
		t.Fatalf("expected created containers to be cleaned up, got %v", engine.removedContainerIDs)
	}
	if len(engine.removedNetworkIDs) != 1 || engine.removedNetworkIDs[0] != "net-rollback" {
		t.Fatalf("expected created network to be cleaned up, got %v", engine.removedNetworkIDs)
	}
}

func TestServiceListTeacherInstancesScopesTeacherAndAppliesFilters(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	seedUser(t, repo.db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &model.User{ID: 2, Username: "alice", StudentNo: "S-1001", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &model.User{ID: 3, Username: "bob", StudentNo: "S-1002", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedChallenge(t, repo.db, &model.Challenge{ID: 11, Title: "web-101", Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &model.Instance{ID: 101, UserID: 2, ChallengeID: 11, ContainerID: "inst-a", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &model.Instance{ID: 102, UserID: 3, ChallengeID: 11, ContainerID: "inst-b", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &model.Instance{ID: 103, UserID: 2, ChallengeID: 11, ContainerID: "inst-stopped", Status: model.InstanceStatusStopped, ExpiresAt: now.Add(30 * time.Minute), CreatedAt: now, UpdatedAt: now})

	items, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, nil)
	if err != nil {
		t.Fatalf("ListTeacherInstances() error = %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 visible instance, got %d (%+v)", len(items), items)
	}
	if items[0].StudentUsername != "alice" || items[0].ClassName != "Class A" {
		t.Fatalf("unexpected item: %+v", items[0])
	}

	filtered, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, &dto.TeacherInstanceQuery{
		Keyword:   "ali",
		StudentNo: "S-1001",
	})
	if err != nil {
		t.Fatalf("ListTeacherInstances() with filters error = %v", err)
	}
	if len(filtered) != 1 || filtered[0].ID != 101 {
		t.Fatalf("unexpected filtered result: %+v", filtered)
	}

	byStudentNoKeyword, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, &dto.TeacherInstanceQuery{
		Keyword: "1001",
	})
	if err != nil {
		t.Fatalf("ListTeacherInstances() with student_no keyword error = %v", err)
	}
	if len(byStudentNoKeyword) != 1 || byStudentNoKeyword[0].ID != 101 {
		t.Fatalf("expected keyword to match student_no, got %+v", byStudentNoKeyword)
	}
}

func TestServiceListTeacherInstancesRejectsTeacherCrossClassFilter(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	seedUser(t, repo.db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})

	_, err := service.ListTeacherInstances(context.Background(), 1, model.RoleTeacher, &dto.TeacherInstanceQuery{ClassName: "Class B"})
	if err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden, got %v", err)
	}
}

func TestServiceDestroyTeacherInstanceHonorsClassScope(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestRuntimeModule(repo, nil)
	now := time.Now()

	seedUser(t, repo.db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &model.User{ID: 2, Username: "alice", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &model.User{ID: 3, Username: "bob", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedChallenge(t, repo.db, &model.Challenge{ID: 11, Title: "web-101", Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &model.Instance{ID: 201, UserID: 2, ChallengeID: 11, Status: model.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &model.Instance{ID: 202, UserID: 3, ChallengeID: 11, Status: model.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})

	if err := service.DestroyTeacherInstance(context.Background(), 202, 1, model.RoleTeacher); err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden destroy, got %v", err)
	}

	if err := service.DestroyTeacherInstance(context.Background(), 201, 1, model.RoleTeacher); err != nil {
		t.Fatalf("DestroyTeacherInstance() error = %v", err)
	}

	instance, err := repo.FindByID(context.Background(), 201)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if instance.Status != model.InstanceStatusStopped {
		t.Fatalf("expected stopped status, got %s", instance.Status)
	}
}

type runtimeTestRepository struct {
	*runtimeinfra.Repository
	db *gorm.DB
}

func newTestRepository(t *testing.T) *runtimeTestRepository {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Challenge{}, &model.Instance{}, &model.PortAllocation{}, &model.ContestRegistration{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	if err := db.AutoMigrate(&model.Team{}, &model.TeamMember{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	if err := db.AutoMigrate(&model.Contest{}, &model.ContestAWDService{}); err != nil {
		t.Fatalf("migrate awd tables: %v", err)
	}
	return &runtimeTestRepository{
		Repository: runtimeinfra.NewRepository(db),
		db:         db,
	}
}

type testRuntimeService struct {
	commands *runtimecmd.InstanceService
	queries  *runtimeqry.InstanceService
}

func (s *testRuntimeService) DestroyInstance(ctx context.Context, instanceID, userID int64) error {
	return s.commands.DestroyInstance(ctx, instanceID, userID)
}

func (s *testRuntimeService) ExtendInstance(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error) {
	return s.commands.ExtendInstance(ctx, instanceID, userID)
}

func (s *testRuntimeService) GetUserInstances(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error) {
	return s.queries.GetUserInstances(ctx, userID)
}

func (s *testRuntimeService) GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error) {
	return s.queries.GetAccessURL(ctx, instanceID, userID)
}

func (s *testRuntimeService) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error) {
	return s.queries.ListTeacherInstances(ctx, requesterID, requesterRole, query)
}

func (s *testRuntimeService) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	return s.commands.DestroyTeacherInstance(ctx, instanceID, requesterID, requesterRole)
}

func newTestRuntimeModule(repo *runtimeTestRepository, engine *fakeRuntimeEngine) *testRuntimeService {
	cfg := &config.ContainerConfig{
		MaxExtends:        2,
		ExtendDuration:    30 * time.Minute,
		OrphanGracePeriod: 5 * time.Minute,
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, nil)
	return &testRuntimeService{
		commands: runtimecmd.NewInstanceService(repo, cleanupService, cfg, nil),
		queries:  runtimeqry.NewInstanceService(repo),
	}
}

type fakeRuntimeEngine struct {
	networkID                      string
	networkIDs                     []string
	containerID                    string
	containerIDs                   []string
	startErr                       error
	applyACLErr                    error
	removeContainerErr             error
	removeNetworkErr               error
	resolvedServicePort            int
	resolveServicePortErr          error
	createdNetworkName             string
	createdNetworkNames            []string
	createdNetworkLabel            map[string]string
	createdContainerCfg            *model.ContainerConfig
	createdContainerCfgs           []*model.ContainerConfig
	removedContainerID             string
	removedContainerIDs            []string
	removedNetworkID               string
	removedNetworkIDs              []string
	appliedACLRules                []model.InstanceRuntimeACLRule
	removedACLRules                []model.InstanceRuntimeACLRule
	connectedNetworks              map[string][]string
	writtenFiles                   map[string]map[string]string
	imageSize                      int64
	imageInspectErr                error
	removedImageRef                string
	managedContainerStats          []runtimeports.ManagedContainerStat
	inspectContainerNetworkIPsFunc func(containerID string, engine *fakeRuntimeEngine) map[string]string
	stopContainerFn                func(ctx context.Context, containerID string, timeout time.Duration) error
	removeContainerFn              func(ctx context.Context, containerID string, force bool) error
	removeNetworkFn                func(ctx context.Context, networkID string) error
	removeACLRulesFn               func(ctx context.Context, rules []model.InstanceRuntimeACLRule) error
}

func (f *fakeRuntimeEngine) CreateNetwork(_ context.Context, name string, labels map[string]string, _ bool) (string, error) {
	f.createdNetworkName = name
	f.createdNetworkNames = append(f.createdNetworkNames, name)
	f.createdNetworkLabel = labels
	if len(f.networkIDs) > 0 {
		networkID := f.networkIDs[0]
		f.networkIDs = f.networkIDs[1:]
		return networkID, nil
	}
	return f.networkID, nil
}

func (f *fakeRuntimeEngine) CreateContainer(_ context.Context, cfg *model.ContainerConfig) (string, error) {
	f.createdContainerCfg = cfg
	f.createdContainerCfgs = append(f.createdContainerCfgs, cfg)
	if len(f.containerIDs) > 0 {
		containerID := f.containerIDs[0]
		f.containerIDs = f.containerIDs[1:]
		return containerID, nil
	}
	return f.containerID, nil
}

func (f *fakeRuntimeEngine) ResolveServicePort(_ context.Context, _ string, preferredPort int) (int, error) {
	if f.resolveServicePortErr != nil {
		return 0, f.resolveServicePortErr
	}
	if f.resolvedServicePort > 0 {
		return f.resolvedServicePort, nil
	}
	return preferredPort, nil
}

func (f *fakeRuntimeEngine) InspectImageSize(_ context.Context, _ string) (int64, error) {
	if f.imageInspectErr != nil {
		return 0, f.imageInspectErr
	}
	return f.imageSize, nil
}

func (f *fakeRuntimeEngine) RemoveImage(_ context.Context, imageRef string) error {
	f.removedImageRef = imageRef
	return nil
}

func (f *fakeRuntimeEngine) ListManagedContainerStats(_ context.Context) ([]runtimeports.ManagedContainerStat, error) {
	return append([]runtimeports.ManagedContainerStat(nil), f.managedContainerStats...), nil
}

func (f *fakeRuntimeEngine) ConnectContainerToNetwork(_ context.Context, containerID, networkName string) error {
	if f.connectedNetworks == nil {
		f.connectedNetworks = make(map[string][]string)
	}
	f.connectedNetworks[containerID] = append(f.connectedNetworks[containerID], networkName)
	return nil
}

func (f *fakeRuntimeEngine) InspectContainerNetworkIPs(_ context.Context, containerID string) (map[string]string, error) {
	if f.inspectContainerNetworkIPsFunc == nil {
		return nil, nil
	}
	return f.inspectContainerNetworkIPsFunc(containerID, f), nil
}

func (f *fakeRuntimeEngine) StartContainer(_ context.Context, _ string) error {
	return f.startErr
}

func (f *fakeRuntimeEngine) StopContainer(ctx context.Context, containerID string, timeout time.Duration) error {
	if f.stopContainerFn != nil {
		return f.stopContainerFn(ctx, containerID, timeout)
	}
	return nil
}

func (f *fakeRuntimeEngine) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	f.removedContainerID = containerID
	f.removedContainerIDs = append(f.removedContainerIDs, containerID)
	if f.removeContainerFn != nil {
		return f.removeContainerFn(ctx, containerID, force)
	}
	return f.removeContainerErr
}

func (f *fakeRuntimeEngine) RemoveNetwork(ctx context.Context, networkID string) error {
	f.removedNetworkID = networkID
	f.removedNetworkIDs = append(f.removedNetworkIDs, networkID)
	if f.removeNetworkFn != nil {
		return f.removeNetworkFn(ctx, networkID)
	}
	return f.removeNetworkErr
}

func (f *fakeRuntimeEngine) ApplyACLRules(_ context.Context, rules []model.InstanceRuntimeACLRule) error {
	if f.applyACLErr != nil {
		return f.applyACLErr
	}
	f.appliedACLRules = append(f.appliedACLRules, rules...)
	return nil
}

func (f *fakeRuntimeEngine) RemoveACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error {
	if f.removeACLRulesFn != nil {
		return f.removeACLRulesFn(ctx, rules)
	}
	f.removedACLRules = append(f.removedACLRules, rules...)
	return nil
}

func (f *fakeRuntimeEngine) WriteFileToContainer(_ context.Context, containerID, filePath string, content []byte) error {
	if f.writtenFiles == nil {
		f.writtenFiles = make(map[string]map[string]string)
	}
	if f.writtenFiles[containerID] == nil {
		f.writtenFiles[containerID] = make(map[string]string)
	}
	f.writtenFiles[containerID][filePath] = string(content)
	return nil
}

func (f *fakeRuntimeEngine) ListManagedContainers(_ context.Context) ([]runtimeports.ManagedContainer, error) {
	return nil, nil
}

func seedInstance(t *testing.T, db *gorm.DB, instance *model.Instance) {
	t.Helper()

	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
}

func seedUser(t *testing.T, db *gorm.DB, user *model.User) {
	t.Helper()

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func seedChallenge(t *testing.T, db *gorm.DB, challenge *model.Challenge) {
	t.Helper()

	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
}
