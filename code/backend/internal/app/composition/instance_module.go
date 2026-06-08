package composition

import (
	"context"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	contestcontracts "ctf-platform/internal/module/contest/contracts"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	instancecmd "ctf-platform/internal/module/instance/application/commands"
	instanceqry "ctf-platform/internal/module/instance/application/queries"
	instancecontracts "ctf-platform/internal/module/instance/contracts"
	instanceports "ctf-platform/internal/module/instance/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type InstanceModule struct {
	Handler *runtimehttp.Handler

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
	RecordAWDProxyTrafficEvent(ctx context.Context, event contestcontracts.AWDProxyTrafficEventInput) error
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

	repo := runtimeinfra.NewRepository(root.DB())
	defaultCleanupService := module.CleanupService
	var cleanupService interface {
		instanceports.RuntimeCleaner
		RemoveContainer(ctx context.Context, containerID string) error
	} = defaultCleanupService
	provisioningService := module.ProvisioningService
	var maintenanceRuntime interface {
		ListManagedContainers(ctx context.Context) ([]instanceports.ManagedContainer, error)
		InspectManagedContainer(ctx context.Context, containerID string) (*instanceports.ManagedContainerState, error)
		StartContainer(ctx context.Context, containerID string) error
	} = newInstanceMaintenanceRuntime(module.ManagedContainerInventory, module.ProvisioningRuntime)
	practiceRuntimeService := newPracticeRuntimeServiceAdapter(defaultCleanupService, provisioningService, module.ManagedContainerInventory)
	if runtime.nodeRouter != nil {
		cleanupService = runtime.nodeRouter
		maintenanceRuntime = runtime.nodeRouter
		practiceRuntimeService = newNodeScopedPracticeRuntimeServiceAdapter(runtime.nodeRouter)
	}
	commandService := instancecmd.NewInstanceService(repo, cleanupService, &cfg.Container, log.Named("instance_service"))
	queryService := instanceqry.NewInstanceService(repo, &cfg.Container, cfg.Pagination)
	proxyTicketService := buildRuntimeProxyTicketService(root, repo)
	maintenanceService := instancecmd.NewInstanceMaintenanceService(
		repo,
		maintenanceRuntime,
		cleanupService,
		&cfg.Container,
		log.Named("instance_maintenance_service"),
	)
	startupRecovery := instancecmd.NewStartupRuntimeRecoveryService(
		maintenanceService,
		contestinfra.NewRepository(root.DB()),
		repo,
		runtimeinfra.NewPlatformRuntimeStateStore(root.Cache()),
		0,
		log.Named("startup_runtime_recovery"),
	)
	root.RegisterBackgroundJob(NewBackgroundJob(
		"startup_runtime_recovery",
		startupRecovery.Start,
		startupRecovery.Stop,
	))
	cleaner := runtimeinfra.NewCleaner(
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
		PracticeInstanceRepository:  repo,
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
		proxyTrafficRecorder: runtimeinfra.NewProxyTrafficEventRecorder(root.DB()),
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
	m.Handler = runtimehttp.NewHandler(m.service, cfg.Container.PublicHost, cfg.Container.AccessHost, auditRecorder, runtimehttp.CookieConfig{
		Secure:   cfg.Auth.SessionCookieSecure,
		SameSite: cfg.Auth.CookieSameSite(),
	}, m.proxyTrafficRecorder)
}

type instanceMaintenanceRuntimeAdapter struct {
	inventory    runtimeports.ManagedContainerInventory
	provisioning runtimeports.ContainerProvisioningRuntime
}

func newInstanceMaintenanceRuntime(inventory runtimeports.ManagedContainerInventory, provisioning runtimeports.ContainerProvisioningRuntime) *instanceMaintenanceRuntimeAdapter {
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
	return a.inventory.ListManagedContainers(ctx)
}

func (a *instanceMaintenanceRuntimeAdapter) InspectManagedContainer(ctx context.Context, containerID string) (*instanceports.ManagedContainerState, error) {
	if a == nil || a.inventory == nil {
		return nil, nil
	}
	return a.inventory.InspectManagedContainer(ctx, containerID)
}

func (a *instanceMaintenanceRuntimeAdapter) StartContainer(ctx context.Context, containerID string) error {
	if a == nil || a.provisioning == nil {
		return nil
	}
	return a.provisioning.StartContainer(ctx, containerID)
}
