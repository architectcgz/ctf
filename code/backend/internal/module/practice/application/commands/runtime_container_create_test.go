package commands

import (
	"context"
	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	"ctf-platform/pkg/errcode"
	"fmt"
	"testing"
	"time"
)

func TestBuildTopologyCreateRequestKeepsFineGrainedPolicies(t *testing.T) {
	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 1, Name: "ctf/web", Tag: "v1", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	service := &Service{
		imageRepo: challengeinfra.NewImageRepository(db),
		config:    &config.Config{},
	}

	request, err := service.buildTopologyCreateRequest(context.Background(), 30001, false, &model.Challenge{ImageID: 1}, "web", model.TopologySpec{
		Nodes: []model.TopologyNode{
			{Key: "web", ServicePort: 8080, InjectFlag: true},
		},
		Policies: []model.TopologyTrafficPolicy{
			{SourceNodeKey: "web", TargetNodeKey: "web", Action: model.TopologyPolicyActionAllow, Protocol: model.TopologyPolicyProtocolTCP, Ports: []int{8080}},
		},
	}, "flag{demo}")
	if err != nil {
		t.Fatalf("buildTopologyCreateRequest() error = %v", err)
	}
	if len(request.Policies) != 1 {
		t.Fatalf("expected fine-grained policy to be kept, got %+v", request.Policies)
	}
	if request.Policies[0].Protocol != model.TopologyPolicyProtocolTCP {
		t.Fatalf("unexpected policy protocol: %+v", request.Policies[0])
	}
}

func TestBuildTopologyCreateRequestRejectsSharedChallengeFlagInjection(t *testing.T) {
	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{ID: 2, Name: "ctf/web", Tag: "v2", Status: model.ImageStatusAvailable, CreatedAt: now, UpdatedAt: now}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	service := &Service{
		imageRepo: challengeinfra.NewImageRepository(db),
		config:    &config.Config{},
	}

	_, err := service.buildTopologyCreateRequest(context.Background(), 30002, false, &model.Challenge{
		ImageID:         2,
		InstanceSharing: model.InstanceSharingShared,
	}, "web", model.TopologySpec{
		Nodes: []model.TopologyNode{
			{Key: "web", ServicePort: 8080, InjectFlag: true},
		},
	}, "flag{demo}")
	if err == nil || err.Error() != errcode.ErrInvalidParams.Error() {
		t.Fatalf("expected invalid params for shared topology flag injection, got %v", err)
	}
}

func TestBuildRuntimeContainerNameUsesChallengeSlugAndContestIdentity(t *testing.T) {
	t.Parallel()

	contestID := int64(8)
	teamID := int64(15)
	serviceID := int64(21)
	packageSlug := "Bank Portal"

	got := buildRuntimeContainerName(&model.Challenge{PackageSlug: &packageSlug}, &model.Instance{
		ContestID: &contestID,
		TeamID:    &teamID,
		ServiceID: &serviceID,
	})
	want := "ctf-instance-bank-portal-c8-t15"
	if got != want {
		t.Fatalf("expected runtime container name %q, got %q", want, got)
	}
}

func TestCreateSingleAWDContainerUsesPrivateTopology(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{
		ID:        501,
		Name:      "ctf/awd-web",
		Tag:       "v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(7001)
	teamID := int64(7101)
	serviceID := int64(8001)
	serviceSnapshot, err := model.EncodeContestAWDServiceSnapshot(model.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":         501,
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
	createTopologyCalls := 0
	service := &Service{
		repo: &stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*model.ContestAWDService, error) {
				return &model.ContestAWDService{
					ID:              gotServiceID,
					ContestID:       gotContestID,
					AWDChallengeID:  501,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
		},
		imageRepo:    challengeinfra.NewImageRepository(db),
		instanceRepo: runtimeinfrarepo.NewRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				createTopologyCalls++
				switch createTopologyCalls {
				case 1:
					if req.ReservedHostPort != 0 {
						t.Fatalf("expected no reserved host port, got %d", req.ReservedHostPort)
					}
					if req.ContainerName != "ctf-instance-challenge-c7001-t7101" {
						t.Fatalf("expected awd container name, got %q", req.ContainerName)
					}
					if !req.DisableEntryPortPublishing {
						t.Fatal("expected entry port publishing to be disabled")
					}
					if len(req.Networks) != 1 || req.Networks[0].Name != "ctf-awd-contest-7001" || !req.Networks[0].Shared {
						t.Fatalf("expected stable shared AWD contest network, got %+v", req.Networks)
					}
					if len(req.Nodes) != 1 || !req.Nodes[0].IsEntryPoint || req.Nodes[0].Image != "ctf/awd-web:v1" {
						t.Fatalf("unexpected runtime topology request: %+v", req)
					}
					if len(req.Nodes[0].NetworkAliases) != 1 || req.Nodes[0].NetworkAliases[0] != "awd-c7001-t7101-s8001" {
						t.Fatalf("expected stable AWD service alias, got %+v", req.Nodes[0].NetworkAliases)
					}
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "awd-private-ctr",
						NetworkID:          "net-awd-contest-7001",
						AccessURL:          "http://awd-c7001-t7101-s8001:8080",
						RuntimeDetails: model.InstanceRuntimeDetails{
							Networks: []model.InstanceRuntimeNetwork{
								{
									Key:       model.TopologyDefaultNetworkKey,
									Name:      "ctf-awd-contest-7001",
									NetworkID: "net-awd-contest-7001",
									Shared:    true,
								},
							},
							Containers: []model.InstanceRuntimeContainer{
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
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "workspace-ctr",
						NetworkID:          "net-awd-contest-7001",
						AccessURL:          "tcp://172.30.0.20:22",
						RuntimeDetails: model.InstanceRuntimeDetails{
							Containers: []model.InstanceRuntimeContainer{
								{NodeKey: "workspace", ContainerID: "workspace-ctr", ServicePort: 22, ServiceProtocol: model.ChallengeTargetProtocolTCP, IsEntryPoint: true},
							},
						},
					}, nil
				default:
					t.Fatalf("unexpected topology create call #%d", createTopologyCalls)
					return nil, nil
				}
			},
			createContainerFn: func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error) {
				t.Fatal("AWD service instances must not use host-port CreateContainer")
				return "", "", 0, 0, nil
			},
		},
	}
	instance := &model.Instance{
		ID:          9001,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 501,
	}
	challenge := &model.Challenge{
		ID:       501,
		ImageID:  501,
		FlagType: model.FlagTypeStatic,
	}

	if err := service.createSingleContainer(context.Background(), instance, challenge, "flag{demo}"); err != nil {
		t.Fatalf("createSingleContainer() error = %v", err)
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
}

func TestCreateTopologyAWDContainerUsesStableContestNetwork(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{
		ID:        503,
		Name:      "ctf/awd-topology",
		Tag:       "v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(7003)
	teamID := int64(7103)
	serviceID := int64(8003)
	serviceSnapshot, err := model.EncodeContestAWDServiceSnapshot(model.ContestAWDServiceSnapshot{
		Name: "AWD Topology",
		RuntimeConfig: map[string]any{
			"image_id":         503,
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
	createTopologyCalls := 0
	service := &Service{
		repo: &stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*model.ContestAWDService, error) {
				return &model.ContestAWDService{
					ID:              gotServiceID,
					ContestID:       gotContestID,
					AWDChallengeID:  503,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
		},
		imageRepo:    challengeinfra.NewImageRepository(db),
		instanceRepo: runtimeinfrarepo.NewRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				createTopologyCalls++
				switch createTopologyCalls {
				case 1:
					if req.ReservedHostPort != 0 {
						t.Fatalf("expected no reserved host port, got %d", req.ReservedHostPort)
					}
					if req.ContainerName != "ctf-instance-challenge-c7003-t7103" {
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
					if len(req.Nodes[0].NetworkAliases) != 1 || req.Nodes[0].NetworkAliases[0] != "awd-c7003-t7103-s8003" {
						t.Fatalf("expected stable AWD service alias, got %+v", req.Nodes[0].NetworkAliases)
					}
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "awd-topology-ctr",
						NetworkID:          "net-awd-contest-7003",
						AccessURL:          "http://awd-c7003-t7103-s8003:8080",
						RuntimeDetails: model.InstanceRuntimeDetails{
							Networks: []model.InstanceRuntimeNetwork{
								{Key: model.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-7003", NetworkID: "net-awd-contest-7003", Shared: true},
							},
							Containers: []model.InstanceRuntimeContainer{
								{NodeKey: "web", ContainerID: "awd-topology-ctr", ServicePort: 8080, IsEntryPoint: true, NetworkAliases: []string{"awd-c7003-t7103-s8003"}},
							},
						},
					}, nil
				case 2:
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "workspace-ctr",
						NetworkID:          "net-awd-contest-7003",
						AccessURL:          "tcp://172.30.0.21:22",
						RuntimeDetails: model.InstanceRuntimeDetails{
							Containers: []model.InstanceRuntimeContainer{
								{NodeKey: "workspace", ContainerID: "workspace-ctr", ServicePort: 22, ServiceProtocol: model.ChallengeTargetProtocolTCP, IsEntryPoint: true},
							},
						},
					}, nil
				default:
					t.Fatalf("unexpected topology create call #%d", createTopologyCalls)
					return nil, nil
				}
			},
		},
	}
	instance := &model.Instance{
		ID:          9003,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 503,
	}
	challenge := &model.Challenge{
		ID:       503,
		ImageID:  503,
		FlagType: model.FlagTypeStatic,
	}
	topology, err := model.EncodeTopologySpec(model.TopologySpec{
		Nodes: []model.TopologyNode{
			{Key: "web", ServicePort: 8080, InjectFlag: true},
		},
	})
	if err != nil {
		t.Fatalf("encode topology: %v", err)
	}

	if err := service.createContainer(context.Background(), instance, challenge, &model.ChallengeTopology{
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

func TestCreateSingleAWDContainerCreatesWorkspaceCompanionWithSharedMounts(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{
		ID:        601,
		Name:      "ctf/awd-workspace",
		Tag:       "v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(7601)
	teamID := int64(7701)
	serviceID := int64(7801)
	serviceSnapshot, err := model.EncodeContestAWDServiceSnapshot(model.ContestAWDServiceSnapshot{
		Name: "Campus Drive",
		RuntimeConfig: map[string]any{
			"image_id":         601,
			"instance_sharing": string(model.InstanceSharingPerTeam),
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
		findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*model.ContestAWDService, error) {
			if gotContestID != contestID || gotServiceID != serviceID {
				t.Fatalf("unexpected awd service lookup: contest=%d service=%d", gotContestID, gotServiceID)
			}
			return &model.ContestAWDService{
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
		instanceRepo: runtimeinfrarepo.NewRepository(db),
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
						RuntimeDetails: model.InstanceRuntimeDetails{
							Networks: []model.InstanceRuntimeNetwork{
								{Key: model.TopologyDefaultNetworkKey, Name: "ctf-awd-contest-7601", NetworkID: "net-awd-contest-7601", Shared: true},
							},
							Containers: []model.InstanceRuntimeContainer{
								{NodeKey: "default", ContainerID: "runtime-ctr", ServicePort: 8080, IsEntryPoint: true, NetworkAliases: []string{"awd-c7601-t7701-s7801"}},
							},
						},
					}, nil
				case 2:
					if len(req.Nodes) != 1 || req.Nodes[0].WorkingDir != "/workspace" {
						t.Fatalf("unexpected workspace topology request: %+v", req)
					}
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
						RuntimeDetails: model.InstanceRuntimeDetails{
							Containers: []model.InstanceRuntimeContainer{
								{NodeKey: "workspace", ContainerID: "workspace-ctr", ServicePort: 22, ServiceProtocol: model.ChallengeTargetProtocolTCP, IsEntryPoint: true},
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

	instance := &model.Instance{
		ID:          9001,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ServiceID:   &serviceID,
		ChallengeID: 601,
	}
	challenge := &model.Challenge{
		ID:          601,
		ImageID:     601,
		FlagType:    model.FlagTypeStatic,
		PackageSlug: stringPtr("campus-drive"),
	}

	if err := service.createSingleContainer(context.Background(), instance, challenge, "flag{demo}"); err != nil {
		t.Fatalf("createSingleContainer() error = %v", err)
	}
	if len(requests) != 2 {
		t.Fatalf("expected runtime and workspace topology creation, got %d calls", len(requests))
	}
	if instance.ContainerID != "runtime-ctr" || instance.AccessURL != "http://awd-c7601-t7701-s7801:8080" {
		t.Fatalf("unexpected runtime instance after createSingleContainer(): %+v", instance)
	}

	workspace, err := runtimeinfrarepo.NewRepository(db).FindAWDDefenseWorkspace(context.Background(), contestID, teamID, serviceID)
	if err != nil {
		t.Fatalf("FindAWDDefenseWorkspace() error = %v", err)
	}
	if workspace == nil {
		t.Fatal("expected workspace row to be created")
	}
	if workspace.WorkspaceRevision != 1 || workspace.Status != model.AWDDefenseWorkspaceStatusRunning || workspace.ContainerID != "workspace-ctr" {
		t.Fatalf("unexpected workspace state: %+v", workspace)
	}
}

func TestLoadRuntimeSubjectWithScopePropagatesContextToChallengeContract(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("runtime-subject")
	expectedCtxValue := "ctx-runtime-subject"
	challengeLookupCalled := false
	topologyLookupCalled := false
	service := NewService(
		nil,
		&stubPracticeChallengeContract{
			findByIDWithContextFn: func(ctx context.Context, id int64) (*model.Challenge, error) {
				challengeLookupCalled = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected challenge lookup ctx value %v, got %v", expectedCtxValue, got)
				}
				return &model.Challenge{ID: id, Status: model.ChallengeStatusPublished}, nil
			},
			findChallengeTopologyByChallengeIDFn: func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
				topologyLookupCalled = true
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected topology lookup ctx value %v, got %v", expectedCtxValue, got)
				}
				return nil, nil
			},
		},
		nil,
		nil,
		nil,
		nil,
		nil,
		nil,
		&config.Config{},
		nil,
	)

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	challenge, topology, err := service.loadRuntimeSubjectWithScope(ctx, practiceports.InstanceScope{}, 42)
	if err != nil {
		t.Fatalf("loadRuntimeSubjectWithScope() error = %v", err)
	}
	if challenge == nil || challenge.ID != 42 {
		t.Fatalf("expected challenge 42, got %+v", challenge)
	}
	if topology != nil {
		t.Fatalf("expected nil topology, got %+v", topology)
	}
	if !challengeLookupCalled {
		t.Fatal("expected challenge lookup to be called")
	}
	if !topologyLookupCalled {
		t.Fatal("expected topology lookup to be called")
	}
}

func TestBuildTopologyCreateRequestPropagatesContextToImageRepository(t *testing.T) {
	t.Parallel()

	ctxKey := practiceServiceContextKey("topology-image")
	expectedCtxValue := "ctx-topology-image"
	lookups := make([]int64, 0, 2)
	service := &Service{
		imageRepo: &stubPracticeImageStore{
			findByIDFn: func(ctx context.Context, id int64) (*model.Image, error) {
				if got := ctx.Value(ctxKey); got != expectedCtxValue {
					t.Fatalf("expected image lookup ctx value %v, got %v", expectedCtxValue, got)
				}
				lookups = append(lookups, id)
				return &model.Image{ID: id, Name: fmt.Sprintf("repo/%d", id), Tag: "latest", Status: model.ImageStatusAvailable}, nil
			},
		},
		config: &config.Config{},
	}

	ctx := context.WithValue(context.Background(), ctxKey, expectedCtxValue)
	request, err := service.buildTopologyCreateRequest(ctx, 30001, false, &model.Challenge{ImageID: 1}, "web", model.TopologySpec{
		Nodes: []model.TopologyNode{
			{Key: "web", Name: "Web", ServicePort: 8080},
			{Key: "worker", Name: "Worker", ImageID: 2, ServicePort: 9000},
		},
	}, "flag{ctx-image}")
	if err != nil {
		t.Fatalf("buildTopologyCreateRequest() error = %v", err)
	}
	if len(request.Nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %+v", request.Nodes)
	}
	if len(lookups) != 2 || lookups[0] != 1 || lookups[1] != 2 {
		t.Fatalf("expected image lookups [1 2], got %v", lookups)
	}
}

func stringPtr(value string) *string {
	return &value
}
