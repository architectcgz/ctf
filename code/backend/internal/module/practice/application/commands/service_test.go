package commands

import (
	"context"
	"ctf-platform/internal/config"
	"ctf-platform/internal/model"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeinfra "ctf-platform/internal/module/challenge/infrastructure"
	practiceinfra "ctf-platform/internal/module/practice/infrastructure"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
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
	service *Service,
	repo practiceports.PracticeContestScopeRepository,
	challengeRepo challengecontracts.PracticeChallengeContract,
) *Service {
	if service == nil {
		return nil
	}
	return service.
		SetContestScopeRepository(practiceinfra.NewContestScopeRepository(repo)).
		SetRuntimeSubjectRepository(practiceinfra.NewRuntimeSubjectRepository(challengeRepo))
}

func wirePracticeManualReviewAdapters(
	service *Service,
	repo practiceports.PracticeManualReviewRepository,
	challengeRepo challengecontracts.PracticeChallengeContract,
) *Service {
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
	service *Service,
	repo practiceports.PracticeSolvedSubmissionRepository,
	challengeRepo challengecontracts.PracticeChallengeContract,
) *Service {
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

type stubPracticeRuntimeService struct {
	cleanupRuntimeFn          func(ctx context.Context, instance *model.Instance) error
	createTopologyFn          func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error)
	createContainerFn         func(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error)
	inspectManagedContainerFn func(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error)
}

func (s *stubPracticeRuntimeService) CleanupRuntime(ctx context.Context, instance *model.Instance) error {
	if s.cleanupRuntimeFn == nil {
		return nil
	}
	return s.cleanupRuntimeFn(ctx, instance)
}

func (s *stubPracticeRuntimeService) CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if s.createTopologyFn == nil {
		return nil, errors.New("unexpected CreateTopology call")
	}
	return s.createTopologyFn(ctx, req)
}

func (s *stubPracticeRuntimeService) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (string, string, int, int, error) {
	if s.createContainerFn == nil {
		return "", "", 0, 0, errors.New("unexpected CreateContainer call")
	}
	return s.createContainerFn(ctx, imageName, env, reservedHostPort)
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

	service := &Service{
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

	_, err := service.createAWDDefenseWorkspaceCompanion(context.Background(), &model.Instance{
		ContestID: &contestID,
		TeamID:    &teamID,
		ServiceID: &serviceID,
	}, &awdDefenseWorkspacePlan{
		workspaceRevision:      2,
		workspaceContainerName: "ctf-workspace-custom",
		workspaceMounts: []model.ContainerMount{
			{Source: "ws-app", Target: "/workspace/app"},
			{Source: "ws-templates", Target: "/workspace/templates"},
			{Source: "ws-data", Target: "/workspace/data", ReadOnly: true},
		},
	})
	if err != nil {
		t.Fatalf("createAWDDefenseWorkspaceCompanion() error = %v", err)
	}
}

func TestParseAWDDefenseWorkspaceConfigTreatsRootsOutsideWritableSetAsReadonly(t *testing.T) {
	config, err := parseAWDDefenseWorkspaceConfig(map[string]any{
		"defense_workspace": map[string]any{
			"seed_root":       "docker/workspace",
			"workspace_roots": []string{"docker/workspace/src", "docker/workspace/templates", "docker/workspace/data"},
			"writable_roots":  []string{"docker/workspace/src"},
			"readonly_roots":  []string{"docker/workspace/data"},
			"runtime_mounts": []any{
				map[string]any{"source": "docker/workspace/src", "target": "/workspace/src", "mode": "rw"},
				map[string]any{"source": "docker/workspace/templates", "target": "/workspace/templates", "mode": "ro"},
				map[string]any{"source": "docker/workspace/data", "target": "/workspace/data", "mode": "ro"},
			},
		},
	})
	if err != nil {
		t.Fatalf("parseAWDDefenseWorkspaceConfig() error = %v", err)
	}
	if len(config.workspaceRoots) != 3 {
		t.Fatalf("expected three workspace roots, got %+v", config.workspaceRoots)
	}

	readonlyBySource := make(map[string]bool, len(config.workspaceRoots))
	for _, root := range config.workspaceRoots {
		readonlyBySource[root.source] = root.readOnly
	}
	if readonlyBySource["docker/workspace/src"] {
		t.Fatalf("expected src root to stay writable, got %+v", config.workspaceRoots)
	}
	if !readonlyBySource["docker/workspace/templates"] {
		t.Fatalf("expected template root outside writable_roots to default readonly, got %+v", config.workspaceRoots)
	}
	if !readonlyBySource["docker/workspace/data"] {
		t.Fatalf("expected readonly root to stay readonly, got %+v", config.workspaceRoots)
	}
}

func TestBuildAWDDefenseWorkspaceBootstrapCommandDegradesGracefullyWithoutPackageInstall(t *testing.T) {
	command := buildAWDDefenseWorkspaceBootstrapCommand([]model.ContainerMount{
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
	if err := db.Create(&model.Image{
		ID:        601,
		Name:      "ctf/awd-web",
		Tag:       "v1",
		Digest:    "sha256:awd-web-v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(801)
	teamID := int64(802)
	serviceID := int64(803)
	challengeID := int64(804)
	if err := db.Create(&model.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        9001,
		WorkspaceRevision: 1,
		Status:            model.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-stale-ctr",
		SeedSignature:     "seed-signature",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create workspace state: %v", err)
	}

	serviceSnapshot, err := model.EncodeContestAWDServiceSnapshot(model.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":         601,
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

	cleanupCalls := 0
	createTopologyCalls := 0
	service := &Service{
		repo: &stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*model.ContestAWDService, error) {
				return &model.ContestAWDService{
					ID:              gotServiceID,
					ContestID:       gotContestID,
					AWDChallengeID:  challengeID,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
		},
		imageRepo:    challengeinfra.NewImageRepository(db),
		instanceRepo: runtimeinfrarepo.NewRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, instance *model.Instance) error {
				cleanupCalls++
				details, err := model.DecodeInstanceRuntimeDetails(instance.RuntimeDetails)
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
						RuntimeDetails: model.InstanceRuntimeDetails{
							Networks: []model.InstanceRuntimeNetwork{
								{
									Key:       model.TopologyDefaultNetworkKey,
									Name:      "ctf-awd-contest-801",
									NetworkID: "net-awd-contest-801",
									Shared:    true,
								},
							},
							Containers: []model.InstanceRuntimeContainer{
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

	instance := &model.Instance{
		ID:          9001,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
	}
	if err := service.createSingleContainer(context.Background(), instance, &model.Challenge{
		ID:             challengeID,
		ImageID:        601,
		TargetPort:     8080,
		TargetProtocol: model.ChallengeTargetProtocolHTTP,
	}, "flag{demo}"); err != nil {
		t.Fatalf("createSingleContainer() error = %v", err)
	}

	if cleanupCalls != 1 {
		t.Fatalf("expected one stale workspace cleanup call, got %d", cleanupCalls)
	}
	if createTopologyCalls != 2 {
		t.Fatalf("expected runtime and workspace topology creation, got %d", createTopologyCalls)
	}
}

func TestCreateSingleAWDContainerPreservesStaleWorkspaceReferenceWhenCleanupFails(t *testing.T) {
	db := newPracticeCommandTestDB(t)
	now := time.Now()
	if err := db.Create(&model.Image{
		ID:        602,
		Name:      "ctf/awd-web",
		Tag:       "v1",
		Digest:    "sha256:awd-web-v1",
		Status:    model.ImageStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}).Error; err != nil {
		t.Fatalf("create image: %v", err)
	}

	contestID := int64(811)
	teamID := int64(812)
	serviceID := int64(813)
	challengeID := int64(814)
	if err := db.Create(&model.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        9011,
		WorkspaceRevision: 1,
		Status:            model.AWDDefenseWorkspaceStatusRunning,
		ContainerID:       "workspace-stale-ctr",
		SeedSignature:     "seed-signature",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create workspace state: %v", err)
	}

	serviceSnapshot, err := model.EncodeContestAWDServiceSnapshot(model.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":         602,
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

	service := &Service{
		repo: &stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*model.ContestAWDService, error) {
				return &model.ContestAWDService{
					ID:              gotServiceID,
					ContestID:       gotContestID,
					AWDChallengeID:  challengeID,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
		},
		imageRepo:    challengeinfra.NewImageRepository(db),
		instanceRepo: runtimeinfrarepo.NewRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			cleanupRuntimeFn: func(ctx context.Context, instance *model.Instance) error {
				return fmt.Errorf("cleanup stale workspace failed")
			},
			createTopologyFn: func(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
				return &practiceports.TopologyCreateResult{
					PrimaryContainerID: "awd-private-ctr",
					NetworkID:          "net-awd-contest-811",
					AccessURL:          "http://awd-c811-t812-s813:8080",
					RuntimeDetails: model.InstanceRuntimeDetails{
						Networks: []model.InstanceRuntimeNetwork{
							{
								Key:       model.TopologyDefaultNetworkKey,
								Name:      "ctf-awd-contest-811",
								NetworkID: "net-awd-contest-811",
								Shared:    true,
							},
						},
						Containers: []model.InstanceRuntimeContainer{
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

	instance := &model.Instance{
		ID:          9011,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
	}
	err = service.createSingleContainer(context.Background(), instance, &model.Challenge{
		ID:             challengeID,
		ImageID:        602,
		TargetPort:     8080,
		TargetProtocol: model.ChallengeTargetProtocolHTTP,
	}, "flag{demo}")
	if err == nil {
		t.Fatal("expected createSingleContainer() to fail when stale workspace cleanup fails")
	}

	workspace, err := runtimeinfrarepo.NewRepository(db).FindAWDDefenseWorkspace(context.Background(), contestID, teamID, serviceID)
	if err != nil {
		t.Fatalf("FindAWDDefenseWorkspace() error = %v", err)
	}
	if workspace == nil {
		t.Fatal("expected workspace row to exist")
	}
	if workspace.Status != model.AWDDefenseWorkspaceStatusFailed {
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
	if err := db.Create(&model.AWDDefenseWorkspace{
		ContestID:         contestID,
		TeamID:            teamID,
		ServiceID:         serviceID,
		InstanceID:        9021,
		WorkspaceRevision: 2,
		Status:            model.AWDDefenseWorkspaceStatusFailed,
		ContainerID:       "workspace-stale-ctr",
		SeedSignature:     "seed-signature",
		CreatedAt:         now,
		UpdatedAt:         now,
	}).Error; err != nil {
		t.Fatalf("create workspace state: %v", err)
	}

	serviceSnapshot, err := model.EncodeContestAWDServiceSnapshot(model.ContestAWDServiceSnapshot{
		Name: "AWD Service",
		RuntimeConfig: map[string]any{
			"image_id":         challengeID,
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

	service := &Service{
		repo: &stubPracticeRepository{
			findContestAWDServiceFn: func(ctx context.Context, gotContestID, gotServiceID int64) (*model.ContestAWDService, error) {
				return &model.ContestAWDService{
					ID:              gotServiceID,
					ContestID:       gotContestID,
					AWDChallengeID:  challengeID,
					IsVisible:       true,
					ServiceSnapshot: serviceSnapshot,
				}, nil
			},
		},
		instanceRepo: runtimeinfrarepo.NewRepository(db),
		runtimeService: &stubPracticeRuntimeService{
			inspectManagedContainerFn: func(ctx context.Context, containerID string) (*practiceports.ManagedContainerState, error) {
				t.Fatalf("unexpected managed container inspection for failed workspace state: %s", containerID)
				return nil, nil
			},
		},
		config: &config.Config{},
	}

	plan, err := service.prepareAWDDefenseWorkspacePlan(context.Background(), &model.Instance{
		ID:          9021,
		ContestID:   &contestID,
		TeamID:      &teamID,
		ChallengeID: challengeID,
		ServiceID:   &serviceID,
	}, &model.Challenge{ID: challengeID})
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
		&model.Image{},
		&model.Challenge{},
		&model.ChallengeTopology{},
		&model.Contest{},
		&model.ContestAWDService{},
		&model.ContestRegistration{},
		&model.User{},
		&model.Team{},
		&model.Instance{},
		&model.AWDServiceOperation{},
		&model.AWDDefenseWorkspace{},
		&model.PortAllocation{},
		&model.Submission{},
	); err != nil {
		t.Fatalf("migrate practice command tables: %v", err)
	}
	return db
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
	findByIDWithContextFn                func(ctx context.Context, id int64) (*model.Challenge, error)
	findChallengeTopologyByChallengeIDFn func(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error)
}

func (s *stubPracticeChallengeContract) FindByID(ctx context.Context, id int64) (*model.Challenge, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubPracticeChallengeContract) FindChallengeTopologyByChallengeID(ctx context.Context, challengeID int64) (*model.ChallengeTopology, error) {
	if s.findChallengeTopologyByChallengeIDFn != nil {
		return s.findChallengeTopologyByChallengeIDFn(ctx, challengeID)
	}
	return nil, nil
}

type stubPracticeImageStore struct {
	findByIDFn func(ctx context.Context, id int64) (*model.Image, error)
}

func (s *stubPracticeImageStore) FindByID(ctx context.Context, id int64) (*model.Image, error) {
	if s.findByIDFn != nil {
		return s.findByIDFn(ctx, id)
	}
	return nil, nil
}

type stubPracticeInstanceStore struct {
	findByIDWithContextFn                   func(ctx context.Context, id int64) (*model.Instance, error)
	updateRuntimeWithContextFn              func(ctx context.Context, instance *model.Instance) error
	finishActiveAWDServiceOperationFn       func(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error
	refreshInstanceExpiryWithContextFn      func(ctx context.Context, instanceID int64, expiresAt time.Time) error
	updateStatusAndReleasePortWithContextFn func(ctx context.Context, id int64, status string) error
	findByUserAndChallengeWithContextFn     func(ctx context.Context, userID, challengeID int64) (*model.Instance, error)
}

func (s *stubPracticeInstanceStore) FindByID(ctx context.Context, id int64) (*model.Instance, error) {
	if s.findByIDWithContextFn != nil {
		return s.findByIDWithContextFn(ctx, id)
	}
	return nil, nil
}

func (s *stubPracticeInstanceStore) UpdateRuntime(ctx context.Context, instance *model.Instance) error {
	if s.updateRuntimeWithContextFn != nil {
		return s.updateRuntimeWithContextFn(ctx, instance)
	}
	return nil
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

func (s *stubPracticeInstanceStore) FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*model.Instance, error) {
	if s.findByUserAndChallengeWithContextFn != nil {
		return s.findByUserAndChallengeWithContextFn(ctx, userID, challengeID)
	}
	return nil, nil
}

func (s *stubPracticeInstanceStore) ListPendingInstances(ctx context.Context, limit int) ([]*model.Instance, error) {
	return []*model.Instance{}, nil
}

func (s *stubPracticeInstanceStore) TryTransitionStatus(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error) {
	return false, nil
}

func (s *stubPracticeInstanceStore) CountInstancesByStatus(ctx context.Context, statuses []string) (int64, error) {
	return 0, nil
}

type interceptAWDDefenseWorkspaceRepository struct {
	*runtimeinfrarepo.Repository
	upsertFn func(ctx context.Context, workspace *model.AWDDefenseWorkspace) error
}

func (r *interceptAWDDefenseWorkspaceRepository) UpsertAWDDefenseWorkspace(ctx context.Context, workspace *model.AWDDefenseWorkspace) error {
	if r.upsertFn != nil {
		if err := r.upsertFn(ctx, workspace); err != nil {
			return err
		}
	}
	return r.Repository.UpsertAWDDefenseWorkspace(ctx, workspace)
}

type practiceServiceContextKey string
