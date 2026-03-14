package container

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

func TestSelectOrphanContainersSkipsActiveAndGracePeriod(t *testing.T) {
	t.Parallel()

	now := time.Now()
	managedContainers := []ManagedContainer{
		{ID: "active", Name: "ctf-instance-active", CreatedAt: now.Add(-10 * time.Minute)},
		{ID: "fresh", Name: "ctf-instance-fresh", CreatedAt: now.Add(-2 * time.Minute)},
		{ID: "orphan", Name: "ctf-instance-orphan", CreatedAt: now.Add(-12 * time.Minute)},
	}
	activeContainerIDs := map[string]struct{}{
		"active": {},
	}

	orphanContainers := selectOrphanContainers(managedContainers, activeContainerIDs, 5*time.Minute, now)
	if len(orphanContainers) != 1 {
		t.Fatalf("expected 1 orphan container, got %d (%v)", len(orphanContainers), orphanContainers)
	}
	if orphanContainers[0].ID != "orphan" {
		t.Fatalf("unexpected orphan container: %+v", orphanContainers[0])
	}
}

func TestManagedContainerLabels(t *testing.T) {
	t.Parallel()

	labels := managedContainerLabels()
	if labels[managedByLabelKey] != managedByLabelValue {
		t.Fatalf("expected managed-by label, got %v", labels)
	}
	if labels[challengeInstanceLabelKey] != challengeInstanceLabelValue {
		t.Fatalf("expected component label, got %v", labels)
	}
}

func TestManagedNetworkLabels(t *testing.T) {
	t.Parallel()

	labels := managedNetworkLabels()
	if labels[managedByLabelKey] != managedByLabelValue {
		t.Fatalf("expected managed-by label, got %v", labels)
	}
	if labels[challengeInstanceLabelKey] != challengeInstanceLabelValue {
		t.Fatalf("expected component label, got %v", labels)
	}
}

func TestNewServiceTreatsTypedNilEngineAsNil(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	cfg := &config.ContainerConfig{
		PortRangeStart:       30000,
		PortRangeEnd:         30010,
		DefaultExposedPort:   8080,
		PublicHost:           "127.0.0.1",
		DefaultTTL:           time.Hour,
		MaxExtends:           2,
		MaxConcurrentPerUser: 3,
		CreateTimeout:        time.Second,
	}

	var typedNil *Engine
	service := NewService(repo, typedNil, cfg, nil)
	if service.engine != nil {
		t.Fatalf("expected typed nil engine to be normalized to nil, got %#v", service.engine)
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
	service := NewService(repo, engine, &config.ContainerConfig{
		PortRangeStart:     30000,
		PortRangeEnd:       30010,
		DefaultExposedPort: 8080,
	}, nil)

	containerID, networkID, hostPort, servicePort, err := service.CreateContainer(context.Background(), "ctf/web:v1", map[string]string{"FLAG": "flag{1}"})
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

func TestServiceCreateContainerRemovesNetworkWhenStartFails(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:   "net-456",
		containerID: "ctr-456",
		startErr:    errors.New("start failed"),
	}
	service := NewService(repo, engine, &config.ContainerConfig{
		PortRangeStart:     30000,
		PortRangeEnd:       30010,
		DefaultExposedPort: 8080,
	}, nil)

	_, _, _, _, err := service.CreateContainer(context.Background(), "ctf/web:v1", nil)
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

func TestServiceDestroyInstanceAllowsContestTeamMember(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestService(repo)
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
		ContainerID: "contest-shared",
		Status:      model.InstanceStatusRunning,
		ExpiresAt:   now.Add(time.Hour),
	})

	if err := service.DestroyInstance(901, 2); err != nil {
		t.Fatalf("DestroyInstance() error = %v", err)
	}

	instance, err := repo.FindByID(901)
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
	service := newTestService(repo)
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

	resp, err := service.ExtendInstance(902, 2)
	if err != nil {
		t.Fatalf("ExtendInstance() error = %v", err)
	}
	if resp == nil {
		t.Fatal("expected extend response")
	}
	if resp.RemainingExtends != 1 {
		t.Fatalf("expected remaining extends 1, got %d", resp.RemainingExtends)
	}

	instance, err := repo.FindByID(902)
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

func TestServiceCreateTopologyCreatesMultipleContainersOnSharedNetwork(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkID:    "net-789",
		containerIDs: []string{"web-ctr", "db-ctr"},
	}
	service := NewService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &TopologyCreateRequest{
		Networks: []TopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey},
		},
		Nodes: []TopologyCreateNode{
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
	service := NewService(repo, engine, &config.ContainerConfig{
		MaxExtends:        2,
		ExtendDuration:    30 * time.Minute,
		OrphanGracePeriod: 5 * time.Minute,
	}, nil)

	instance := &model.Instance{
		ID:             1,
		UserID:         1,
		ChallengeID:    1,
		ContainerID:    "web-ctr",
		NetworkID:      "net-1",
		RuntimeDetails: `{"containers":[{"container_id":"web-ctr"},{"container_id":"db-ctr"}],"acl_rules":[{"comment":"ctf:acl:test","source_ip":"172.30.0.2","target_ip":"172.30.0.3","action":"allow","protocol":"tcp","ports":[3306]}]}`,
		Status:         model.InstanceStatusRunning,
		ExpiresAt:      time.Now().Add(time.Hour),
	}
	seedInstance(t, repo.db, instance)

	if err := service.destroyManagedInstance(instance); err != nil {
		t.Fatalf("destroyManagedInstance() error = %v", err)
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
}

func TestServiceCreateTopologyCreatesAndConnectsMultipleNetworks(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	engine := &fakeRuntimeEngine{
		networkIDs:   []string{"net-public", "net-backend"},
		containerIDs: []string{"web-ctr", "db-ctr"},
	}
	service := NewService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &TopologyCreateRequest{
		Networks: []TopologyCreateNetwork{
			{Key: "public"},
			{Key: "backend", Internal: true},
		},
		Nodes: []TopologyCreateNode{
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
	service := NewService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	result, err := service.CreateTopology(context.Background(), &TopologyCreateRequest{
		Networks: []TopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey},
		},
		Nodes: []TopologyCreateNode{
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
	service := NewService(repo, engine, &config.ContainerConfig{
		PortRangeStart: 30000,
		PortRangeEnd:   30010,
		PublicHost:     "127.0.0.1",
	}, nil)

	_, err := service.CreateTopology(context.Background(), &TopologyCreateRequest{
		Networks: []TopologyCreateNetwork{
			{Key: model.TopologyDefaultNetworkKey},
		},
		Nodes: []TopologyCreateNode{
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
	service := newTestService(repo)
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
}

func TestServiceListTeacherInstancesRejectsTeacherCrossClassFilter(t *testing.T) {
	t.Parallel()

	repo := newTestRepository(t)
	service := newTestService(repo)
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
	service := newTestService(repo)
	now := time.Now()

	seedUser(t, repo.db, &model.User{ID: 1, Username: "teacher-a", Role: model.RoleTeacher, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &model.User{ID: 2, Username: "alice", Role: model.RoleStudent, ClassName: "Class A", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedUser(t, repo.db, &model.User{ID: 3, Username: "bob", Role: model.RoleStudent, ClassName: "Class B", Status: model.UserStatusActive, CreatedAt: now, UpdatedAt: now})
	seedChallenge(t, repo.db, &model.Challenge{ID: 11, Title: "web-101", Status: model.ChallengeStatusPublished, CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &model.Instance{ID: 201, UserID: 2, ChallengeID: 11, ContainerID: "inst-a", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})
	seedInstance(t, repo.db, &model.Instance{ID: 202, UserID: 3, ChallengeID: 11, ContainerID: "inst-b", Status: model.InstanceStatusRunning, ExpiresAt: now.Add(time.Hour), CreatedAt: now, UpdatedAt: now})

	if err := service.DestroyTeacherInstance(context.Background(), 202, 1, model.RoleTeacher); err == nil || err.Error() != errcode.ErrForbidden.Error() {
		t.Fatalf("expected forbidden destroy, got %v", err)
	}

	if err := service.DestroyTeacherInstance(context.Background(), 201, 1, model.RoleTeacher); err != nil {
		t.Fatalf("DestroyTeacherInstance() error = %v", err)
	}

	instance, err := repo.FindByID(201)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if instance.Status != model.InstanceStatusStopped {
		t.Fatalf("expected stopped status, got %s", instance.Status)
	}
}

func newTestRepository(t *testing.T) *Repository {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.Challenge{}, &model.Instance{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	if err := db.AutoMigrate(&model.Team{}, &model.TeamMember{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	return NewRepository(db)
}

func newTestService(repo *Repository) *Service {
	return NewService(repo, nil, &config.ContainerConfig{
		MaxExtends:        2,
		ExtendDuration:    30 * time.Minute,
		OrphanGracePeriod: 5 * time.Minute,
	}, nil)
}

type fakeRuntimeEngine struct {
	networkID                      string
	networkIDs                     []string
	containerID                    string
	containerIDs                   []string
	startErr                       error
	applyACLErr                    error
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
	inspectContainerNetworkIPsFunc func(containerID string, engine *fakeRuntimeEngine) map[string]string
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

func (f *fakeRuntimeEngine) StopContainer(_ context.Context, _ string, _ time.Duration) error {
	return nil
}

func (f *fakeRuntimeEngine) RemoveContainer(_ context.Context, containerID string, _ bool) error {
	f.removedContainerID = containerID
	f.removedContainerIDs = append(f.removedContainerIDs, containerID)
	return nil
}

func (f *fakeRuntimeEngine) RemoveNetwork(_ context.Context, networkID string) error {
	f.removedNetworkID = networkID
	f.removedNetworkIDs = append(f.removedNetworkIDs, networkID)
	return nil
}

func (f *fakeRuntimeEngine) ApplyACLRules(_ context.Context, rules []model.InstanceRuntimeACLRule) error {
	if f.applyACLErr != nil {
		return f.applyACLErr
	}
	f.appliedACLRules = append(f.appliedACLRules, rules...)
	return nil
}

func (f *fakeRuntimeEngine) RemoveACLRules(_ context.Context, rules []model.InstanceRuntimeACLRule) error {
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

func (f *fakeRuntimeEngine) ListManagedContainers(_ context.Context, _ string) ([]ManagedContainer, error) {
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
