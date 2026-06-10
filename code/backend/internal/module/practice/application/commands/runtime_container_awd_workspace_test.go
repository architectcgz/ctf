package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/internal/module/practice/testsupport/contestentity"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	"testing"
	"time"
)

func TestCreateSingleAWDContainerCreatesWorkspaceCompanionWithSharedMounts(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{
		ID:        601,
		Name:      "ctf/awd-workspace",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(7601)
	teamID := int64(7701)
	serviceID := int64(7801)
	serviceSnapshot, err := contestentity.EncodeContestAWDServiceSnapshot(contestentity.ContestAWDServiceSnapshot{
		Name: "Campus Drive",
		RuntimeConfig: map[string]any{
			"image_id":         601,
			"instance_sharing": string(challengecontracts.InstanceSharingPerTeam),
			"defense_workspace": map[string]any{
				"entry_mode":      "ssh",
				"seed_root":       "docker/workspace",
				"workspace_roots": []string{"docker/workspace/src", "docker/workspace/data"},
				"writable_roots":  []string{"docker/workspace/src"},
				"readonly_roots":  []string{"docker/workspace/data"},
				"runtime_mounts": []map[string]any{
					{"source": "docker/workspace/src", "target": "/workspace/src", "mode": "rw"},
					{"source": "docker/workspace/data", "target": "/workspace/data", "mode": "ro"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("encode service snapshot: %v", err)
	}

	repo := &stubPracticeRepository{
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
			if gotContestID != contestID || gotServiceID != serviceID {
				t.Fatalf("unexpected awd service lookup: contest=%d service=%d", gotContestID, gotServiceID)
			}
			return &practiceports.ContestAWDServiceRecord{
				ID:              serviceID,
				ContestID:       contestID,
				DisplayName:     "Campus Drive",
				AWDChallengeID:  8601,
				IsVisible:       true,
				ServiceSnapshot: serviceSnapshot,
			}, nil
		},
	}

	var requests []*practiceports.TopologyCreateRequest
	service := &Service{
		repo:         repo,
		imageRepo:    challengeinfra.NewImageRepository(db),
		instanceRepo: newPracticeTestInstanceRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				requests = append(requests, req)
				switch len(requests) {
				case 1:
					if len(req.Nodes) != 1 || req.Nodes[0].Image != "ctf/awd-workspace:v1" {
						t.Fatalf("unexpected runtime topology request: %+v", req)
					}
					if len(req.Nodes[0].Mounts) != 2 {
						t.Fatalf("expected runtime mounts, got %+v", req.Nodes[0].Mounts)
					}
					if req.Nodes[0].Mounts[0].Target != "/workspace/src" || req.Nodes[0].Mounts[1].Target != "/workspace/data" {
						t.Fatalf("unexpected runtime mount targets: %+v", req.Nodes[0].Mounts)
					}
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "runtime-ctr",
						NetworkID:          "net-awd-contest-7601",
						AccessURL:          "http://awd-c7601-t7701-s7801:8080",
						RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
							Networks: []runtimecontracts.InstanceRuntimeNetwork{
								{Key: challengecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-7601", NetworkID: "net-awd-contest-7601", Shared: true},
							},
							Containers: []runtimecontracts.InstanceRuntimeContainer{
								{NodeKey: "default", ContainerID: "runtime-ctr", ServicePort: 8080, IsEntryPoint: true, NetworkAliases: []string{"awd-c7601-t7701-s7801"}},
							},
						},
					}, nil
				case 2:
					if len(req.Nodes) != 1 || req.Nodes[0].WorkingDir != "/workspace" {
						t.Fatalf("unexpected workspace topology request: %+v", req)
					}
					assertAWDDefenseWorkspaceShellNode(t, req.Nodes[0])
					if len(req.Nodes[0].Mounts) != 2 {
						t.Fatalf("expected workspace mounts, got %+v", req.Nodes[0].Mounts)
					}
					if req.Nodes[0].Mounts[0].Source != requests[0].Nodes[0].Mounts[0].Source || req.Nodes[0].Mounts[1].Source != requests[0].Nodes[0].Mounts[1].Source {
						t.Fatalf("expected shared workspace sources, runtime=%+v workspace=%+v", requests[0].Nodes[0].Mounts, req.Nodes[0].Mounts)
					}
					if req.Nodes[0].Mounts[0].Target != "/workspace/src" || req.Nodes[0].Mounts[1].Target != "/workspace/data" {
						t.Fatalf("unexpected workspace mount targets: %+v", req.Nodes[0].Mounts)
					}
					if req.Nodes[0].Mounts[0].ReadOnly {
						t.Fatalf("expected src root to stay writable, got %+v", req.Nodes[0].Mounts[0])
					}
					if !req.Nodes[0].Mounts[1].ReadOnly {
						t.Fatalf("expected data root to stay readonly, got %+v", req.Nodes[0].Mounts[1])
					}
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "workspace-ctr",
						NetworkID:          "net-awd-contest-7601",
						AccessURL:          "tcp://172.30.0.40:22",
						RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
							Containers: []runtimecontracts.InstanceRuntimeContainer{
								{NodeKey: "workspace", ContainerID: "workspace-ctr", ServicePort: 22, ServiceProtocol: challengecontracts.ChallengeTargetProtocolTCP, IsEntryPoint: true},
							},
						},
					}, nil
				default:
					t.Fatalf("unexpected topology create call #%d", len(requests))
					return nil, nil
				}
			},
		},
		config: &config.Config{},
	}

	instance := &instanceentity.Instance{
		ID:          9001,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 601,
	}
	challenge := &challengecontracts.PracticeRuntimeChallenge{
		ID:          601,
		ImageID:     601,
		FlagType:    challengecontracts.FlagTypeStatic,
		PackageSlug: stringPtr("campus-drive"),
	}

	if err := service.createSingleContainer(context.Background(), instance, toPracticeChallenge(challenge), "flag{demo}"); err != nil {
		t.Fatalf("createSingleContainer() error = %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("expected runtime and workspace topology creation, got %d calls", len(requests))
	}
	if instance.ContainerID != "runtime-ctr" || instance.AccessURL != "http://awd-c7601-t7701-s7801:8080" {
		t.Fatalf("unexpected runtime instance after createSingleContainer(): %+v", instance)
	}

	workspace, err := runtimeinfrarepo.NewAWDRepository(db).FindAWDDefenseWorkspace(context.Background(), contestID, teamID, serviceID)
	if err != nil {
		t.Fatalf("FindAWDDefenseWorkspace() error = %v", err)
	}
	if workspace == nil {
		t.Fatal("expected workspace row to be created")
	}
	if workspace.WorkspaceRevision != 1 || workspace.Status != runtimeentity.AWDDefenseWorkspaceStatusRunning || workspace.ContainerID != "workspace-ctr" {
		t.Fatalf("unexpected workspace state: %+v", workspace)
	}
}
