package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	runtimeentity "ctf-platform/internal/module/contest/entity"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceports "ctf-platform/internal/module/practice/ports"
	"ctf-platform/internal/module/practice/testsupport/contestentity"
	"errors"
	"testing"
	"time"
)

func TestCreateSingleAWDContainerUsesPublishedAccessHostWhenConfigured(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	const checkerSecret = "practice-secret-12345678901234567890"
	if err := db.Create(&practiceCommandImageRow{
		ID:        511,
		Name:      "ctf/awd-web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(7011)
	teamID := int64(7111)
	serviceID := int64(8011)
	serviceSnapshot, err := contestentity.EncodeContestAWDServiceSnapshot(contestentity.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":          511,
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
					AWDChallengeID:  511,
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
					if req.ReservedHostPort != 30011 {
						t.Fatalf("expected reserved host port, got %d", req.ReservedHostPort)
					}
					if req.DisableEntryPortPublishing {
						t.Fatal("expected awd entry port publishing when access host is configured")
					}
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "awd-private-ctr",
						NetworkID:          "net-awd-contest-7011",
						AccessURL:          "http://host-gateway.internal:30011",
						RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
							Networks: []runtimecontracts.InstanceRuntimeNetwork{
								{
									Key:       challengecontracts.TopologyDefaultNetworkKey,
									Name:      "ctf-awd-contest-7011",
									NetworkID: "net-awd-contest-7011",
									Shared:    true,
								},
							},
							Containers: []runtimecontracts.InstanceRuntimeContainer{
								{
									NodeKey:        "default",
									ContainerID:    "awd-private-ctr",
									ServicePort:    8080,
									HostPort:       30011,
									IsEntryPoint:   true,
									NetworkAliases: []string{"awd-c7011-t7111-s8011"},
								},
							},
						},
					}, nil
				case 2:
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "workspace-ctr",
						NetworkID:          "net-awd-contest-7011",
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
			Container: config.ContainerConfig{
				PublicHost:       "127.0.0.1",
				AccessHost:       "host-gateway.internal",
				FlagGlobalSecret: checkerSecret,
			},
		},
	}
	instance := &instanceentity.Instance{
		ID:          9011,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 511,
		HostPort:    30011,
	}
	challenge := &challengecontracts.PracticeRuntimeChallenge{
		ID:       511,
		ImageID:  511,
		FlagType: challengecontracts.FlagTypeStatic,
	}

	if err := service.createSingleContainer(context.Background(), instance, toPracticeChallenge(challenge), "flag{demo}"); err != nil {
		t.Fatalf(
			"createSingleContainer() error = %v cause=%v conflict=%t calls=%d host_port=%d",
			err,
			errors.Unwrap(err),
			errors.Is(err, runtimeports.ErrPublishedHostPortConflict),
			createTopologyCalls,
			instance.HostPort,
		)
	}
	if createTopologyCalls != 2 {
		t.Fatalf("expected runtime and workspace topology creation, got %d calls", createTopologyCalls)
	}
	if instance.AccessURL != "http://host-gateway.internal:30011" {
		t.Fatalf("unexpected access url: %s", instance.AccessURL)
	}
	if instance.HostPort != 30011 {
		t.Fatalf("expected host port to stay reserved, got %+v", instance)
	}
}

func TestCreateSingleAWDContainerRebindsHostPortAfterPublishConflict(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{
		ID:        512,
		Name:      "ctf/awd-web",
		Tag:       "v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(7012)
	teamID := int64(7112)
	serviceID := int64(8012)
	serviceSnapshot, err := contestentity.EncodeContestAWDServiceSnapshot(contestentity.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":         512,
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
	if err := db.Create(&runtimeentity.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        9012,
		WorkspaceRevision: 1,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-ctr",
		SeedSignature:     "seed:v1",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create awd defense workspace: %v", err)
	}

	createTopologyCalls := 0
	reboundBound := false
	releasedOldPort := false
	service := &serviceCore{
		repo: &stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
				return &practiceports.ContestAWDServiceRecord{
					ID:              gotServiceID,
					ContestID:       gotContestID,
					AWDChallengeID:  512,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
			reserveAvailablePortExcludingFn: func(ctx context.Context, start, end, excludedPort int) (int, error) {
				if excludedPort != 30011 {
					t.Fatalf("expected original conflicted host port to be excluded, got %d", excludedPort)
				}
				return 30012, nil
			},
			bindReservedPortFn: func(ctx context.Context, port int, instanceID int64) error {
				reboundBound = true
				if port != 30012 {
					t.Fatalf("expected rebound host port 30012, got %d", port)
				}
				if instanceID != 9012 {
					t.Fatalf("expected instance 9012, got %d", instanceID)
				}
				return nil
			},
			releaseReservedPortFn: func(ctx context.Context, port int) error {
				t.Fatalf("expected conflicted old host port release to use instance-aware path, got reserved release for %d", port)
				return nil
			},
			releasePortForInstanceFn: func(ctx context.Context, port int, instanceID int64) error {
				releasedOldPort = true
				if port != 30011 {
					t.Fatalf("expected old conflicted host port 30011 to be released, got %d", port)
				}
				if instanceID != 9012 {
					t.Fatalf("expected instance 9012 when releasing old port, got %d", instanceID)
				}
				return nil
			},
		},
		imageRepo:    challengeinfra.NewImageRepository(db),
		instanceRepo: newPracticeTestInstanceRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				createTopologyCalls++
				switch createTopologyCalls {
				case 1:
					if req.ReservedHostPort != 30011 {
						t.Fatalf("expected first attempt to use reserved host port 30011, got %d", req.ReservedHostPort)
					}
					return nil, runtimeports.ErrPublishedHostPortConflict
				case 2:
					if req.ReservedHostPort != 30012 {
						t.Fatalf("expected retry to use rebound host port 30012, got %d", req.ReservedHostPort)
					}
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "awd-rebound-ctr",
						NetworkID:          "net-awd-contest-7012",
						AccessURL:          "http://host-gateway.internal:30012",
						RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
							Networks: []runtimecontracts.InstanceRuntimeNetwork{
								{Key: challengecontracts.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-7012", NetworkID: "net-awd-contest-7012", Shared: true},
							},
							Containers: []runtimecontracts.InstanceRuntimeContainer{
								{NodeKey: "default", ContainerID: "awd-rebound-ctr", ServicePort: 8080, HostPort: 30012, IsEntryPoint: true, NetworkAliases: []string{"awd-c7012-t7112-s8012"}},
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
				PublicHost: "127.0.0.1",
				AccessHost: "host-gateway.internal",
			},
		},
	}
	instance := &instanceentity.Instance{
		ID:          9012,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 512,
		HostPort:    30011,
	}
	challenge := &challengecontracts.PracticeRuntimeChallenge{
		ID:       512,
		ImageID:  512,
		FlagType: challengecontracts.FlagTypeStatic,
	}

	if err := service.createSingleContainer(context.Background(), instance, toPracticeChallenge(challenge), "flag{demo}"); err != nil {
		t.Fatalf(
			"createSingleContainer() error = %v cause=%v conflict=%t calls=%d host_port=%d",
			err,
			errors.Unwrap(err),
			errors.Is(err, runtimeports.ErrPublishedHostPortConflict),
			createTopologyCalls,
			instance.HostPort,
		)
	}
	if createTopologyCalls != 2 {
		t.Fatalf("expected one retry after publish conflict, got %d calls", createTopologyCalls)
	}
	if !reboundBound {
		t.Fatal("expected rebound host port to be reserved and bound")
	}
	if !releasedOldPort {
		t.Fatal("expected old conflicted host port to be released after successful retry")
	}
	if instance.HostPort != 30012 {
		t.Fatalf("expected instance host port to update to rebound port, got %d", instance.HostPort)
	}
	if instance.AccessURL != "http://host-gateway.internal:30012" {
		t.Fatalf("unexpected access url after rebound retry: %s", instance.AccessURL)
	}
}
