package composition

import (
	"context"
	"fmt"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	challengemodule "ctf-platform/internal/module/challenge"
	contestmodule "ctf-platform/internal/module/contest"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimedomain "ctf-platform/internal/module/runtime/domain"
	runtimeinfrarepo "ctf-platform/internal/module/runtime/infrastructure"
	runtimeinfra "ctf-platform/internal/module/runtimeinfra"
	"ctf-platform/pkg/errcode"
)

type RuntimeModule struct {
	Handler *runtimehttp.Handler

	http      runtimeHTTPDeps
	practice  runtimePracticeDeps
	challenge runtimeChallengeDeps
	system    runtimeSystemDeps
	contest   runtimeContestDeps
}

type runtimeHTTPDeps struct {
	service runtimeHTTPService
}

type runtimePracticeDeps struct {
	instanceRepository practiceRuntimeRepositoryBridge
	runtimeService     practiceRuntimeInstanceService
}

type runtimeChallengeDeps struct {
	imageRuntime challengemodule.ImageRuntime
}

type runtimeSystemDeps struct {
	query         runtimeSystemQuery
	statsProvider runtimeSystemStatsProvider
}

type runtimeContestDeps struct {
	containerFiles contestmodule.AWDContainerFileWriter
}

func BuildRuntimeModule(root *Root, infra *RuntimeInfraModule) *RuntimeModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()
	repo := runtimeinfrarepo.NewRepository(db)
	engineAdapter := newRuntimeEngineAdapter(infra.Engine)
	cleanupService := runtimeapp.NewRuntimeCleanupService(infra.Engine, log.Named("runtime_cleanup_service"))
	maintenanceService := runtimeapp.NewRuntimeMaintenanceService(repo, engineAdapter, cleanupService, &cfg.Container, log.Named("runtime_maintenance_service"))
	instanceService := runtimeapp.NewInstanceService(repo, cleanupService, &cfg.Container, log.Named("runtime_instance_service"))
	proxyTicketService := runtimeapp.NewProxyTicketService(runtimeinfrarepo.NewProxyTicketStore(cache), cfg.Container.ProxyTicketTTL)
	cleaner := runtimeinfra.NewCleaner(maintenanceService, cache, cfg.Container.CleanupLockTTL, log.Named("runtime_cleaner"))
	root.RegisterBackgroundJob(NewBackgroundJob(
		"runtime_cleaner",
		func(context.Context) error {
			return cleaner.Start(cfg.Container.CleanupInterval)
		},
		cleaner.Stop,
	))

	httpService := newRuntimeHTTPServiceAdapter(
		instanceService,
		proxyTicketService,
		cfg.Container.ProxyBodyPreviewSize,
	)

	return &RuntimeModule{
		http: runtimeHTTPDeps{
			service: httpService,
		},
		practice: runtimePracticeDeps{
			instanceRepository: repo,
			runtimeService:     newPracticeRuntimeInstanceServiceAdapter(cleanupService, runtimeapp.NewProvisioningService(repo, infra.Engine, &cfg.Container, log.Named("runtime_provisioning_service"))),
		},
		challenge: runtimeChallengeDeps{
			imageRuntime: runtimeapp.NewImageRuntimeService(infra.Engine),
		},
		system: runtimeSystemDeps{
			query:         runtimeapp.NewQueryService(repo),
			statsProvider: newSystemRuntimeStatsProvider(runtimeapp.NewContainerStatsService(engineAdapter)),
		},
		contest: runtimeContestDeps{
			containerFiles: runtimeapp.NewContainerFileService(infra.Engine, log.Named("runtime_container_file_service")),
		},
	}
}

func (m *RuntimeModule) BuildHandler(root *Root, system *SystemModule) {
	if m == nil {
		return
	}

	cfg := root.Config()
	m.Handler = runtimehttp.NewHandler(
		m.http.service,
		system.AuditService,
		runtimehttp.CookieConfig{
			Secure:   cfg.Auth.RefreshCookieSecure,
			SameSite: cfg.Auth.CookieSameSite(),
		},
	)
}

type runtimeHTTPService interface {
	DestroyInstanceWithContext(ctx context.Context, instanceID, userID int64) error
	ExtendInstanceWithContext(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error)
	GetAccessURLWithContext(ctx context.Context, instanceID, userID int64) (string, error)
	GetUserInstancesWithContext(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error)
	ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error)
	DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error
	IssueProxyTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, error)
	ResolveProxyTicket(ctx context.Context, ticket string) (*runtimeapp.ProxyTicketClaims, error)
	ProxyTicketMaxAge() int
	ProxyBodyPreviewSize() int
}

type runtimeHTTPInstanceService interface {
	DestroyInstanceWithContext(ctx context.Context, instanceID, userID int64) error
	ExtendInstanceWithContext(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error)
	GetAccessURLWithContext(ctx context.Context, instanceID, userID int64) (string, error)
	GetUserInstancesWithContext(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error)
	ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error)
	DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error
}

type runtimeHTTPProxyTicketService interface {
	IssueTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, time.Time, error)
	ResolveTicket(ctx context.Context, ticket string) (*runtimeapp.ProxyTicketClaims, error)
	MaxAge() int
}

type runtimeHTTPServiceAdapter struct {
	instanceService      runtimeHTTPInstanceService
	proxyTickets         runtimeHTTPProxyTicketService
	proxyBodyPreviewSize int
}

func newRuntimeHTTPServiceAdapter(instanceService runtimeHTTPInstanceService, proxyTickets runtimeHTTPProxyTicketService, proxyBodyPreviewSize int) *runtimeHTTPServiceAdapter {
	return &runtimeHTTPServiceAdapter{
		instanceService:      instanceService,
		proxyTickets:         proxyTickets,
		proxyBodyPreviewSize: proxyBodyPreviewSize,
	}
}

func (a *runtimeHTTPServiceAdapter) DestroyInstanceWithContext(ctx context.Context, instanceID, userID int64) error {
	if a == nil || a.instanceService == nil {
		return errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.instanceService.DestroyInstanceWithContext(ctx, instanceID, userID)
}

func (a *runtimeHTTPServiceAdapter) ExtendInstanceWithContext(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error) {
	if a == nil || a.instanceService == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.instanceService.ExtendInstanceWithContext(ctx, instanceID, userID)
}

func (a *runtimeHTTPServiceAdapter) GetAccessURLWithContext(ctx context.Context, instanceID, userID int64) (string, error) {
	if a == nil || a.instanceService == nil {
		return "", errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.instanceService.GetAccessURLWithContext(ctx, instanceID, userID)
}

func (a *runtimeHTTPServiceAdapter) GetUserInstancesWithContext(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error) {
	if a == nil || a.instanceService == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.instanceService.GetUserInstancesWithContext(ctx, userID)
}

func (a *runtimeHTTPServiceAdapter) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error) {
	if a == nil || a.instanceService == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.instanceService.ListTeacherInstances(ctx, requesterID, requesterRole, query)
}

func (a *runtimeHTTPServiceAdapter) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	if a == nil || a.instanceService == nil {
		return errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.instanceService.DestroyTeacherInstance(ctx, instanceID, requesterID, requesterRole)
}

func (a *runtimeHTTPServiceAdapter) IssueProxyTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, error) {
	if a == nil || a.proxyTickets == nil {
		return "", errRuntimeHTTPProxyTicketServiceUnavailable()
	}

	ticket, _, err := a.proxyTickets.IssueTicket(ctx, user, instanceID)
	return ticket, err
}

func (a *runtimeHTTPServiceAdapter) ResolveProxyTicket(ctx context.Context, ticket string) (*runtimeapp.ProxyTicketClaims, error) {
	if a == nil || a.proxyTickets == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	return a.proxyTickets.ResolveTicket(ctx, ticket)
}

func (a *runtimeHTTPServiceAdapter) ProxyTicketMaxAge() int {
	if a == nil || a.proxyTickets == nil {
		return 0
	}
	return a.proxyTickets.MaxAge()
}

func (a *runtimeHTTPServiceAdapter) ProxyBodyPreviewSize() int {
	if a == nil {
		return 0
	}
	return a.proxyBodyPreviewSize
}

func errRuntimeHTTPInstanceServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("instance application service is not configured"))
}

func errRuntimeHTTPProxyTicketServiceUnavailable() error {
	return errcode.ErrInternal.WithCause(fmt.Errorf("proxy ticket service is not configured"))
}

type runtimeEngineAdapter struct {
	engine *runtimeinfra.Engine
}

func newRuntimeEngineAdapter(engine *runtimeinfra.Engine) *runtimeEngineAdapter {
	return &runtimeEngineAdapter{engine: engine}
}

func (r *runtimeEngineAdapter) ListManagedContainers(ctx context.Context) ([]runtimeapp.ManagedContainer, error) {
	if r == nil || r.engine == nil {
		return []runtimeapp.ManagedContainer{}, nil
	}

	containers, err := r.engine.ListManagedContainers(ctx, runtimedomain.ManagedByFilter())
	if err != nil {
		return nil, err
	}

	result := make([]runtimeapp.ManagedContainer, 0, len(containers))
	for _, item := range containers {
		result = append(result, runtimeapp.ManagedContainer{
			ID:        item.ID,
			Name:      item.Name,
			CreatedAt: item.CreatedAt,
		})
	}
	return result, nil
}

func (r *runtimeEngineAdapter) ListManagedContainerStats(ctx context.Context) ([]runtimeapp.ManagedContainerStat, error) {
	if r == nil || r.engine == nil {
		return []runtimeapp.ManagedContainerStat{}, nil
	}

	stats, err := r.engine.ListManagedContainerStats(ctx, runtimedomain.ManagedByFilter())
	if err != nil {
		return nil, err
	}

	result := make([]runtimeapp.ManagedContainerStat, 0, len(stats))
	for _, item := range stats {
		result = append(result, runtimeapp.ManagedContainerStat{
			ContainerID:   item.ContainerID,
			ContainerName: item.ContainerName,
			CPUPercent:    item.CPUPercent,
			MemoryPercent: item.MemoryPercent,
			MemoryUsage:   item.MemoryUsage,
			MemoryLimit:   item.MemoryLimit,
		})
	}
	return result, nil
}
