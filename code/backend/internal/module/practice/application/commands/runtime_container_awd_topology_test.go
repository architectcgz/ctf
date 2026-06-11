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
	"testing"
	"time"
)

func TestCreateTopologyAWDContainerUsesStableContestNetwork(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	const checkerSecret = "practice-topology-secret-1234567890123"
	if err := db.Create(&practiceCommandImageRow{
		ID:        503,
		Name:      "ctf/awd-topology",
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
		Name: "AWD Topology",
		RuntimeConfig: map[string]any{
			"image_id":          503,
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
					AWDChallengeID:  503,
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
					if req.ContainerName != "ctf-instance-challenge-c7003-t7103-s8003" {
						t.Fatalf("expected awd container name, got %q", req.ContainerName)
					}
					if !req.DisableEntryPortPublishing {
						t.Fatal("expected entry port publishing to be disabled")
					}
					if len(req.Networks) != 1 || req.Networks[0].Name != "ctf-awd-contest-7003" || !req.Networks[0].Shared {
						t.Fatalf("expected stable shared AWD contest network, got %+v", req.Networks)
					}
					if len(req.Nodes) != 1 || req.Nodes[0].Key != "web" || !req.Nodes[0].IsEntryPoint {
						t.Fatalf("unexpected topology nodes: %+v", req.Nodes)
					}
					if req.Nodes[0].Env["CHECKER_TOKEN"] != contestdomain.BuildAWDCheckerToken(contestID, teamID, serviceID, 503, checkerSecret) {
						t.Fatalf("unexpected checker token env: %+v", req.Nodes[0].Env)
					}
					if len(req.Nodes[0].NetworkAliases) != 1 || req.Nodes[0].NetworkAliases[0] != "awd-c7003-t7103-s8003" {
						t.Fatalf("expected stable AWD service alias, got %+v", req.Nodes[0].NetworkAliases)
					}
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "awd-topology-ctr",
						NetworkID:          "net-awd-contest-7003",
						AccessURL:          "http://awd-c7003-t7103-s8003:8080",
						RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
							Networks: []runtimecontracts.InstanceRuntimeNetwork{
								{Key: challengecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-7003", NetworkID: "net-awd-contest-7003", Shared: true},
							},
							Containers: []runtimecontracts.InstanceRuntimeContainer{
								{NodeKey: "web", ContainerID: "awd-topology-ctr", ServicePort: 8080, IsEntryPoint: true, NetworkAliases: []string{"awd-c7003-t7103-s8003"}},
							},
						},
					}, nil
				case 2:
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "workspace-ctr",
						NetworkID:          "net-awd-contest-7003",
						AccessURL:          "tcp://172.30.0.21:22",
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
		},
		config: &config.Config{
			Container: config.ContainerConfig{FlagGlobalSecret: checkerSecret},
		},
	}
	instance := &instanceentity.Instance{
		ID:          9003,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 503,
	}
	challenge := &challengecontracts.PracticeRuntimeChallenge{
		ID:       503,
		ImageID:  503,
		FlagType: challengecontracts.FlagTypeStatic,
	}
	topology, err := challengecontracts.EncodeTopologySpec(challengecontracts.TopologySpec{
		Nodes: []challengecontracts.TopologyNode{
			{Key: "web", ServicePort: 8080, InjectFlag: true},
		},
	})
	if err != nil {
		t.Fatalf("encode topology: %v", err)
	}

	if err := service.createContainer(context.Background(), instance, toPracticeChallenge(challenge), &practiceports.RuntimeChallengeTopology{
		ChallengeID:  503,
		EntryNodeKey: "web",
		Spec:         topology,
	}, "flag{demo}"); err != nil {
		t.Fatalf("createContainer() error = %v", err)
	}
	if createTopologyCalls != 2 {
		t.Fatalf("expected runtime and workspace topology creation, got %d calls", createTopologyCalls)
	}
	if instance.AccessURL != "http://awd-c7003-t7103-s8003:8080" {
		t.Fatalf("unexpected access url: %s", instance.AccessURL)
	}
}

func TestCreateTopologyAWDContainerUsesPublishedAccessHostWhenConfigured(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	const checkerSecret = "practice-topology-secret-1234567890123"
	if err := db.Create(&practiceCommandImageRow{
		ID:        513,
		Name:      "ctf/awd-topology",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(7013)
	teamID := int64(7113)
	serviceID := int64(8013)
	serviceSnapshot, err := contestentity.EncodeContestAWDServiceSnapshot(contestentity.ContestAWDServiceSnapshot{
		Name: "AWD Topology",
		RuntimeConfig: map[string]any{
			"image_id":          513,
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
					AWDChallengeID:  513,
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
					if req.ReservedHostPort != 30013 {
						t.Fatalf("expected reserved host port, got %d", req.ReservedHostPort)
					}
					if req.DisableEntryPortPublishing {
						t.Fatal("expected topology awd entry port publishing when access host is configured")
					}
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "awd-topology-ctr",
						NetworkID:          "net-awd-contest-7013",
						AccessURL:          "http://host-gateway.internal:30013",
						RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
							Networks: []runtimecontracts.InstanceRuntimeNetwork{
								{Key: challengecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-7013", NetworkID: "net-awd-contest-7013", Shared: true},
							},
							Containers: []runtimecontracts.InstanceRuntimeContainer{
								{NodeKey: "web", ContainerID: "awd-topology-ctr", ServicePort: 8080, HostPort: 30013, IsEntryPoint: true, NetworkAliases: []string{"awd-c7013-t7113-s8013"}},
							},
						},
					}, nil
				case 2:
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "workspace-ctr",
						NetworkID:          "net-awd-contest-7013",
						AccessURL:          "tcp://172.30.0.21:22",
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
		},
		config: &config.Config{
			Container: config.ContainerConfig{
				PublicHost:       "127.0.0.1",
				AccessHost:       "host-gateway.internal",
				FlagGlobalSecret: checkerSecret,
			},
		},
	}
	instance := &instanceentity.Instance{
		ID:          9013,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 513,
		HostPort:    30013,
	}
	challenge := &challengecontracts.PracticeRuntimeChallenge{
		ID:       513,
		ImageID:  513,
		FlagType: challengecontracts.FlagTypeStatic,
	}
	topology, err := challengecontracts.EncodeTopologySpec(challengecontracts.TopologySpec{
		Nodes: []challengecontracts.TopologyNode{
			{Key: "web", ServicePort: 8080, InjectFlag: true},
		},
	})
	if err != nil {
		t.Fatalf("encode topology: %v", err)
	}

	if err := service.createContainer(context.Background(), instance, toPracticeChallenge(challenge), &practiceports.RuntimeChallengeTopology{
		ChallengeID:  513,
		EntryNodeKey: "web",
		Spec:         topology,
	}, "flag{demo}"); err != nil {
		t.Fatalf("createContainer() error = %v", err)
	}
	if createTopologyCalls != 2 {
		t.Fatalf("expected runtime and workspace topology creation, got %d calls", createTopologyCalls)
	}
	if instance.AccessURL != "http://host-gateway.internal:30013" {
		t.Fatalf("unexpected access url: %s", instance.AccessURL)
	}
	if instance.HostPort != 30013 {
		t.Fatalf("expected host port to stay reserved, got %+v", instance)
	}
}
