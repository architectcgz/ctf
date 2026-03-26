package composition

import (
	"context"
	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
	"fmt"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	challengeports "ctf-platform/internal/module/challenge/ports"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	"ctf-platform/pkg/errcode"
	"go.uber.org/zap"
)

type RuntimeModule struct {
	Handler *runtimehttp.Handler

	http      runtimeHTTPDeps
	practice  runtimePracticeDeps
	challenge runtimeChallengeDeps
	ops       runtimeOpsDeps
	contest   runtimeContestDeps
}

type runtimeHTTPDeps struct {
	service runtimeHTTPService
}

type runtimePracticeDeps struct {
	instanceRepository practiceports.InstanceRepository
	runtimeService     practiceports.RuntimeInstanceService
}

type runtimeChallengeDeps struct {
	imageRuntime challengeports.ImageRuntime
}

type runtimeOpsDeps struct {
	query         runtimeOpsQuery
	statsProvider runtimeOpsStatsProvider
}

type runtimeContestDeps struct {
	containerFiles contestports.AWDContainerFileWriter
}

func BuildRuntimeModule(root *Root) *RuntimeModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()
	repo := runtimeinfra.NewRepository(db)
	engine := buildRuntimeEngine(root)
	cleanupService := runtimeapp.NewRuntimeCleanupService(engine, log.Named("runtime_cleanup_service"))
	maintenanceService := runtimeapp.NewRuntimeMaintenanceService(repo, engine, cleanupService, &cfg.Container, log.Named("runtime_maintenance_service"))
	instanceService := runtimeapp.NewInstanceService(repo, cleanupService, &cfg.Container, log.Named("runtime_instance_service"))
	provisioningService := runtimeapp.NewProvisioningService(repo, engine, &cfg.Container, log.Named("runtime_provisioning_service"))
	var containerStatsService *runtimeapp.ContainerStatsService
	if engine != nil {
		containerStatsService = runtimeapp.NewContainerStatsService(engine)
	}
	proxyTicketService := runtimeapp.NewProxyTicketService(runtimeinfra.NewProxyTicketStore(cache), cfg.Container.ProxyTicketTTL)
	containerFileService := runtimeapp.NewContainerFileService(engine, log.Named("runtime_container_file_service"))
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
			runtimeService:     newRuntimePracticeServiceAdapter(cleanupService, provisioningService),
		},
		challenge: runtimeChallengeDeps{
			imageRuntime: runtimeapp.NewImageRuntimeService(engine),
		},
		ops: runtimeOpsDeps{
			query:         runtimeapp.NewQueryService(repo),
			statsProvider: newOpsRuntimeStatsProvider(containerStatsService),
		},
		contest: runtimeContestDeps{
			containerFiles: containerFileService,
		},
	}
}

func buildRuntimeEngine(root *Root) *runtimeinfra.Engine {
	if root == nil {
		return nil
	}

	cfg := root.Config()
	log := root.Logger()
	if cfg == nil {
		return nil
	}
	if cfg.App.Env == "test" {
		if log != nil {
			log.Info("runtime_engine_disabled_in_test_env_for_router")
		}
		return nil
	}

	engine, err := runtimeinfra.NewEngine(&cfg.Container)
	if err != nil {
		if log != nil {
			log.Warn("runtime_engine_init_failed_for_router", zap.Error(err))
		}
		return nil
	}
	return engine
}

func (m *RuntimeModule) BuildHandler(root *Root, ops *OpsModule) {
	if m == nil {
		return
	}

	cfg := root.Config()
	m.Handler = runtimehttp.NewHandler(
		m.http.service,
		ops.AuditService,
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

type runtimePracticeServiceAdapter struct {
	cleaner     *runtimeapp.RuntimeCleanupService
	provisioner *runtimeapp.ProvisioningService
}

func newRuntimePracticeServiceAdapter(cleaner *runtimeapp.RuntimeCleanupService, provisioner *runtimeapp.ProvisioningService) practiceports.RuntimeInstanceService {
	if cleaner == nil && provisioner == nil {
		return nil
	}
	return &runtimePracticeServiceAdapter{
		cleaner:     cleaner,
		provisioner: provisioner,
	}
}

func (a *runtimePracticeServiceAdapter) CleanupRuntime(instance *model.Instance) error {
	if a == nil || a.cleaner == nil {
		return nil
	}
	return a.cleaner.CleanupRuntime(instance)
}

func (a *runtimePracticeServiceAdapter) CreateTopology(ctx context.Context, req *practiceports.TopologyCreateRequest) (*practiceports.TopologyCreateResult, error) {
	if a == nil || a.provisioner == nil || req == nil {
		return nil, nil
	}

	result, err := a.provisioner.CreateTopology(ctx, toRuntimeTopologyCreateRequest(req))
	if err != nil {
		return nil, err
	}
	return fromRuntimeTopologyCreateResult(result), nil
}

func (a *runtimePracticeServiceAdapter) CreateContainer(ctx context.Context, imageName string, env map[string]string, reservedHostPort int) (containerID, networkID string, hostPort, servicePort int, err error) {
	if a == nil || a.provisioner == nil {
		return "", "", 0, 0, nil
	}
	return a.provisioner.CreateContainer(ctx, imageName, env, reservedHostPort)
}

func toRuntimeTopologyCreateRequest(req *practiceports.TopologyCreateRequest) *runtimeapp.TopologyCreateRequest {
	if req == nil {
		return nil
	}

	networks := make([]runtimeapp.TopologyCreateNetwork, 0, len(req.Networks))
	for _, network := range req.Networks {
		networks = append(networks, runtimeapp.TopologyCreateNetwork{
			Key:      network.Key,
			Internal: network.Internal,
		})
	}

	nodes := make([]runtimeapp.TopologyCreateNode, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodes = append(nodes, runtimeapp.TopologyCreateNode{
			Key:          node.Key,
			Image:        node.Image,
			Env:          cloneRuntimeStringMap(node.Env),
			ServicePort:  node.ServicePort,
			IsEntryPoint: node.IsEntryPoint,
			NetworkKeys:  append([]string(nil), node.NetworkKeys...),
			Resources:    cloneRuntimeResourceLimits(node.Resources),
		})
	}

	return &runtimeapp.TopologyCreateRequest{
		Networks:         networks,
		Nodes:            nodes,
		Policies:         append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		ReservedHostPort: req.ReservedHostPort,
	}
}

func fromRuntimeTopologyCreateResult(result *runtimeapp.TopologyCreateResult) *practiceports.TopologyCreateResult {
	if result == nil {
		return nil
	}
	return &practiceports.TopologyCreateResult{
		PrimaryContainerID: result.PrimaryContainerID,
		NetworkID:          result.NetworkID,
		AccessURL:          result.AccessURL,
		RuntimeDetails:     result.RuntimeDetails,
	}
}

func cloneRuntimeStringMap(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	output := make(map[string]string, len(input))
	for key, value := range input {
		output[key] = value
	}
	return output
}

func cloneRuntimeResourceLimits(input *model.ResourceLimits) *model.ResourceLimits {
	if input == nil {
		return nil
	}
	return &model.ResourceLimits{
		CPUQuota:  input.CPUQuota,
		Memory:    input.Memory,
		PidsLimit: input.PidsLimit,
	}
}
