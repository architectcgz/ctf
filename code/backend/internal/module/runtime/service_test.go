package runtime_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"ctf-platform/internal/config"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	identitycontracts "ctf-platform/internal/module/identity/contracts"
	instancecmd "ctf-platform/internal/module/instance/application/commands"
	instanceqry "ctf-platform/internal/module/instance/application/queries"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceentity "ctf-platform/internal/module/instance/entity"
	instanceports "ctf-platform/internal/module/instance/ports"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimecontracts "ctf-platform/internal/module/runtime/contracts"
	runtimeentity "ctf-platform/internal/module/runtime/entity"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

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
	if err := db.AutoMigrate(&identitycontracts.User{}, &runtimeChallengeTestRow{}, &instanceentity.Instance{}, &runtimeentity.PortAllocation{}, &runtimeentity.NetworkAllocation{}, &contestcontracts.ContestRegistration{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	if err := db.AutoMigrate(&contestcontracts.Team{}, &contestcontracts.TeamMember{}); err != nil {
		t.Fatalf("migrate tables: %v", err)
	}
	if err := db.AutoMigrate(&contestcontracts.Contest{}, &contestcontracts.ContestAWDService{}); err != nil {
		t.Fatalf("migrate awd tables: %v", err)
	}
	if err := db.AutoMigrate(&runtimecontracts.AWDScopeControl{}); err != nil {
		t.Fatalf("migrate awd scope control tables: %v", err)
	}
	if err := db.AutoMigrate(&runtimeentity.AWDServiceOperation{}); err != nil {
		t.Fatalf("migrate awd operation tables: %v", err)
	}
	return &runtimeTestRepository{
		Repository: runtimeinfra.NewRepository(db),
		db:         db,
	}
}

type testRuntimeService struct {
	commands    *instancecmd.InstanceService
	queries     *instanceqry.InstanceService
	maintenance *instancecmd.InstanceMaintenanceService
}

func (s *testRuntimeService) DestroyInstance(ctx context.Context, instanceID, userID int64) error {
	return s.commands.DestroyInstance(ctx, instanceID, userID)
}

func (s *testRuntimeService) ExtendInstance(ctx context.Context, instanceID, userID int64) (*instancecontracts.InstanceResp, error) {
	return s.commands.ExtendInstance(ctx, instanceID, userID)
}

func (s *testRuntimeService) GetUserInstances(ctx context.Context, userID int64) ([]*instancecontracts.InstanceInfo, error) {
	return s.queries.GetUserInstances(ctx, userID)
}

func (s *testRuntimeService) GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error) {
	return s.queries.GetAccessURL(ctx, instanceID, userID)
}

func (s *testRuntimeService) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query instancecontracts.TeacherInstanceListQuery) (*instancecontracts.TeacherInstancePageResult, error) {
	return s.queries.ListTeacherInstances(ctx, requesterID, requesterRole, query)
}

func (s *testRuntimeService) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	return s.commands.DestroyTeacherInstance(ctx, instanceID, requesterID, requesterRole)
}

func (s *testRuntimeService) RunStoppingCleanupLoop(ctx context.Context) {
	if s == nil || s.maintenance == nil {
		return
	}
	s.maintenance.RunStoppingCleanupLoop(ctx)
}

func newTestRuntimeModule(repo *runtimeTestRepository, engine *fakeRuntimeEngine) *testRuntimeService {
	cfg := &config.ContainerConfig{
		MaxExtends:          2,
		ExtendDuration:      30 * time.Minute,
		OrphanGracePeriod:   5 * time.Minute,
		DeletePollInterval:  5 * time.Millisecond,
		DeleteMaxConcurrent: 2,
	}
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, repo, nil)
	instanceCleaner := newRuntimeTestCleanerAdapter(cleanupService)
	return &testRuntimeService{
		commands:    instancecmd.NewInstanceService(repo, instanceCleaner, cfg, nil),
		queries:     instanceqry.NewInstanceService(repo, cfg),
		maintenance: instancecmd.NewInstanceMaintenanceService(newRuntimeTestMaintenanceRepository(repo), nil, instanceCleaner, cfg, nil),
	}
}

type runtimeTestMaintenanceRepository struct {
	repo *runtimeTestRepository
}

func newRuntimeTestMaintenanceRepository(repo *runtimeTestRepository) *runtimeTestMaintenanceRepository {
	return &runtimeTestMaintenanceRepository{repo: repo}
}

func (r *runtimeTestMaintenanceRepository) UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error {
	return r.repo.UpdateStatusAndReleasePort(ctx, id, status)
}

func (r *runtimeTestMaintenanceRepository) FindExpired(ctx context.Context) ([]*instanceentity.Instance, error) {
	return r.repo.FindExpired(ctx)
}

func (r *runtimeTestMaintenanceRepository) ListStoppingInstances(ctx context.Context, updatedBefore time.Time, limit int) ([]*instanceentity.Instance, error) {
	return r.repo.ListStoppingInstances(ctx, updatedBefore, limit)
}

func (r *runtimeTestMaintenanceRepository) ListRecoverableActiveInstances(ctx context.Context) ([]*instanceentity.Instance, error) {
	return r.repo.ListRecoverableActiveInstances(ctx)
}

func (r *runtimeTestMaintenanceRepository) FindRunningAWDDefenseWorkspaceByInstanceID(ctx context.Context, instanceID int64) (*instanceports.AWDDefenseWorkspace, error) {
	workspace, err := r.repo.FindRunningAWDDefenseWorkspaceByInstanceID(ctx, instanceID)
	if err != nil || workspace == nil {
		return nil, err
	}
	return &instanceports.AWDDefenseWorkspace{ContainerID: workspace.ContainerID}, nil
}

func (r *runtimeTestMaintenanceRepository) CreateAWDServiceOperation(ctx context.Context, operation *instanceports.AWDServiceOperation) error {
	if operation == nil {
		return nil
	}
	row := runtimeentity.AWDServiceOperation{
		ID:            operation.ID,
		ContestID:     operation.ContestID,
		TeamID:        operation.TeamID,
		ServiceID:     operation.ServiceID,
		InstanceID:    operation.InstanceID,
		OperationType: operation.OperationType,
		RequestedBy:   operation.RequestedBy,
		RequestedByID: operation.RequestedByID,
		Reason:        operation.Reason,
		SLABillable:   operation.SLABillable,
		Status:        operation.Status,
		ErrorMessage:  operation.ErrorMessage,
		StartedAt:     operation.StartedAt,
		FinishedAt:    operation.FinishedAt,
		CreatedAt:     operation.CreatedAt,
		UpdatedAt:     operation.UpdatedAt,
	}
	if err := r.repo.CreateAWDServiceOperation(ctx, &row); err != nil {
		return err
	}
	operation.ID = row.ID
	return nil
}

func (r *runtimeTestMaintenanceRepository) FinishAWDServiceOperation(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error {
	return r.repo.FinishAWDServiceOperation(ctx, operationID, status, errorMessage, finishedAt)
}

func (r *runtimeTestMaintenanceRepository) FinalizeStoppedRuntime(ctx context.Context, id int64) error {
	return r.repo.FinalizeStoppedRuntime(ctx, id)
}

func (r *runtimeTestMaintenanceRepository) RequeueLostRuntime(ctx context.Context, id int64) (bool, error) {
	return r.repo.RequeueLostRuntime(ctx, id)
}

func (r *runtimeTestMaintenanceRepository) ListActiveContainerIDs(ctx context.Context) ([]string, error) {
	return r.repo.ListActiveContainerIDs(ctx)
}

type fakeRuntimeEngine struct {
	networkID                      string
	networkIDs                     []string
	createNetworkErrs              []error
	listNetworkSubnets             []string
	listNetworkSubnetsErr          error
	listNetworkSubnetsCalls        int
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
	createdNetworkAllowExisting    bool
	createdNetworkAllowExistingSeq []bool
	createdNetworkLabel            map[string]string
	createdNetworkSubnet           string
	createdNetworkSubnets          []string
	createdContainerCfg            *runtimecontracts.ContainerConfig
	createdContainerCfgs           []*runtimecontracts.ContainerConfig
	removedContainerID             string
	removedContainerIDs            []string
	removedNetworkID               string
	removedNetworkIDs              []string
	appliedACLHandle               *runtimecontracts.InstanceRuntimeACLHandle
	removedACLHandle               *runtimecontracts.InstanceRuntimeACLHandle
	appliedACLRules                []runtimecontracts.InstanceRuntimeACLRule
	removedACLRules                []runtimecontracts.InstanceRuntimeACLRule
	connectedNetworks              map[string][]string
	writtenFiles                   map[string]map[string]string
	imageSize                      int64
	imageInspectErr                error
	removedImageRef                string
	managedContainerStats          []runtimeports.ManagedContainerStat
	managedContainerStates         map[string]*runtimeports.ManagedContainerState
	inspectContainerNetworkIPsFunc func(containerID string, engine *fakeRuntimeEngine) map[string]string
	stopContainerFn                func(ctx context.Context, containerID string, timeout time.Duration) error
	removeContainerFn              func(ctx context.Context, containerID string, force bool) error
	removeNetworkFn                func(ctx context.Context, networkID string) error
	removeACLRulesFn               func(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error
}

func (f *fakeRuntimeEngine) CreateNetwork(_ context.Context, name string, labels map[string]string, _ bool, allowExisting bool, subnet string) (string, error) {
	f.createdNetworkName = name
	f.createdNetworkNames = append(f.createdNetworkNames, name)
	f.createdNetworkAllowExisting = allowExisting
	f.createdNetworkAllowExistingSeq = append(f.createdNetworkAllowExistingSeq, allowExisting)
	f.createdNetworkLabel = labels
	f.createdNetworkSubnet = subnet
	f.createdNetworkSubnets = append(f.createdNetworkSubnets, subnet)
	if len(f.createNetworkErrs) > 0 {
		err := f.createNetworkErrs[0]
		f.createNetworkErrs = f.createNetworkErrs[1:]
		if err != nil {
			return "", err
		}
	}
	if len(f.networkIDs) > 0 {
		networkID := f.networkIDs[0]
		f.networkIDs = f.networkIDs[1:]
		return networkID, nil
	}
	return f.networkID, nil
}

func (f *fakeRuntimeEngine) ListNetworkSubnets(_ context.Context) ([]string, error) {
	f.listNetworkSubnetsCalls++
	if f.listNetworkSubnetsErr != nil {
		return nil, f.listNetworkSubnetsErr
	}
	return append([]string(nil), f.listNetworkSubnets...), nil
}

func (f *fakeRuntimeEngine) CreateContainer(_ context.Context, cfg *runtimecontracts.ContainerConfig) (string, error) {
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

func (f *fakeRuntimeEngine) ApplyACLRules(_ context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	if f.applyACLErr != nil {
		return f.applyACLErr
	}
	f.appliedACLRules = append(f.appliedACLRules, rules...)
	return nil
}

func (f *fakeRuntimeEngine) RemoveACLRules(ctx context.Context, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	if f.removeACLRulesFn != nil {
		return f.removeACLRulesFn(ctx, rules)
	}
	f.removedACLRules = append(f.removedACLRules, rules...)
	return nil
}

func (f *fakeRuntimeEngine) ApplyACL(_ context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle, rules []runtimecontracts.InstanceRuntimeACLRule) error {
	if f.applyACLErr != nil {
		return f.applyACLErr
	}
	f.appliedACLHandle = handle
	f.appliedACLRules = append(f.appliedACLRules, rules...)
	return nil
}

func (f *fakeRuntimeEngine) RemoveACL(_ context.Context, handle *runtimecontracts.InstanceRuntimeACLHandle) error {
	f.removedACLHandle = handle
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

func (f *fakeRuntimeEngine) InspectManagedContainer(_ context.Context, containerID string) (*runtimeports.ManagedContainerState, error) {
	if f.managedContainerStates == nil {
		return &runtimeports.ManagedContainerState{ID: containerID, Exists: true, Running: true, Status: "running"}, nil
	}
	if state, exists := f.managedContainerStates[containerID]; exists {
		return state, nil
	}
	return &runtimeports.ManagedContainerState{ID: containerID, Exists: false}, nil
}

func seedInstance(t *testing.T, db *gorm.DB, instance *instanceentity.Instance) {
	t.Helper()

	if err := db.Create(instance).Error; err != nil {
		t.Fatalf("seed instance: %v", err)
	}
}

func seedAWDDefenseWorkspace(t *testing.T, db *gorm.DB, workspace *runtimeentity.AWDDefenseWorkspace) {
	t.Helper()

	if err := db.Create(workspace).Error; err != nil {
		t.Fatalf("seed awd defense workspace: %v", err)
	}
}

func seedUser(t *testing.T, db *gorm.DB, user *identitycontracts.User) {
	t.Helper()

	if err := db.Create(user).Error; err != nil {
		t.Fatalf("seed user: %v", err)
	}
}

func seedChallenge(t *testing.T, db *gorm.DB, challenge *runtimeChallengeTestRow) {
	t.Helper()

	if err := db.Create(challenge).Error; err != nil {
		t.Fatalf("seed challenge: %v", err)
	}
}
