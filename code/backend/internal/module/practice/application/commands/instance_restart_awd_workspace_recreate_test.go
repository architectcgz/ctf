package commands

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
)

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
		newPracticeRepositoryWithRuntimePortOwner(db),

		challengeinfra.NewImageRepository(db),
		newPracticeTestInstanceRepository(db),
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

		newPracticeRepositoryWithRuntimePortOwner(db), challengeinfra.NewRepository(db))

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

	workspace, err := runtimeinfrarepo.NewAWDRepository(db).FindAWDDefenseWorkspace(context.Background(), contestID, teamID, serviceID)
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
