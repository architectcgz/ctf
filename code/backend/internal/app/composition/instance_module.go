package composition

import (
	"context"
	"time"

	"go.uber.org/zap"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/model"
	instanceapp "ctf-platform/internal/module/instance/application"
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
	cleanupService := runtimecmd.NewRuntimeCleanupService(module.Engine, repo, log.Named("runtime_cleanup_service"))
	commandService := instancecmd.NewInstanceService(repo, cleanupService, &cfg.Container, log.Named("instance_service"))
	queryService := instanceqry.NewInstanceService(repo)
	proxyTicketService := instanceqry.NewProxyTicketService(
		runtimeinfra.NewProxyTicketStore(root.Cache()),
		repo,
		cfg.Container.ProxyTicketTTL,
	)
	defenseWorkbenchService := instanceapp.NewAWDDefenseWorkbenchService(
		repo,
		newInstanceAWDDefenseWorkbenchRuntime(module.Engine),
		instanceapp.AWDDefenseWorkbenchConfig{
			ReadOnlyEnabled: cfg.Container.DefenseWorkbenchReadOnlyEnabled,
			Root:            cfg.Container.DefenseWorkbenchRoot,
		},
	)
	maintenanceService := instancecmd.NewInstanceMaintenanceService(
		repo,
		module.Engine,
		cleanupService,
		&cfg.Container,
		log.Named("instance_maintenance_service"),
	)
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

	if cfg.Container.DefenseSSHEnabled && module.Engine != nil {
		gateway := NewAWDDefenseSSHGateway(
			proxyTicketService,
			repo,
			module.Engine,
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
		PracticeInstanceRepository: module.PracticeInstanceRepository,
		PracticeRuntimeService:     module.PracticeRuntimeService,
		service: newRuntimeHTTPServiceAdapter(
			commandService,
			queryService,
			proxyTicketService,
			defenseWorkbenchService,
			cfg.Container.ProxyBodyPreviewSize,
			int(cfg.Container.ProxyTicketTTL.Seconds()),
			cfg.Container.DefenseSSHEnabled && module.Engine != nil,
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
	m.Handler = runtimehttp.NewHandler(m.service, auditRecorder, runtimehttp.CookieConfig{
		Secure:   cfg.Auth.SessionCookieSecure,
		SameSite: cfg.Auth.CookieSameSite(),
	}, m.proxyTrafficRecorder)
}

type instanceAWDDefenseWorkbenchRuntimeAdapter struct {
	runtime runtimeports.ContainerFileRuntime
}

func newInstanceAWDDefenseWorkbenchRuntime(runtime runtimeports.ContainerFileRuntime) *instanceAWDDefenseWorkbenchRuntimeAdapter {
	if runtime == nil {
		return nil
	}
	return &instanceAWDDefenseWorkbenchRuntimeAdapter{runtime: runtime}
}

func (a *instanceAWDDefenseWorkbenchRuntimeAdapter) ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error) {
	if a == nil || a.runtime == nil {
		return nil, nil
	}
	return a.runtime.ReadFileFromContainer(ctx, containerID, filePath, limit)
}

func (a *instanceAWDDefenseWorkbenchRuntimeAdapter) ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]instanceports.ContainerDirectoryEntry, error) {
	if a == nil || a.runtime == nil {
		return nil, nil
	}
	entries, err := a.runtime.ListDirectoryFromContainer(ctx, containerID, dirPath, limit)
	if err != nil {
		return nil, err
	}
	result := make([]instanceports.ContainerDirectoryEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, instanceports.ContainerDirectoryEntry{
			Name: entry.Name,
			Type: entry.Type,
			Size: entry.Size,
		})
	}
	return result, nil
}

func (a *instanceAWDDefenseWorkbenchRuntimeAdapter) WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error {
	if a == nil || a.runtime == nil {
		return nil
	}
	return a.runtime.WriteFileToContainer(ctx, containerID, filePath, content)
}

func (a *instanceAWDDefenseWorkbenchRuntimeAdapter) ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error) {
	if a == nil || a.runtime == nil {
		return nil, nil
	}
	return a.runtime.ExecContainerCommand(ctx, containerID, command, stdin, limit)
}
