package composition

import (
	"context"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/model"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	instancecmd "ctf-platform/internal/module/instance/application/commands"
	instanceqry "ctf-platform/internal/module/instance/application/queries"
	instanceports "ctf-platform/internal/module/instance/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type InstanceModule struct {
	Handler *runtimehttp.Handler

	PracticeInstanceRepository interface {
		FindByID(ctx context.Context, id int64) (*model.Instance, error)
		UpdateRuntime(ctx context.Context, instance *model.Instance) error
		FinishActiveAWDServiceOperationForInstance(ctx context.Context, instanceID int64, status, errorMessage string, finishedAt time.Time) error
		RefreshInstanceExpiry(ctx context.Context, instanceID int64, expiresAt time.Time) error
		UpdateStatusAndReleasePort(ctx context.Context, id int64, status string) error
		FindByUserAndChallenge(ctx context.Context, userID, challengeID int64) (*model.Instance, error)
		ListPendingInstances(ctx context.Context, limit int) ([]*model.Instance, error)
		TryTransitionStatus(ctx context.Context, id int64, fromStatus, toStatus string) (bool, error)
		CountInstancesByStatus(ctx context.Context, statuses []string) (int64, error)
	}
	PracticeRuntimeService practiceports.RuntimeInstanceService

	service              *runtimeHTTPServiceAdapter
	proxyTrafficRecorder runtimeProxyTrafficRecorder
}

type runtimeProxyTrafficRecorder interface {
	RecordRuntimeProxyTrafficEvent(ctx context.Context, instanceID, userID int64, method, requestPath string, statusCode int) error
	RecordAWDProxyTrafficEvent(ctx context.Context, event model.AWDProxyTrafficEventInput) error
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
	cleanupService := runtimecmd.NewRuntimeCleanupService(module.CleanupRuntime, repo, log.Named("runtime_cleanup_service"))
	provisioningService := runtimecmd.NewProvisioningService(repo, module.ProvisioningRuntime, &cfg.Container, log.Named("runtime_provisioning_service"))
	commandService := instancecmd.NewInstanceService(repo, cleanupService, &cfg.Container, log.Named("instance_service"))
	queryService := instanceqry.NewInstanceService(repo, &cfg.Container)
	proxyTicketService := instanceqry.NewProxyTicketService(
		runtimeinfra.NewProxyTicketStore(root.Cache()),
		repo,
		cfg.Container.ProxyTicketTTL,
	)
	maintenanceService := instancecmd.NewInstanceMaintenanceService(
		repo,
		newInstanceMaintenanceRuntime(module.ManagedContainerInventory, module.ProvisioningRuntime),
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

	if cfg.Container.DefenseSSHEnabled && module.InteractiveExecutor != nil {
		gateway := NewAWDDefenseSSHGateway(
			proxyTicketService,
			repo,
			module.InteractiveExecutor,
			cfg.Container.DefenseSSHHostKeyPath,
			cfg.Container.DefenseSSHPort,
			log.Named("awd_defense_ssh_gateway"),
		)
		root.RegisterBackgroundJob(NewBackgroundJob(
			"awd_defense_ssh_gateway",
			gateway.Start,
			gateway.Stop,
		))
	}

	return &InstanceModule{
		PracticeInstanceRepository: repo,
		PracticeRuntimeService:     newPracticeRuntimeServiceAdapter(cleanupService, provisioningService, module.ManagedContainerInventory),
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
		proxyTrafficRecorder: runtimeinfra.NewProxyTrafficEventRecorder(root.DB()),
	}
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
