package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	contestdomain "ctf-platform/internal/module/contest/domain"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/internal/module/practice/testsupport/contestentity"
	"errors"
	"testing"
	"time"
)

func TestCreateSingleAWDContainerUsesPrivateTopology(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	const checkerSecret = "practice-secret-12345678901234567890"
	if err := db.Create(&practiceCommandImageRow{
		ID:        501,
		Name:      "ctf/awd-web",
		Tag:       "v1",
		Digest:    "sha256:awd-web-v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(7001)
	teamID := int64(7101)
	serviceID := int64(8001)
	serviceSnapshot, err := contestentity.EncodeContestAWDServiceSnapshot(contestentity.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":          501,
			"instance_sharing":  string(challengecontracts.InstanceSharingPerTeam),
			"checker_token_env": "CHECKER_TOKEN",
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
	createTopologyCalls := 0
	service := &serviceCore{
		repo: &stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
				return &practiceports.ContestAWDServiceRecord{
					ID:              gotServiceID,
					ContestID:       gotContestID,
					AWDChallengeID:  501,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
		},
		imageRepo:    challengeinfra.NewImageRepository(db),
		instanceRepo: newPracticeTestInstanceRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				createTopologyCalls++
				switch createTopologyCalls {
				case 1:
					if req.ReservedHostPort != 0 {
						t.Fatalf("expected no reserved host port, got %d", req.ReservedHostPort)
					}
					if req.ContainerName != "ctf-instance-challenge-c7001-t7101-s8001" {
						t.Fatalf("expected awd container name, got %q", req.ContainerName)
					}
					if !req.DisableEntryPortPublishing {
						t.Fatal("expected entry port publishing to be disabled")
					}
					if len(req.Networks) != 1 || req.Networks[0].Name != "ctf-awd-contest-7001" || !req.Networks[0].Shared {
						t.Fatalf("expected stable shared AWD contest network, got %+v", req.Networks)
					}
					if len(req.Nodes) != 1 || !req.Nodes[0].IsEntryPoint || req.Nodes[0].Image != "ctf/awd-web@sha256:awd-web-v1" {
						t.Fatalf("unexpected runtime topology request: %+v", req)
					}
					if req.Nodes[0].Env["CHECKER_TOKEN"] != contestdomain.BuildAWDCheckerToken(contestID, teamID, serviceID, 501, checkerSecret) {
						t.Fatalf("unexpected checker token env: %+v", req.Nodes[0].Env)
					}
					if len(req.Nodes[0].NetworkAliases) != 1 || req.Nodes[0].NetworkAliases[0] != "awd-c7001-t7101-s8001" {
						t.Fatalf("expected stable AWD service alias, got %+v", req.Nodes[0].NetworkAliases)
					}
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "awd-private-ctr",
						NetworkID:          "net-awd-contest-7001",
						AccessURL:          "http://awd-c7001-t7101-s8001:8080",
						RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
							Networks: []runtimecontracts.InstanceRuntimeNetwork{
								{
									Key:       challengecontracts.TopologyDefaultNetworkKey,
									Name:      "ctf-awd-contest-7001",
									NetworkID: "net-awd-contest-7001",
									Shared:    true,
								},
							},
							Containers: []runtimecontracts.InstanceRuntimeContainer{
								{
									NodeKey:        "default",
									ContainerID:    "awd-private-ctr",
									ServicePort:    8080,
									IsEntryPoint:   true,
									NetworkAliases: []string{"awd-c7001-t7101-s8001"},
								},
							},
						},
					}, nil
				case 2:
					if len(req.Nodes) != 1 || len(req.Nodes[0].NetworkAliases) != 1 || req.Nodes[0].NetworkAliases[0] != "awd-ws-c7001-t7101-s8001-r1" {
						t.Fatalf("expected stable workspace alias, got %+v", req.Nodes)
					}
					assertAWDDefenseWorkspaceShellNode(t, req.Nodes[0])
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "workspace-ctr",
						NetworkID:          "net-awd-contest-7001",
						AccessURL:          "tcp://172.30.0.20:22",
						RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
							Containers: []runtimecontracts.InstanceRuntimeContainer{
								{NodeKey: "workspace", ContainerID: "workspace-ctr", ServicePort: 22, ServiceProtocol: challengecontracts.ChallengeTargetProtocolTCP, IsEntryPoint: true},
							},
						},
					}, nil
				default:
					t.Fatalf("unexpected topology create call #%d", createTopologyCalls)
					return nil, nil
				}
			},
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, _ int64) (string, string, int, int, error) {
				t.Fatal("AWD service instances must not use host-port CreateContainer")
				return "", "", 0, 0, nil
			},
		},
		config: &config.Config{
			Container: config.ContainerConfig{FlagGlobalSecret: checkerSecret},
		},
	}
	instance := &instanceentity.Instance{
		ID:          9001,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 501,
	}
	challenge := &challengecontracts.PracticeRuntimeChallenge{
		ID:       501,
		ImageID:  501,
		FlagType: challengecontracts.FlagTypeStatic,
	}

	if err := service.createSingleContainer(context.Background(), instance, toPracticeChallenge(challenge), "flag{demo}"); err != nil {
		t.Fatalf("createSingleContainer() error = %v cause=%v", err, errors.Unwrap(err))
	}
	if createTopologyCalls != 2 {
		t.Fatalf("expected runtime and workspace topology creation, got %d calls", createTopologyCalls)
	}
	if instance.HostPort != 0 {
		t.Fatalf("expected instance host port to remain empty, got %d", instance.HostPort)
	}
	if instance.AccessURL != "http://awd-c7001-t7101-s8001:8080" {
		t.Fatalf("unexpected access url: %s", instance.AccessURL)
	}
	details, err := runtimecontracts.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
	if err != nil {
		t.Fatalf("DecodeInstanceRuntimeDetails() error = %v", err)
	}
	if token := details.FindAWDCheckerToken("CHECKER_TOKEN"); token != contestdomain.BuildAWDCheckerToken(contestID, teamID, serviceID, 501, checkerSecret) {
		t.Fatalf("unexpected persisted checker token: %+v", details)
	}
}
