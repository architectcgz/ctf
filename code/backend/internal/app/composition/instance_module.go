package composition

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/auditlog"
	runtimecmd "ctf-platform/internal/module/container_runtime/application/commands"
	containerruntimeinfra "ctf-platform/internal/module/container_runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/container_runtime/ports"
	contestentity "ctf-platform/internal/module/contest/entity"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	instancehttp "ctf-platform/internal/module/instance/api/http"
	instancecmd "ctf-platform/internal/module/instance/application/commands"
	instanceqry "ctf-platform/internal/module/instance/application/queries"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceinfra "ctf-platform/internal/module/instance/infrastructure"
	instanceports "ctf-platform/internal/module/instance/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
)

type InstanceModule struct {
	Handler *instancehttp.Handler

	PracticeInstanceRepository interface {
		FindByID(ctx context.Context, id int64) (*instancecontracts.Instance, error)
		FailProvisioning(ctx context.Context, id int64) (bool, error)
		UpdateRuntime(ctx context.Context, instance *instancecontracts.Instance) error
		PersistProvisionedRuntime(ctx context.Context, instance *instancecontracts.Instance) (bool, error)
		FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error
		RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error
		UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error
		FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*instancecontracts.Instance, error)
		ListPendingInstances(ctx context.Context, limit int) ([]*instancecontracts.Instance, error)
		TryTransitionStatus(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error)
		CountInstancesByStatus(ctx context.Context, statuses []string) (int64, error)
	}
	PracticeRuntimeService      practiceports.RuntimeInstanceService
	PracticeRuntimeNodeSelector practiceports.RuntimeNodeSelector

	service              *runtimeHTTPServiceAdapter
	proxyTrafficRecorder runtimeProxyTrafficRecorder
	startupRecovery      *instancecmd.StartupRuntimeRecoveryService
}

type runtimeProxyTrafficRecorder interface {
	RecordRuntimeProxyTrafficEvent(ctx context.Context, instanceID, userID int64, method, requestPath string, statusCode int) error
	RecordAWDProxyTrafficEvent(ctx context.Context, event instanceports.AWDProxyTrafficEventInput) error
}

func BuildInstanceModule(root *Root, runtime *ContainerRuntimeModule) *InstanceModule {
	if root == nil || runtime == nil || runtime.runtime == nil {
		return &InstanceModule{}
	}

	module := runtime.runtime
	cfg := root.Config()
	log := root.Logger()
	if log == nil {
		log = zap.NewNop()
	}

	inventoryRepo := newCompositeActiveContainerInventory(
		newInstanceRuntimeInventoryProvider(instanceinfra.NewContainerInventoryRepository(root.DB())),
		contestinfra.NewAWDContainerInventoryRepository(root.DB()),
	)
	allocationRepo := containerruntimeinfra.NewAllocationRepository(root.DB())
	awdRepo := contestinfra.NewAWDRepository(root.DB())
	instanceRepo := instanceinfra.NewRepository(root.DB())
	proxyTicketReader := newInstanceProxyTicketReader(instanceRepo, awdRepo)
	defaultCleanupService := module.CleanupService
	var cleanupService interface {
		instanceports.RuntimeCleaner
		RemoveContainer(ctx context.Context, containerID string) error
	} = newInstanceRuntimeCleanupAdapter(defaultCleanupService)
	provisioningService := module.ProvisioningService
	var maintenanceRuntime interface {
		ListManagedContainers(ctx context.Context) ([]instanceports.ManagedContainer, error)
		InspectManagedContainer(ctx context.Context, containerID string) (*instanceports.ManagedContainerState, error)
		StartContainer(ctx context.Context, containerID string) error
	} = newInstanceMaintenanceRuntime(module.ManagedContainerInventory, module.ProvisioningRuntime)
	practiceRuntimeService := newPracticeRuntimeServiceAdapter(defaultCleanupService, provisioningService, module.ManagedContainerInventory)
	if runtime.nodeRouter != nil {
		cleanupService = newInstanceRuntimeCleanupRouterAdapter(runtime.nodeRouter)
		maintenanceRuntime = newInstanceMaintenanceRuntime(runtime.nodeRouter, runtime.nodeRouter)
		practiceRuntimeService = newNodeScopedPracticeRuntimeServiceAdapter(runtime.nodeRouter)
	}
	commandService := instancecmd.NewInstanceService(instanceRepo, cleanupService, &cfg.Container, log.Named("instance_service")).SetEventBus(root.Events)
	queryService := instanceqry.NewInstanceService(instanceRepo, &cfg.Container, cfg.Pagination)
	proxyTicketService := buildRuntimeProxyTicketService(root, proxyTicketReader)
	maintenanceService := instancecmd.NewInstanceMaintenanceService(
		newInstanceMaintenanceRepository(root.DB(), instanceRepo, allocationRepo, awdRepo, inventoryRepo),
		maintenanceRuntime,
		cleanupService,
		&cfg.Container,
		log.Named("instance_maintenance_service"),
		instanceinfra.NewStoppingCleanupLockStore(root.Cache(), cfg.Container.CleanupLockTTL, log.Named("instance_stopping_cleanup_lock")),
	)
	maintenanceService.RegisterStoppingCleanupWakeup(root.Events)
	startupRecovery := instancecmd.NewStartupRuntimeRecoveryService(
		maintenanceService,
		newStartupRuntimeContestRepository(contestinfra.NewRepository(root.DB())),
		instanceRepo,
		instanceinfra.NewPlatformRuntimeStateStore(root.Cache()),
		instanceinfra.NewHostBootIDReader(""),
		0,
		log.Named("startup_runtime_recovery"),
	).SetLockTTL(cfg.Container.StartupRecoveryLockTTL)
	root.RegisterBackgroundJob(NewBackgroundJob(
		"startup_runtime_recovery",
		startupRecovery.Start,
		startupRecovery.Stop,
	))
	cleaner := instanceinfra.NewCleaner(
		maintenanceService,
		root.Cache(),
		cfg.Container.CleanupLockTTL,
		log.Named("runtime_cleaner"),
	)
	root.RegisterBackgroundJob(NewBackgroundJob(
		"runtime_cleaner",
		func(ctx context.Context) error {
			return cleaner.Start(ctx, cfg.Container.CleanupInterval)
		},
		cleaner.Stop,
	))
	root.RegisterBackgroundJob(NewLoopBackgroundJob(
		"instance_stopping_cleanup",
		maintenanceService.RunStoppingCleanupLoop,
	))

	return &InstanceModule{
		PracticeInstanceRepository:  newPracticeInstanceRepository(root.DB(), instanceRepo, allocationRepo, awdRepo),
		PracticeRuntimeService:      practiceRuntimeService,
		PracticeRuntimeNodeSelector: newPracticeRuntimeNodeSelectorAdapter(runtime.RuntimeNodeSelector),
		service: newRuntimeHTTPServiceAdapter(
			commandService,
			queryService,
			proxyTicketService,
			cfg.Container.ProxyBodyPreviewSize,
			int(cfg.Container.ProxyTicketTTL.Seconds()),
			cfg.Container.DefenseSSHEnabled && module.InteractiveExecutor != nil,
			cfg.Container.DefenseSSHHost,
			cfg.Container.DefenseSSHPort,
		),
		startupRecovery:      startupRecovery,
		proxyTrafficRecorder: newInstanceProxyTrafficRecorder(awdRepo),
	}
}

func (m *InstanceModule) SetAWDDesiredRuntimeReconciler(reconciler interface {
	ReconcileDesiredAWDInstances(ctx context.Context) error
}) {
	if m == nil || m.startupRecovery == nil || reconciler == nil {
		return
	}
	m.startupRecovery.SetDesiredRuntimeReconciler(reconciler)
}

func (m *InstanceModule) BuildHandler(root *Root, ops *OpsModule) {
	if m == nil || root == nil || m.service == nil {
		return
	}

	cfg := root.Config()
	var auditRecorder auditlog.Recorder
	if ops != nil {
		auditRecorder = ops.AuditService
	}
	m.Handler = instancehttp.NewHandler(m.service, cfg.Container.PublicHost, cfg.Container.AccessHost, auditRecorder, instancehttp.CookieConfig{
		Secure:   cfg.Auth.SessionCookieSecure,
		SameSite: cfg.Auth.CookieSameSite(),
	}, m.proxyTrafficRecorder)
}

type instanceRuntimeCleanupAdapter struct {
	cleaner *runtimecmd.RuntimeCleanupService
}

func newInstanceRuntimeCleanupAdapter(cleaner *runtimecmd.RuntimeCleanupService) *instanceRuntimeCleanupAdapter {
	if cleaner == nil {
		return nil
	}
	return &instanceRuntimeCleanupAdapter{cleaner: cleaner}
}

func (a *instanceRuntimeCleanupAdapter) CleanupRuntime(ctx context.Context, instance *instancecontracts.Instance) error {
	if a == nil || a.cleaner == nil {
		return nil
	}
	return a.cleaner.CleanupRuntime(ctx, runtimeCleanupTargetFromInstance(instance))
}

func (a *instanceRuntimeCleanupAdapter) RemoveContainer(ctx context.Context, containerID string) error {
	if a == nil || a.cleaner == nil {
		return nil
	}
	return a.cleaner.RemoveContainer(ctx, containerID)
}

type instanceRuntimeCleanupRouterAdapter struct {
	router *runtimeNodeExecutionRouter
}

func newInstanceRuntimeCleanupRouterAdapter(router *runtimeNodeExecutionRouter) *instanceRuntimeCleanupRouterAdapter {
	if router == nil {
		return nil
	}
	return &instanceRuntimeCleanupRouterAdapter{router: router}
}

func (a *instanceRuntimeCleanupRouterAdapter) CleanupRuntime(ctx context.Context, instance *instancecontracts.Instance) error {
	if a == nil || a.router == nil {
		return nil
	}
	return a.router.CleanupRuntime(ctx, runtimeCleanupTargetFromInstance(instance))
}

func (a *instanceRuntimeCleanupRouterAdapter) RemoveContainer(ctx context.Context, containerID string) error {
	if a == nil || a.router == nil {
		return nil
	}
	return a.router.RemoveContainer(ctx, containerID)
}

type instanceMaintenanceRuntimeAdapter struct {
	inventory    runtimeports.ManagedContainerInventory
	provisioning runtimeContainerStarter
}

type runtimeContainerStarter interface {
	StartContainer(ctx context.Context, containerID string) error
}

func newInstanceMaintenanceRuntime(inventory runtimeports.ManagedContainerInventory, provisioning runtimeContainerStarter) *instanceMaintenanceRuntimeAdapter {
	if inventory == nil || provisioning == nil {
		return nil
	}
	return &instanceMaintenanceRuntimeAdapter{
		inventory:    inventory,
		provisioning: provisioning,
	}
}

func (a *instanceMaintenanceRuntimeAdapter) ListManagedContainers(ctx context.Context) ([]instanceports.ManagedContainer, error) {
	if a == nil || a.inventory == nil {
		return nil, nil
	}
	containers, err := a.inventory.ListManagedContainers(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]instanceports.ManagedContainer, len(containers))
	for idx, container := range containers {
		result[idx] = instanceports.ManagedContainer{
			ID:        container.ID,
			Name:      container.Name,
			CreatedAt: container.CreatedAt,
		}
	}
	return result, nil
}

func (a *instanceMaintenanceRuntimeAdapter) InspectManagedContainer(ctx context.Context, containerID string) (*instanceports.ManagedContainerState, error) {
	if a == nil || a.inventory == nil {
		return nil, nil
	}
	state, err := a.inventory.InspectManagedContainer(ctx, containerID)
	if err != nil || state == nil {
		return nil, err
	}
	return &instanceports.ManagedContainerState{
		ID:      state.ID,
		Exists:  state.Exists,
		Running: state.Running,
		Status:  state.Status,
	}, nil
}

func (a *instanceMaintenanceRuntimeAdapter) StartContainer(ctx context.Context, containerID string) error {
	if a == nil || a.provisioning == nil {
		return nil
	}
	return a.provisioning.StartContainer(ctx, containerID)
}

type instanceMaintenanceRepositoryAdapter struct {
	db             *gorm.DB
	instanceRepo   *instanceinfra.Repository
	allocationRepo *containerruntimeinfra.AllocationRepository
	awdRepo        *contestinfra.AWDRepository
	inventoryRepo  activeContainerInventoryProvider
}

func newInstanceMaintenanceRepository(db *gorm.DB, instanceRepo *instanceinfra.Repository, allocationRepo *containerruntimeinfra.AllocationRepository, awdRepo *contestinfra.AWDRepository, inventoryRepo activeContainerInventoryProvider) *instanceMaintenanceRepositoryAdapter {
	if instanceRepo == nil && allocationRepo == nil && awdRepo == nil && inventoryRepo == nil {
		return nil
	}
	return &instanceMaintenanceRepositoryAdapter{
		db:             db,
		instanceRepo:   instanceRepo,
		allocationRepo: allocationRepo,
		awdRepo:        awdRepo,
		inventoryRepo:  inventoryRepo,
	}
}

func (a *instanceMaintenanceRepositoryAdapter) UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error {
	if a == nil || a.db == nil || a.instanceRepo == nil || a.allocationRepo == nil {
		return nil
	}
	return withInstanceRuntimeLifecycleTx(ctx, a.db, a.instanceRepo, a.allocationRepo, func(instanceTx *instanceinfra.Repository, allocationTx *containerruntimeinfra.AllocationRepository) error {
		release, err := instanceTx.UpdateStatus(ctx, id, status)
		if err != nil || release == nil {
			return err
		}
		return allocationTx.ReleaseRuntimeAllocationsForInstance(ctx, release.InstanceID, release.HostPort)
	})
}

func (a *instanceMaintenanceRepositoryAdapter) FindExpired(ctx context.Context) ([]*instancecontracts.Instance, error) {
	return a.instanceRepo.FindExpired(ctx)
}

func (a *instanceMaintenanceRepositoryAdapter) ListStoppingInstances(ctx context.Context, updatedBefore time.Time, limit int) ([]*instancecontracts.Instance, error) {
	return a.instanceRepo.ListStoppingInstances(ctx, updatedBefore, limit)
}

func (a *instanceMaintenanceRepositoryAdapter) ListRecoverableActiveInstances(ctx context.Context) ([]*instancecontracts.Instance, error) {
	return a.instanceRepo.ListRecoverableActiveInstances(ctx)
}

func (a *instanceMaintenanceRepositoryAdapter) FindRunningAWDDefenseWorkspaceByInstanceID(ctx context.Context, instanceID int64) (*instanceports.AWDDefenseWorkspace, error) {
	if a == nil || a.awdRepo == nil {
		return nil, nil
	}
	workspace, err := a.awdRepo.FindRunningAWDDefenseWorkspaceByInstanceID(ctx, instanceID)
	if err != nil || workspace == nil {
		return nil, err
	}
	return &instanceports.AWDDefenseWorkspace{
		ContainerID: workspace.ContainerID,
	}, nil
}

func (a *instanceMaintenanceRepositoryAdapter) CreateAWDServiceOperation(ctx context.Context, operation *instanceports.AWDServiceOperation) error {
	if a == nil || operation == nil {
		return nil
	}
	row := contestentity.AWDServiceOperation{
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
	if a.awdRepo == nil {
		return nil
	}
	if err := a.awdRepo.CreateAWDServiceOperation(ctx, &row); err != nil {
		return err
	}
	operation.ID = row.ID
	return nil
}

func (a *instanceMaintenanceRepositoryAdapter) FinishAWDServiceOperation(ctx context.Context, operationID int64, status, errorMessage string, finishedAt time.Time) error {
	if a == nil || a.awdRepo == nil {
		return nil
	}
	return a.awdRepo.FinishAWDServiceOperation(ctx, operationID, status, errorMessage, finishedAt)
}

func (a *instanceMaintenanceRepositoryAdapter) FinalizeStoppedRuntime(ctx context.Context, id int64) error {
	if a == nil || a.db == nil || a.instanceRepo == nil || a.allocationRepo == nil {
		return nil
	}
	return withInstanceRuntimeLifecycleTx(ctx, a.db, a.instanceRepo, a.allocationRepo, func(instanceTx *instanceinfra.Repository, allocationTx *containerruntimeinfra.AllocationRepository) error {
		release, err := instanceTx.FinalizeStoppedRuntime(ctx, id)
		if err != nil || release == nil {
			return err
		}
		return allocationTx.ReleaseRuntimeAllocationsForInstance(ctx, release.InstanceID, release.HostPort)
	})
}

func (a *instanceMaintenanceRepositoryAdapter) RequeueLostRuntime(ctx context.Context, id int64) (bool, error) {
	return a.instanceRepo.RequeueLostRuntime(ctx, id)
}

func (a *instanceMaintenanceRepositoryAdapter) ListActiveContainerIDs(ctx context.Context) ([]string, error) {
	if a == nil || a.inventoryRepo == nil {
		return nil, nil
	}
	return a.inventoryRepo.ListActiveContainerIDs(ctx)
}

type practiceInstanceRepositoryAdapter struct {
	db             *gorm.DB
	instanceRepo   *instanceinfra.Repository
	allocationRepo *containerruntimeinfra.AllocationRepository
	awdRepo        *contestinfra.AWDRepository
}

func newPracticeInstanceRepository(db *gorm.DB, instanceRepo *instanceinfra.Repository, allocationRepo *containerruntimeinfra.AllocationRepository, awdRepo *contestinfra.AWDRepository) *practiceInstanceRepositoryAdapter {
	if instanceRepo == nil && allocationRepo == nil && awdRepo == nil {
		return nil
	}
	return &practiceInstanceRepositoryAdapter{
		db:             db,
		instanceRepo:   instanceRepo,
		allocationRepo: allocationRepo,
		awdRepo:        awdRepo,
	}
}

func (a *practiceInstanceRepositoryAdapter) FindByID(ctx context.Context, id int64) (*instancecontracts.Instance, error) {
	if a == nil || a.instanceRepo == nil {
		return nil, nil
	}
	return a.instanceRepo.FindByID(ctx, id)
}

func (a *practiceInstanceRepositoryAdapter) FailProvisioning(ctx context.Context, id int64) (bool, error) {
	if a == nil || a.db == nil || a.instanceRepo == nil || a.allocationRepo == nil {
		return false, nil
	}
	changed := false
	err := withInstanceRuntimeLifecycleTx(ctx, a.db, a.instanceRepo, a.allocationRepo, func(instanceTx *instanceinfra.Repository, allocationTx *containerruntimeinfra.AllocationRepository) error {
		release, failed, err := instanceTx.FailProvisioning(ctx, id)
		if err != nil {
			return err
		}
		changed = failed
		if !failed || release == nil {
			return nil
		}
		return allocationTx.ReleaseRuntimeAllocationsForInstance(ctx, release.InstanceID, release.HostPort)
	})
	return changed, err
}

func (a *practiceInstanceRepositoryAdapter) UpdateRuntime(ctx context.Context, instance *instancecontracts.Instance) error {
	if a == nil || a.instanceRepo == nil {
		return nil
	}
	return a.instanceRepo.UpdateRuntime(ctx, instance)
}

func (a *practiceInstanceRepositoryAdapter) PersistProvisionedRuntime(ctx context.Context, instance *instancecontracts.Instance) (bool, error) {
	if a == nil || a.instanceRepo == nil {
		return false, nil
	}
	return a.instanceRepo.PersistProvisionedRuntime(ctx, instance)
}

func (a *practiceInstanceRepositoryAdapter) FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error {
	if a == nil || a.awdRepo == nil {
		return nil
	}
	return a.awdRepo.FinishActiveAWDServiceOperationForInstance(ctx, instanceID, status, errorMessage, finishedAt)
}

func (a *practiceInstanceRepositoryAdapter) RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error {
	if a == nil || a.instanceRepo == nil {
		return nil
	}
	return a.instanceRepo.RefreshInstanceExpiry(ctx, instanceID, expiresAt)
}

func (a *practiceInstanceRepositoryAdapter) UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error {
	if a == nil || a.db == nil || a.instanceRepo == nil || a.allocationRepo == nil {
		return nil
	}
	return withInstanceRuntimeLifecycleTx(ctx, a.db, a.instanceRepo, a.allocationRepo, func(instanceTx *instanceinfra.Repository, allocationTx *containerruntimeinfra.AllocationRepository) error {
		release, err := instanceTx.UpdateStatus(ctx, id, status)
		if err != nil || release == nil {
			return err
		}
		return allocationTx.ReleaseRuntimeAllocationsForInstance(ctx, release.InstanceID, release.HostPort)
	})
}

func (a *practiceInstanceRepositoryAdapter) FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*instancecontracts.Instance, error) {
	if a == nil || a.instanceRepo == nil {
		return nil, nil
	}
	return a.instanceRepo.FindByUserAndChallenge(ctx, userID, challengeID)
}

func (a *practiceInstanceRepositoryAdapter) ListPendingInstances(ctx context.Context, limit int) ([]*instancecontracts.Instance, error) {
	if a == nil || a.instanceRepo == nil {
		return nil, nil
	}
	return a.instanceRepo.ListPendingInstances(ctx, limit)
}

func (a *practiceInstanceRepositoryAdapter) TryTransitionStatus(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error) {
	if a == nil || a.instanceRepo == nil {
		return false, nil
	}
	return a.instanceRepo.TryTransitionStatus(ctx, id, fromStatus, toStatus)
}

func (a *practiceInstanceRepositoryAdapter) CountInstancesByStatus(ctx context.Context, statuses []string) (int64, error) {
	if a == nil || a.instanceRepo == nil {
		return 0, nil
	}
	return a.instanceRepo.CountInstancesByStatus(ctx, statuses)
}
