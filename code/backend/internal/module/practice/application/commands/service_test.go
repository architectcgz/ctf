package commands

import (
	"context"
	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	runtimecontracts "ctf-platform/internal/module/container_runtime/contracts"
	containerruntimeentity "ctf-platform/internal/module/container_runtime/entity"
	runtimeentity "ctf-platform/internal/module/contest/entity"
	contestinfrarepo "ctf-platform/internal/module/contest/infrastructure"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	contestentity "ctf-platform/internal/module/practice/testsupport/contestentity"
	"ctf-platform/internal/platform/events"
	"errors"
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
	"net"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func wirePracticeScopeAdapters(
	service *serviceCore,
	repo practiceports.PracticeContestScopeRepository,
	challengeRepo challengecontracts.PracticeChallengeContract,
) *serviceCore {
	if service == nil {
		return nil
	}
	return service.
		SetContestScopeRepository(practiceinfra.NewContestScopeRepository(repo)).
		SetRuntimeSubjectRepository(practiceinfra.NewRuntimeSubjectRepository(challengeRepo))
}

func wirePracticeManualReviewAdapters(
	service *serviceCore,
	repo practiceports.PracticeManualReviewRepository,
	challengeRepo challengecontracts.PracticeChallengeContract,
) *serviceCore {
	if service == nil {
		return nil
	}
	if repo != nil {
		service = service.SetManualReviewRepository(practiceinfra.NewManualReviewRepository(repo))
		service = service.SetSolvedSubmissionRepository(practiceinfra.NewSolvedSubmissionRepository(repo))
	}
	if challengeRepo != nil {
		service = service.SetRuntimeSubjectRepository(practiceinfra.NewRuntimeSubjectRepository(challengeRepo))
	}
	return service
}

func wirePracticeSubmissionAdapters(
	service *serviceCore,
	repo practiceports.PracticeSolvedSubmissionRepository,
	challengeRepo challengecontracts.PracticeChallengeContract,
) *serviceCore {
	if service == nil {
		return nil
	}
	if repo != nil {
		service = service.SetSolvedSubmissionRepository(practiceinfra.NewSolvedSubmissionRepository(repo))
	}
	if challengeRepo != nil {
		service = service.SetRuntimeSubjectRepository(practiceinfra.NewRuntimeSubjectRepository(challengeRepo))
	}
	return service
}

func wirePracticeSubmissionHistoryAdapters(
	service *serviceCore,
	challengeRepo challengecontracts.PracticeChallengeContract,
) *serviceCore {
	if service == nil {
		return nil
	}
	if challengeRepo != nil {
		service = service.SetRuntimeSubjectRepository(practiceinfra.NewRuntimeSubjectRepository(challengeRepo))
	}
	return service
}

type stubPracticeRuntimeService struct {
	cleanupRuntimeFn          func(ctx context.Context, instance *instanceentity.Instance) error
	createTopologyFn          func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error)
	createContainerFn         func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, nodeID int64) (containerID, networkID string, hostPort, servicePort int, err error)
	inspectManagedContainerFn func(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error)
}

func (s *stubPracticeRuntimeService) CleanupRuntime(ctx context.Context, instance *instanceentity.Instance) error {
	if s.cleanupRuntimeFn == nil {
		return nil
	}
	return s.cleanupRuntimeFn(ctx, instance)
}

func (s *stubPracticeRuntimeService) CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if s.createTopologyFn == nil {
		if s.createContainerFn == nil {
			return nil, errors.New("unexpected CreateTopology call")
		}
		if req == nil || len(req.Nodes) != 1 {
			return nil, errors.New("unexpected CreateTopology call")
		}
		node := req.Nodes[0]
		if !node.IsEntryPoint {
			return nil, errors.New("unexpected CreateTopology call")
		}
		containerID, networkID, hostPort, servicePort, err := s.createContainerFn(ctx, node.Image, node.Env, req.ReservedHostPort, req.NodeID)
		if err != nil {
			return nil, err
		}
		result := &practiceports.TopologyCreateResult{
			PrimaryContainerID: containerID,
			NetworkID:          networkID,
			RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
				Networks: []runtimecontracts.InstanceRuntimeNetwork{
					{
						Key:       resolvedPracticeTopologyNetworkKey(req.Networks, node.NetworkKeys),
						Name:      resolvedPracticeTopologyNetworkName(req.Networks, node.NetworkKeys),
						NetworkID: networkID,
						Subnet:    resolvedPracticeTopologyNetworkSubnet(req.Networks, node.NetworkKeys),
					},
				},
				Containers: []runtimecontracts.InstanceRuntimeContainer{
					{
						NodeKey:         node.Key,
						ContainerID:     containerID,
						HostPort:        hostPort,
						ServicePort:     servicePort,
						ServiceProtocol: node.ServiceProtocol,
						IsEntryPoint:    true,
						NetworkKeys:     append([]string(nil), node.NetworkKeys...),
						NetworkAliases:  append([]string(nil), node.NetworkAliases...),
					},
				},
			},
		}
		if hostPort > 0 {
			host := "127.0.0.1"
			if servicePort <= 0 {
				servicePort = node.ServicePort
			}
			if servicePort > 0 {
				result.AccessURL = fmt.Sprintf("http://%s:%d", host, hostPort)
			}
		}
		return result, nil
	}
	return s.createTopologyFn(ctx, req)
}

func (s *stubPracticeRuntimeService) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int, nodeID int64) (string, string, int, int, error) {
	if s.createContainerFn == nil {
		return "", "", 0, 0, errors.New("unexpected CreateContainer call")
	}
	return s.createContainerFn(ctx, imageName, env, reservedHostPort, nodeID)
}

func (s *stubPracticeRuntimeService) InspectManagedContainer(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error) {
	if s.inspectManagedContainerFn == nil {
		return &practiceports.ManagedContainerState{
			ID:      containerID,
			Exists:  true,
			Running: true,
			Status:  "running",
		}, nil
	}
	return s.inspectManagedContainerFn(ctx, containerID)
}

type stubPracticeRuntimeNodeSelector struct {
	selectRuntimeNodeFn func(ctx context.Context, scope practiceports.InstanceScope) (*practiceports.RuntimeNodeBinding, error)
}

func (s *stubPracticeRuntimeNodeSelector) SelectRuntimeNode(ctx context.Context, scope practiceports.InstanceScope) (*practiceports.RuntimeNodeBinding, error) {
	if s == nil || s.selectRuntimeNodeFn == nil {
		return nil, nil
	}
	return s.selectRuntimeNodeFn(ctx, scope)
}

func resolvedPracticeTopologyNetworkKey(networks []practiceports.TopologyCreateNetwork, nodeNetworkKeys []string) string {
	if len(nodeNetworkKeys) > 0 && strings.TrimSpace(nodeNetworkKeys[0]) != "" {
		return strings.TrimSpace(nodeNetworkKeys[0])
	}
	if len(networks) > 0 && strings.TrimSpace(networks[0].Key) != "" {
		return strings.TrimSpace(networks[0].Key)
	}
	return challengecontracts.TopologyDefaultNetworkKey
}

func resolvedPracticeTopologyNetworkName(networks []practiceports.TopologyCreateNetwork, nodeNetworkKeys []string) string {
	targetKey := resolvedPracticeTopologyNetworkKey(networks, nodeNetworkKeys)
	for _, network := range networks {
		if strings.TrimSpace(network.Key) == targetKey {
			return network.Name
		}
	}
	return ""
}

func resolvedPracticeTopologyNetworkSubnet(networks []practiceports.TopologyCreateNetwork, nodeNetworkKeys []string) string {
	targetKey := resolvedPracticeTopologyNetworkKey(networks, nodeNetworkKeys)
	for _, network := range networks {
		if strings.TrimSpace(network.Key) == targetKey {
			return network.Subnet
		}
	}
	return ""
}

type stubPracticeEventBus struct {
	publishFn func(ctx context.Context, evt events.Event) error
}

func (s *stubPracticeEventBus) Subscribe(string, events.Handler) {}

func (s *stubPracticeEventBus) Publish(ctx context.Context, evt events.Event) error {
	if s.publishFn != nil {
		return s.publishFn(ctx, evt)
	}
	return nil
}

func requireEventually(t *testing.T, timeout time.Duration, check func() bool) {
	t.Helper()

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if check() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("condition was not satisfied before timeout")
}

func TestCreateAWDDefenseWorkspaceCompanionInitializesGitReposForWritableMounts(t *testing.T) {
	contestID := int64(8)
	teamID := int64(15)
	serviceID := int64(21)

	service := &serviceCore{
		runtimeService: &stubPracticeRuntimeService{
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				if len(req.Nodes) != 1 {
					t.Fatalf("expected one workspace node, got %+v", req.Nodes)
				}
				if len(req.Nodes[0].Command) != 3 {
					t.Fatalf("unexpected workspace shell command: %+v", req.Nodes[0].Command)
				}

				command := req.Nodes[0].Command[2]
				requiredFragments := []string{
					"set -e",
					"[ -d '/workspace/app' ]",
					"[ ! -d '/workspace/app/.git' ]",
					"git -C '/workspace/app' init",
					"[ -d '/workspace/templates' ]",
					"[ ! -d '/workspace/templates/.git' ]",
					"git -C '/workspace/templates' init",
					"git -C '/workspace/templates' commit --allow-empty -m 'Initial workspace snapshot'",
				}
				for _, fragment := range requiredFragments {
					if !strings.Contains(command, fragment) {
						t.Fatalf("expected workspace bootstrap command to contain %q, got %q", fragment, command)
					}
				}
				if strings.Contains(command, "/workspace/data/.git") {
					t.Fatalf("expected readonly workspace root to skip git initialization, got %q", command)
				}

				return &practiceports.TopologyCreateResult{
					PrimaryContainerID: "workspace-ctr",
				}, nil
			},
		},
	}

	_, err := service.createAWDDefenseWorkspaceCompanion(context.Background(), &instanceentity.Instance{
		ContestID: &contestID,
		TeamID:    &teamID,
		ServiceID: &serviceID,
	}, &awdDefenseWorkspacePlan{
		workspaceRevision:      2,
		workspaceContainerName: "ctf-workspace-custom",
		workspaceMounts: []runtimecontracts.ContainerMount{
			{Source: "ws-app", Target: "/workspace/app"},
			{Source: "ws-templates", Target: "/workspace/templates"},
			{Source: "ws-data", Target: "/workspace/data", ReadOnly: true},
		},
	})
	if err != nil {
		t.Fatalf("createAWDDefenseWorkspaceCompanion() error = %v", err)
	}
}

func TestContestAWDServiceRuntimeSubjectMapsWorkspaceRootsOutsideWritableSetAsReadonly(t *testing.T) {
	service := &contestentity.ContestAWDService{
		ID:              21,
		ContestID:       8,
		AWDChallengeID:  13,
		ServiceSnapshot: `{"name":"Workspace Service","runtime_config":{"defense_workspace":{"seed_root":"docker/workspace","workspace_roots":["docker/workspace/src","docker/workspace/templates","docker/workspace/data"],"writable_roots":["docker/workspace/src"],"readonly_roots":["docker/workspace/data"],"runtime_mounts":[{"source":"docker/workspace/src","target":"/workspace/src","mode":"rw"},{"source":"docker/workspace/templates","target":"/workspace/templates","mode":"ro"},{"source":"docker/workspace/data","target":"/workspace/data","mode":"ro"}]}}}`,
	}

	subject, err := stubContestAWDServiceRuntimeSubject(practiceContestAWDServiceRecordFromEntity(service))
	if err != nil {
		t.Fatalf("stubContestAWDServiceRuntimeSubject() error = %v", err)
	}
	if subject == nil || subject.WorkspaceConfig == nil {
		t.Fatalf("expected workspace config, got %+v", subject)
	}

	config := subject.WorkspaceConfig
	if len(config.WorkspaceRoots) != 3 {
		t.Fatalf("expected three workspace roots, got %+v", config.WorkspaceRoots)
	}

	readonlyBySource := make(map[string]bool, len(config.WorkspaceRoots))
	for _, root := range config.WorkspaceRoots {
		readonlyBySource[root.Source] = root.ReadOnly
	}
	if readonlyBySource["docker/workspace/src"] {
		t.Fatalf("expected src root to stay writable, got %+v", config.WorkspaceRoots)
	}
	if !readonlyBySource["docker/workspace/templates"] {
		t.Fatalf("expected template root outside writable_roots to default readonly, got %+v", config.WorkspaceRoots)
	}
	if !readonlyBySource["docker/workspace/data"] {
		t.Fatalf("expected readonly root to stay readonly, got %+v", config.WorkspaceRoots)
	}
}

func TestBuildAWDDefenseWorkspaceBootstrapCommandDegradesGracefullyWithoutPackageInstall(t *testing.T) {
	command := buildAWDDefenseWorkspaceBootstrapCommand([]runtimecontracts.ContainerMount{
		{Target: "/workspace/src"},
		{Target: "/workspace/data", ReadOnly: true},
	})

	requiredFragments := []string{
		`missing_tools=""`,
		`apk add --no-cache $missing_tools || true`,
		`if command -v git >/dev/null 2>&1 && [ -d '/workspace/src' ] && [ ! -d '/workspace/src/.git' ]; then`,
	}
	for _, fragment := range requiredFragments {
		if !strings.Contains(command, fragment) {
			t.Fatalf("expected workspace bootstrap command to contain %q, got %q", fragment, command)
		}
	}
	if strings.Contains(command, "/workspace/data/.git") {
		t.Fatalf("expected readonly workspace root to skip git initialization, got %q", command)
	}
}

func TestCreateSingleAWDContainerRemovesStoppedWorkspaceCompanionBeforeRecreate(t *testing.T) {
	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{
		ID:        601,
		Name:      "ctf/awd-web",
		Tag:       "v1",
		Digest:    "sha256:awd-web-v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(801)
	teamID := int64(802)
	serviceID := int64(803)
	challengeID := int64(804)
	if err := db.Create(&runtimeentity.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        9001,
		WorkspaceRevision: 1,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-stale-ctr",
		SeedSignature:     "seed-signature",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create workspace state: %v", err)
	}

	serviceSnapshot, err := contestentity.EncodeContestAWDServiceSnapshot(contestentity.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":         601,
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

	cleanupCalls := 0
	createTopologyCalls := 0
	service := &serviceCore{
		repo: &stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
				return &practiceports.ContestAWDServiceRecord{
					ID:              gotServiceID,
					ContestID:       gotContestID,
					AWDChallengeID:  challengeID,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
		},
		imageRepo:    challengeinfra.NewImageRepository(db),
		instanceRepo: newPracticeTestInstanceRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, instance *instanceentity.Instance) error {
				cleanupCalls++
				details, err := runtimecontracts.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
				if err != nil {
					t.Fatalf("decode cleanup runtime details: %v", err)
				}
				if len(details.Containers) != 1 || details.Containers[0].ContainerID != "workspace-stale-ctr" {
					t.Fatalf("expected stale workspace companion cleanup, got %+v", details.Containers)
				}
				if len(details.Networks) != 0 {
					t.Fatalf("expected stale workspace cleanup to avoid network removal, got %+v", details.Networks)
				}
				return nil
			},
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				createTopologyCalls++
				switch createTopologyCalls {
				case 1:
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "awd-private-ctr",
						NetworkID:          "net-awd-contest-801",
						AccessURL:          "http://awd-c801-t802-s803:8080",
						RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
							Networks: []runtimecontracts.InstanceRuntimeNetwork{
								{
									Key:       runtimecontracts.TopologyDefaultNetworkKey,
									Name:      "ctf-awd-contest-801",
									NetworkID: "net-awd-contest-801",
									Shared:    true,
								},
							},
							Containers: []runtimecontracts.InstanceRuntimeContainer{
								{
									NodeKey:        "default",
									ContainerID:    "awd-private-ctr",
									ServicePort:    8080,
									IsEntryPoint:   true,
									NetworkAliases: []string{"awd-c801-t802-s803"},
								},
							},
						},
					}, nil
				case 2:
					return &practiceports.TopologyCreateResult{
						PrimaryContainerID: "workspace-fresh-ctr",
					}, nil
				default:
					t.Fatalf("unexpected create topology call: %+v", req)
					return nil, nil
				}
			},
			inspectManagedContainerFn: func(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error) {
				if containerID != "workspace-stale-ctr" {
					t.Fatalf("unexpected inspected workspace container id: %s", containerID)
				}
				return &practiceports.ManagedContainerState{
					ID:      containerID,
					Exists:  true,
					Running: false,
					Status:  "exited",
				}, nil
			},
		},
		config: &config.Config{},
	}

	instance := &instanceentity.Instance{
		ID:          9001,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
	}
	if err := service.createSingleContainer(context.Background(), instance, toPracticeChallenge(&challengecontracts.PracticeRuntimeChallenge{
		ID:             challengeID,
		ImageID:        int64Ptr(601),
		TargetPort:     8080,
		TargetProtocol: challengecontracts.ChallengeTargetProtocolHTTP,
	}), "flag{demo}"); err != nil {
		t.Fatalf("createSingleContainer() error = %v", err)
	}

	if cleanupCalls != 1 {
		t.Fatalf("expected one stale workspace cleanup call, got %d", cleanupCalls)
	}
	if createTopologyCalls != 2 {
		t.Fatalf("expected runtime and workspace topology creation, got %d", createTopologyCalls)
	}
}

func TestCleanupAWDDefenseWorkspaceCompanionUsesContainerAuthorityPayload(t *testing.T) {
	service := &serviceCore{
		runtimeService: &stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, instance *instanceentity.Instance) error {
				if instance == nil {
					t.Fatal("expected cleanup instance payload")
				}
				if instance.NodeID != nil {
					t.Fatalf("expected workspace cleanup payload to leave node resolution to router, got node_id=%d", *instance.NodeID)
				}
				if strings.TrimSpace(instance.ContainerID) != "" {
					t.Fatalf("expected workspace cleanup payload to use runtime_details only, got container_id=%q", instance.ContainerID)
				}
				details, err := runtimecontracts.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
				if err != nil {
					t.Fatalf("decode cleanup runtime details: %v", err)
				}
				if len(details.Containers) != 1 || details.Containers[0].ContainerID != "workspace-cleanup-ctr" {
					t.Fatalf("expected cleanup runtime details to carry workspace container id, got %+v", details.Containers)
				}
				return nil
			},
		},
	}

	if err := service.cleanupAWDDefenseWorkspaceCompanion(context.Background(), "workspace-cleanup-ctr"); err != nil {
		t.Fatalf("cleanupAWDDefenseWorkspaceCompanion() error = %v", err)
	}
}

func TestCreateSingleAWDContainerPreservesStaleWorkspaceReferenceWhenCleanupFails(t *testing.T) {
	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&practiceCommandImageRow{
		ID:        602,
		Name:      "ctf/awd-web",
		Tag:       "v1",
		Digest:    "sha256:awd-web-v1",
		Status:    challengecontracts.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(811)
	teamID := int64(812)
	serviceID := int64(813)
	challengeID := int64(814)
	if err := db.Create(&runtimeentity.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        9011,
		WorkspaceRevision: 1,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-stale-ctr",
		SeedSignature:     "seed-signature",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create workspace state: %v", err)
	}

	serviceSnapshot, err := contestentity.EncodeContestAWDServiceSnapshot(contestentity.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":         602,
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

	service := &serviceCore{
		repo: &stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
				return &practiceports.ContestAWDServiceRecord{
					ID:              gotServiceID,
					ContestID:       gotContestID,
					AWDChallengeID:  challengeID,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
		},
		imageRepo:    challengeinfra.NewImageRepository(db),
		instanceRepo: newPracticeTestInstanceRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, instance *instanceentity.Instance) error {
				return fmt.Errorf("cleanup stale workspace failed")
			},
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				return &practiceports.TopologyCreateResult{
					PrimaryContainerID: "awd-private-ctr",
					NetworkID:          "net-awd-contest-811",
					AccessURL:          "http://awd-c811-t812-s813:8080",
					RuntimeDetails: runtimecontracts.InstanceRuntimeDetails{
						Networks: []runtimecontracts.InstanceRuntimeNetwork{
							{
								Key:       runtimecontracts.TopologyDefaultNetworkKey,
								Name:      "ctf-awd-contest-811",
								NetworkID: "net-awd-contest-811",
								Shared:    true,
							},
						},
						Containers: []runtimecontracts.InstanceRuntimeContainer{
							{
								NodeKey:        "default",
								ContainerID:    "awd-private-ctr",
								ServicePort:    8080,
								IsEntryPoint:   true,
								NetworkAliases: []string{"awd-c811-t812-s813"},
							},
						},
					},
				}, nil
			},
			inspectManagedContainerFn: func(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error) {
				return &practiceports.ManagedContainerState{
					ID:      containerID,
					Exists:  true,
					Running: false,
					Status:  "exited",
				}, nil
			},
		},
		config: &config.Config{},
	}

	instance := &instanceentity.Instance{
		ID:          9011,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
	}
	err = service.createSingleContainer(context.Background(), instance, toPracticeChallenge(&challengecontracts.PracticeRuntimeChallenge{
		ID:             challengeID,
		ImageID:        int64Ptr(602),
		TargetPort:     8080,
		TargetProtocol: challengecontracts.ChallengeTargetProtocolHTTP,
	}), "flag{demo}")
	if err == nil {
		t.Fatal("expected createSingleContainer() to fail when stale workspace cleanup fails")
	}

	workspace, err := contestinfrarepo.NewAWDRepository(db).FindAWDDefenseWorkspace(context.Background(), contestID, teamID, serviceID)
	if err != nil {
		t.Fatalf("FindAWDDefenseWorkspace() error = %v", err)
	}
	if workspace == nil {
		t.Fatal("expected workspace row to exist")
	}
	if workspace.Status != runtimeentity.AWDDefenseWorkspaceStatusFailed {
		t.Fatalf("expected failed workspace state, got %+v", workspace)
	}
	if workspace.ContainerID != "workspace-stale-ctr" {
		t.Fatalf("expected stale workspace container id to be preserved, got %+v", workspace)
	}
}

func TestPrepareAWDDefenseWorkspacePlanTreatsFailedWorkspaceContainerAsStale(t *testing.T) {
	db := newPracticeCommandTestDB(t)
	now := time.Now()

	contestID := int64(821)
	teamID := int64(822)
	serviceID := int64(823)
	challengeID := int64(824)
	if err := db.Create(&runtimeentity.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        9021,
		WorkspaceRevision: 2,
		Status:            runtimeentity.AWDDefenseWorkspaceStatusFailed,
		ContainerID:       "workspace-stale-ctr",
		SeedSignature:     "seed-signature",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create workspace state: %v", err)
	}

	serviceSnapshot, err := contestentity.EncodeContestAWDServiceSnapshot(contestentity.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":         challengeID,
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

	service := &serviceCore{
		repo: &stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*practiceports.ContestAWDServiceRecord, error) {
				return &practiceports.ContestAWDServiceRecord{
					ID:              gotServiceID,
					ContestID:       gotContestID,
					AWDChallengeID:  challengeID,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
		},
		instanceRepo: newPracticeTestInstanceRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			inspectManagedContainerFn: func(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error) {
				t.Fatalf("unexpected managed container inspection for failed workspace state: %s", containerID)
				return nil, nil
			},
		},
		config: &config.Config{},
	}

	plan, err := service.prepareAWDDefenseWorkspacePlan(context.Background(), &instanceentity.Instance{
		ID:          9021,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
	}, toPracticeChallenge(&challengecontracts.PracticeRuntimeChallenge{ID: challengeID}))
	if err != nil {
		t.Fatalf("prepareAWDDefenseWorkspacePlan() error = %v", err)
	}
	if plan == nil {
		t.Fatal("expected workspace plan")
	}
	if !plan.createWorkspace {
		t.Fatalf("expected failed workspace state to force recreate, got %+v", plan)
	}
	if plan.staleWorkspaceContainerID != "workspace-stale-ctr" {
		t.Fatalf("expected failed workspace container to be marked stale, got %+v", plan)
	}
	if plan.workspaceContainerID != "" {
		t.Fatalf("expected stale workspace container id to be removed from active slot, got %+v", plan)
	}
}

func newPracticeCommandTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("%s/%s.sqlite", t.TempDir(), strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&practiceCommandImageRow{},
		&practiceCommandChallengeRow{},
		&practiceCommandChallengeTopologyRow{},
		&contestentity.Contest{},
		&contestentity.ContestAWDService{},
		&contestentity.ContestRegistration{},
		&identitycontracts.User{},
		&contestentity.Team{},
		&instanceentity.Instance{},
		&runtimeentity.AWDServiceOperation{},
		&runtimeentity.AWDScopeControl{},
		&runtimeentity.AWDDefenseWorkspace{},
		&containerruntimeentity.PortAllocation{},
		&containerruntimeentity.NetworkAllocation{},
		&containerruntimeentity.RuntimeNode{},
		&contestentity.Submission{},
		&events.OutboxRecord{},
	); err != nil {
		t.Fatalf("migrate practice command tables: %v", err)
	}
	return db
}

func TestPracticeCommandDBMigratesRuntimeNodeSchema(t *testing.T) {
	t.Parallel()

	db := newPracticeCommandTestDB(t)

	type sqliteMasterRow struct {
		Name string `gorm:"column:name"`
	}
	var runtimeNodesTable sqliteMasterRow
	if err := db.Raw("SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?", "runtime_nodes").Scan(&runtimeNodesTable).Error; err != nil {
		t.Fatalf("query runtime_nodes table: %v", err)
	}
	if runtimeNodesTable.Name != "runtime_nodes" {
		t.Fatalf("expected runtime_nodes table to be migrated, got %+v", runtimeNodesTable)
	}

	type tableInfoRow struct {
		Name string `gorm:"column:name"`
	}
	var columns []tableInfoRow
	if err := db.Raw("PRAGMA table_info(instances)").Scan(&columns).Error; err != nil {
		t.Fatalf("query instances table info: %v", err)
	}
	for _, column := range columns {
		if column.Name == "node_id" {
			return
		}
	}
	t.Fatalf("expected instances table to contain node_id column, got %+v", columns)
}

func reserveClosedLoopbackPort(t *testing.T) int {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen loopback port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatalf("close loopback listener: %v", err)
	}
	return port
}

func parseHTTPServerEndpoint(t *testing.T, rawURL string) (string, int) {
	t.Helper()

	parsed, err := url.Parse(rawURL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}

	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		t.Fatalf("parse server port: %v", err)
	}
	return parsed.Hostname(), port
}

func assertAWDDefenseWorkspaceShellNode(t *testing.T, node practiceports.TopologyCreateNode) {
	t.Helper()

	if node.Image != awdDefenseWorkspaceShellImage {
		t.Fatalf("unexpected workspace shell image: %q", node.Image)
	}
	if !reflect.DeepEqual(node.Env, awdDefenseWorkspaceShellEnv) {
		t.Fatalf("unexpected workspace shell env: %+v", node.Env)
	}
	wantCommand := []string{"/bin/sh", "-lc", buildAWDDefenseWorkspaceBootstrapCommand(node.Mounts)}
	if !reflect.DeepEqual(node.Command, wantCommand) {
		t.Fatalf("unexpected workspace shell command: %+v", node.Command)
	}
}

type stubScoreUpdater struct {
	updateFn func(ctx context.Context, userID int64) error
	lockWait time.Duration
}

func (s *stubScoreUpdater) UpdateUserScore(ctx context.Context, userID int64) error {
	if s.updateFn == nil {
		return nil
	}
	return s.updateFn(ctx, userID)
}

func (s *stubScoreUpdater) lockTimeout() time.Duration {
	return s.lockWait
}

type stubPracticeChallengeContract struct {
	findByIDWithContextFn                func(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error)
	findChallengeTopologyByChallengeIDFn func(ctx context.Context, challengeID int64) (*challengecontracts.PracticeRuntimeChallengeTopology, error)
}

func (s *stubPracticeChallengeContract) FindPracticeRuntimeChallengeByID(ctx context.Context, id int64) (*challengecontracts.PracticeRuntimeChallenge, error) {
	if s.findByIDWithContextFn != nil {
		challenge, err := s.findByIDWithContextFn(ctx, id)
		if err != nil || challenge == nil {
			return nil, err
		}
		return &challengecontracts.PracticeRuntimeChallenge{
			ID:              challenge.ID,
			PackageSlug:     challenge.PackageSlug,
			Title:           challenge.Title,
			Category:        challenge.Category,
			Difficulty:      challenge.Difficulty,
			Points:          challenge.Points,
			ImageID:         challenge.ImageID,
			Status:          string(challenge.Status),
			FlagType:        challenge.FlagType,
			FlagHash:        challenge.FlagHash,
			FlagSalt:        challenge.FlagSalt,
			FlagRegex:       challenge.FlagRegex,
			FlagPrefix:      challenge.FlagPrefix,
			InstanceSharing: string(challenge.InstanceSharing),
			TargetProtocol:  challenge.TargetProtocol,
			TargetPort:      challenge.TargetPort,
		}, nil
	}
	return nil, nil
}

func (s *stubPracticeChallengeContract) FindPracticeRuntimeChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*challengecontracts.PracticeRuntimeChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDFn != nil {
		topology, err := s.findChallengeTopologyByChallengeIDFn(ctx, challengeID)
		if err != nil || topology == nil {
			return nil, err
		}
		return &challengecontracts.PracticeRuntimeChallengeTopology{
			ChallengeID:  topology.ChallengeID,
			EntryNodeKey: topology.EntryNodeKey,
			Spec:         topology.Spec,
		}, nil
	}
	return nil, nil
}

type stubPracticeImageStore struct {
	findByIDFn func(ctx context.Context, id int64) (*challengecontracts.Image, error)
}

func (s *stubPracticeImageStore) FindByID(ctx context.Context, id int64) (*challengecontracts.Image, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

type stubPracticeInstanceStore struct {
	findByIDWithContextFn                   func(ctx context.Context, id int64) (*instanceentity.Instance, error)
	updateRuntimeWithContextFn              func(ctx context.Context, instance *instanceentity.Instance) error
	bindRuntimeNodeWithContextFn            func(ctx context.Context, id int64, nodeID *int64) (bool, error)
	persistProvisionedRuntimeWithContextFn  func(ctx context.Context, instance *instanceentity.Instance) (bool, error)
	finishActiveAWDServiceOperationFn       func(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error
	refreshInstanceExpiryWithContextFn      func(ctx context.Context, instanceID int64, expiresAt time.Time) error
	updateStatusAndReleasePortWithContextFn func(ctx context.Context, id int64, status string) error
	failProvisioningWithContextFn           func(ctx context.Context, id int64) (bool, error)
	requeueLostRuntimeWithContextFn         func(ctx context.Context, id int64) (bool, error)
	findByUserAndChallengeWithContextFn     func(ctx context.Context, userID, challengeID int64) (*instanceentity.Instance, error)
}

func (s *stubPracticeInstanceStore) FindByID(ctx context.Context, id int64) (*instanceentity.Instance, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubPracticeInstanceStore) UpdateRuntime(ctx context.Context, instance *instanceentity.Instance) error {
	if s.updateRuntimeWithContextFn != nil {
		return s.updateRuntimeWithContextFn(ctx, instance)
	}
	return nil
}

func (s *stubPracticeInstanceStore) BindRuntimeNode(ctx context.Context, id int64, nodeID *int64) (bool, error) {
	if s.bindRuntimeNodeWithContextFn != nil {
		return s.bindRuntimeNodeWithContextFn(ctx, id, nodeID)
	}
	return true, nil
}

func (s *stubPracticeInstanceStore) PersistProvisionedRuntime(ctx context.Context, instance *instanceentity.Instance) (bool, error) {
	if s.persistProvisionedRuntimeWithContextFn != nil {
		return s.persistProvisionedRuntimeWithContextFn(ctx, instance)
	}
	if err := s.UpdateRuntime(ctx, instance); err != nil {
		return false, err
	}
	return true, nil
}

func (s *stubPracticeInstanceStore) FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error {
	if s.finishActiveAWDServiceOperationFn != nil {
		return s.finishActiveAWDServiceOperationFn(ctx, instanceID, status, errorMessage, finishedAt)
	}
	return nil
}

func (s *stubPracticeInstanceStore) RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error {
	if s.refreshInstanceExpiryWithContextFn != nil {
		return s.refreshInstanceExpiryWithContextFn(ctx, instanceID, expiresAt)
	}
	return nil
}

func (s *stubPracticeInstanceStore) UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error {
	if s.updateStatusAndReleasePortWithContextFn != nil {
		return s.updateStatusAndReleasePortWithContextFn(ctx, id, status)
	}
	return nil
}

func (s *stubPracticeInstanceStore) FailProvisioning(ctx context.Context, id int64) (bool, error) {
	if s.failProvisioningWithContextFn != nil {
		return s.failProvisioningWithContextFn(ctx, id)
	}
	if err := s.UpdateStatusAndReleasePort(ctx, id, instanceentity.InstanceStatusFailed); err != nil {
		return false, err
	}
	return true, nil
}

func (s *stubPracticeInstanceStore) RequeueLostRuntime(ctx context.Context, id int64) (bool, error) {
	if s.requeueLostRuntimeWithContextFn != nil {
		return s.requeueLostRuntimeWithContextFn(ctx, id)
	}
	return false, nil
}

func (s *stubPracticeInstanceStore) FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*instanceentity.Instance, error) {
	if s.findByUserAndChallengeWithContextFn != nil {
		return s.findByUserAndChallengeWithContextFn(ctx, userID, challengeID)
	}
	return nil, nil
}

func (s *stubPracticeInstanceStore) ListPendingInstances(ctx context.Context, limit int) ([]*instanceentity.Instance, error) {
	return []*instanceentity.Instance{}, nil
}

func (s *stubPracticeInstanceStore) TryTransitionStatus(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error) {
	return false, nil
}

func (s *stubPracticeInstanceStore) CountInstancesByStatus(ctx context.Context, statuses []string) (int64, error) {
	return 0, nil
}

type interceptAWDDefenseWorkspaceRepository struct {
	*practiceTestInstanceRepository
	upsertFn func(ctx context.Context, workspace *runtimeentity.AWDDefenseWorkspace) error
}

func (r *interceptAWDDefenseWorkspaceRepository) UpsertAWDDefenseWorkspace(ctx context.Context, workspace *runtimeentity.AWDDefenseWorkspace) error {
	if r.upsertFn != nil {
		if err := r.upsertFn(ctx, workspace); err != nil {
			return err
		}
	}
	return r.awdRepo.UpsertAWDDefenseWorkspace(ctx, workspace)
}

type practiceServiceContextKey string
