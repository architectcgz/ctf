package composition

import (
	"context"
	"ctf-platform/internal/model"
	contestports "ctf-platform/internal/module/contest/ports"
	opsports "ctf-platform/internal/module/ops/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimeports "ctf-platform/internal/module/runtime/ports"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	challengeports "ctf-platform/internal/module/challenge/ports"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeqry "ctf-platform/internal/module/runtime/application/queries"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	"ctf-platform/pkg/errcode"
	"go.uber.org/zap"
)

type runtimeEngine interface {
	CreateNetwork(ctx context.Context, name string, labels map[string]string, internal bool, allowExisting bool) (string, error)
	CreateContainer(ctx context.Context, cfg *model.ContainerConfig) (string, error)
	ResolveServicePort(ctx context.Context, imageRef string, preferredPort int) (int, error)
	ConnectContainerToNetwork(ctx context.Context, containerID, networkName string) error
	InspectContainerNetworkIPs(ctx context.Context, containerID string) (map[string]string, error)
	InspectManagedContainer(ctx context.Context, containerID string) (*runtimeports.ManagedContainerState, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string, timeout time.Duration) error
	RemoveContainer(ctx context.Context, containerID string, force bool) error
	RemoveNetwork(ctx context.Context, networkID string) error
	ApplyACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error
	RemoveACLRules(ctx context.Context, rules []model.InstanceRuntimeACLRule) error
	ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error)
	ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]runtimeports.ContainerDirectoryEntry, error)
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
	ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error)
	InspectImageSize(ctx context.Context, imageRef string) (int64, error)
	RemoveImage(ctx context.Context, imageRef string) error
	ListManagedContainers(ctx context.Context) ([]runtimeports.ManagedContainer, error)
	ListManagedContainerStats(ctx context.Context) ([]runtimeports.ManagedContainerStat, error)
	ExecContainerInteractive(ctx context.Context, containerID string, command []string, stdin io.Reader, stdout io.Writer) error
}

type runtimeContainerInteractiveExecutor interface {
	ExecContainerInteractive(ctx context.Context, containerID string, command []string, stdin io.Reader, stdout io.Writer) error
}

type runtimeDefenseWorkbenchRuntime interface {
	ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error)
	ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]runtimeports.ContainerDirectoryEntry, error)
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
	ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error)
}

type RuntimeModule struct {
	Handler *runtimehttp.Handler

	PracticeInstanceRepository *runtimeinfra.Repository
	PracticeRuntimeService     practiceports.RuntimeInstanceService
	ChallengeImageRuntime      challengeports.ImageRuntime
	ChallengeRuntimeProbe      challengeports.ChallengeRuntimeProbe
	OpsRuntimeQuery            opsports.RuntimeQuery
	OpsRuntimeStatsProvider    opsports.RuntimeStatsProvider
	ContestContainerFiles      contestports.AWDContainerFileWriter

	http      runtimeHTTPDeps
	practice  runtimePracticeDeps
	challenge runtimeChallengeDeps
	ops       runtimeOpsDeps
	contest   runtimeContestDeps
}

type runtimeHTTPDeps struct {
	service              runtimeHTTPService
	proxyTrafficRecorder runtimeports.ProxyTrafficEventRecorder
}

type runtimePracticeDeps struct {
	instanceRepository *runtimeinfra.Repository
	runtimeService     practiceports.RuntimeInstanceService
}

type runtimeChallengeDeps struct {
	imageRuntime challengeports.ImageRuntime
	runtimeProbe challengeports.ChallengeRuntimeProbe
}

type runtimeOpsDeps struct {
	query         opsports.RuntimeQuery
	statsProvider opsports.RuntimeStatsProvider
}

type runtimeContestDeps struct {
	containerFiles contestports.AWDContainerFileWriter
}

type runtimeModuleDeps struct {
	repo                            *runtimeinfra.Repository
	practiceInstanceRepo            *runtimeinfra.Repository
	instanceCommands                runtimeHTTPCommandService
	instanceQueries                 runtimeHTTPQueryService
	countRunningQuery               opsports.RuntimeQuery
	proxyTicketService              runtimeHTTPProxyTicketService
	cleanupService                  *runtimecmd.RuntimeCleanupService
	maintenanceService              *runtimecmd.RuntimeMaintenanceService
	provisioningService             *runtimecmd.ProvisioningService
	containerStatsService           *runtimeapp.ContainerStatsService
	imageRuntime                    challengeports.ImageRuntime
	containerFiles                  contestports.AWDContainerFileWriter
	proxyTrafficRecorder            runtimeports.ProxyTrafficEventRecorder
	containerPublicHost             string
	sshExecutor                     runtimeContainerInteractiveExecutor
	defenseWorkbench                runtimeDefenseWorkbenchRuntime
	defenseSSHEnabled               bool
	defenseSSHHost                  string
	defenseSSHPort                  int
	defenseWorkbenchReadOnlyEnabled bool
	defenseWorkbenchRoot            string
}

func BuildRuntimeModule(root *Root) *RuntimeModule {
	engine := buildRuntimeEngine(root)
	deps := buildRuntimeModuleDeps(root, engine)
	registerRuntimeBackgroundJobs(root, deps)
	httpDeps := buildRuntimeHTTPDeps(root, deps)
	practiceDeps := buildRuntimePracticeDeps(deps)
	challengeDeps := buildRuntimeChallengeDeps(deps)
	opsDeps := buildRuntimeOpsDeps(deps)
	contestDeps := buildRuntimeContestDeps(deps)

	return &RuntimeModule{
		PracticeInstanceRepository: practiceDeps.instanceRepository,
		PracticeRuntimeService:     practiceDeps.runtimeService,
		ChallengeImageRuntime:      challengeDeps.imageRuntime,
		ChallengeRuntimeProbe:      challengeDeps.runtimeProbe,
		OpsRuntimeQuery:            opsDeps.query,
		OpsRuntimeStatsProvider:    opsDeps.statsProvider,
		ContestContainerFiles:      contestDeps.containerFiles,
		http:                       httpDeps,
		practice:                   practiceDeps,
		challenge:                  challengeDeps,
		ops:                        opsDeps,
		contest:                    contestDeps,
	}
}

func buildRuntimeModuleDeps(root *Root, engine runtimeEngine) runtimeModuleDeps {
	cfg := root.Config()
	log := root.Logger()
	repo := runtimeinfra.NewRepository(root.DB())
	cleanupService := runtimecmd.NewRuntimeCleanupService(engine, repo, log.Named("runtime_cleanup_service"))
	maintenanceService := runtimecmd.NewRuntimeMaintenanceService(repo, engine, cleanupService, &cfg.Container, log.Named("runtime_maintenance_service"))
	provisioningService := runtimecmd.NewProvisioningService(repo, engine, &cfg.Container, log.Named("runtime_provisioning_service"))
	var containerStatsService *runtimeapp.ContainerStatsService
	if engine != nil {
		containerStatsService = runtimeapp.NewContainerStatsService(engine)
	}
	proxyTicketStore := runtimeinfra.NewProxyTicketStore(root.Cache())
	proxyTicketService := runtimeqry.NewProxyTicketService(proxyTicketStore, repo, cfg.Container.ProxyTicketTTL)

	return runtimeModuleDeps{
		repo:                            repo,
		practiceInstanceRepo:            repo,
		instanceCommands:                runtimecmd.NewInstanceService(repo, cleanupService, &cfg.Container, log.Named("runtime_instance_service")),
		instanceQueries:                 runtimeqry.NewInstanceService(repo),
		countRunningQuery:               runtimeqry.NewCountRunningService(repo),
		proxyTicketService:              proxyTicketService,
		cleanupService:                  cleanupService,
		maintenanceService:              maintenanceService,
		provisioningService:             provisioningService,
		containerStatsService:           containerStatsService,
		proxyTrafficRecorder:            runtimeinfra.NewProxyTrafficEventRecorder(root.DB()),
		imageRuntime:                    runtimeapp.NewImageRuntimeService(engine),
		containerFiles:                  runtimeapp.NewContainerFileService(engine, log.Named("runtime_container_file_service")),
		containerPublicHost:             cfg.Container.PublicHost,
		sshExecutor:                     engine,
		defenseWorkbench:                engine,
		defenseSSHEnabled:               cfg.Container.DefenseSSHEnabled && engine != nil,
		defenseSSHHost:                  cfg.Container.DefenseSSHHost,
		defenseSSHPort:                  cfg.Container.DefenseSSHPort,
		defenseWorkbenchReadOnlyEnabled: cfg.Container.DefenseWorkbenchReadOnlyEnabled && engine != nil,
		defenseWorkbenchRoot:            cfg.Container.DefenseWorkbenchRoot,
	}
}

func registerRuntimeBackgroundJobs(root *Root, deps runtimeModuleDeps) {
	cfg := root.Config()
	log := root.Logger()
	cleaner := runtimeinfra.NewCleaner(deps.maintenanceService, root.Cache(), cfg.Container.CleanupLockTTL, log.Named("runtime_cleaner"))
	root.RegisterBackgroundJob(NewBackgroundJob(
		"runtime_cleaner",
		func(ctx context.Context) error {
			return cleaner.Start(ctx, cfg.Container.CleanupInterval)
		},
		cleaner.Stop,
	))

	if cfg.Container.DefenseSSHEnabled && deps.proxyTicketService != nil && deps.repo != nil && deps.sshExecutor != nil {
		gateway := NewAWDDefenseSSHGateway(
			deps.proxyTicketService,
			deps.repo,
			deps.sshExecutor,
			cfg.Container.DefenseSSHPort,
			log.Named("awd_defense_ssh_gateway"),
		)
		root.RegisterBackgroundJob(NewBackgroundJob(
			"awd_defense_ssh_gateway",
			gateway.Start,
			gateway.Stop,
		))
	}
}

func buildRuntimeHTTPDeps(root *Root, deps runtimeModuleDeps) runtimeHTTPDeps {
	return runtimeHTTPDeps{
		service: newRuntimeHTTPServiceAdapter(
			deps.instanceCommands,
			deps.instanceQueries,
			deps.proxyTicketService,
			deps.repo,
			deps.defenseWorkbench,
			root.Config().Container.ProxyBodyPreviewSize,
			deps.defenseSSHEnabled,
			deps.defenseSSHHost,
			deps.defenseSSHPort,
			deps.defenseWorkbenchReadOnlyEnabled,
			deps.defenseWorkbenchRoot,
		),
		proxyTrafficRecorder: deps.proxyTrafficRecorder,
	}
}

func buildRuntimePracticeDeps(deps runtimeModuleDeps) runtimePracticeDeps {
	return runtimePracticeDeps{
		instanceRepository: deps.practiceInstanceRepo,
		runtimeService:     newRuntimePracticeServiceAdapter(deps.cleanupService, deps.provisioningService),
	}
}

func buildRuntimeChallengeDeps(deps runtimeModuleDeps) runtimeChallengeDeps {
	return runtimeChallengeDeps{
		imageRuntime: deps.imageRuntime,
		runtimeProbe: newRuntimeChallengeServiceAdapter(deps.cleanupService, deps.provisioningService, deps.containerPublicHost),
	}
}

func buildRuntimeOpsDeps(deps runtimeModuleDeps) runtimeOpsDeps {
	return runtimeOpsDeps{
		query:         deps.countRunningQuery,
		statsProvider: newRuntimeOpsStatsProvider(deps.containerStatsService),
	}
}

func buildRuntimeContestDeps(deps runtimeModuleDeps) runtimeContestDeps {
	return runtimeContestDeps{
		containerFiles: deps.containerFiles,
	}
}

type runtimeOpsStatsProviderAdapter struct {
	service *runtimeapp.ContainerStatsService
}

func newRuntimeOpsStatsProvider(service *runtimeapp.ContainerStatsService) opsports.RuntimeStatsProvider {
	return &runtimeOpsStatsProviderAdapter{service: service}
}

func (p *runtimeOpsStatsProviderAdapter) ListManagedContainerStats(ctx context.Context) ([]opsports.ManagedContainerStat, error) {
	if p == nil || p.service == nil {
		return []opsports.ManagedContainerStat{}, nil
	}

	stats, err := p.service.ListManagedContainerStats(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]opsports.ManagedContainerStat, 0, len(stats))
	for _, item := range stats {
		result = append(result, opsports.ManagedContainerStat{
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

func buildRuntimeEngine(root *Root) runtimeEngine {
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
			log.Info("runtime_engine_enabled_with_test_adapter_for_router")
		}
		return newTestRuntimeEngine(log.Named("runtime_test_engine"))
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
			Secure:   cfg.Auth.SessionCookieSecure,
			SameSite: cfg.Auth.CookieSameSite(),
		},
		m.http.proxyTrafficRecorder,
	)
}

type runtimeHTTPService interface {
	DestroyInstance(ctx context.Context, instanceID, userID int64) error
	ExtendInstance(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error)
	GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error)
	GetUserInstances(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error)
	ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error)
	DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error
	IssueProxyTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, error)
	IssueAWDTargetProxyTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID, victimTeamID int64) (string, error)
	IssueAWDDefenseSSHTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64) (*dto.AWDDefenseSSHAccessResp, error)
	ReadAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, filePath string) (*dto.AWDDefenseFileResp, error)
	ListAWDDefenseDirectory(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, dirPath string) (*dto.AWDDefenseDirectoryResp, error)
	SaveAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseFileSaveReq) (*dto.AWDDefenseFileSaveResp, error)
	RunAWDDefenseCommand(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseCommandReq) (*dto.AWDDefenseCommandResp, error)
	ResolveProxyTicket(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error)
	ResolveAWDTargetAccessURL(ctx context.Context, claims *runtimeports.ProxyTicketClaims, contestID, serviceID, victimTeamID int64) (string, error)
	ProxyTicketMaxAge() int
	ProxyBodyPreviewSize() int
}

type runtimeHTTPCommandService interface {
	DestroyInstance(ctx context.Context, instanceID, userID int64) error
	ExtendInstance(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error)
	DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error
}

type runtimeHTTPQueryService interface {
	GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error)
	GetUserInstances(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error)
	ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error)
}

type runtimeHTTPProxyTicketService interface {
	IssueTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, time.Time, error)
	IssueAWDTargetTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID, victimTeamID int64) (string, time.Time, error)
	IssueAWDDefenseSSHTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64) (string, time.Time, error)
	ResolveTicket(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error)
	ResolveAWDTargetAccessURL(ctx context.Context, claims *runtimeports.ProxyTicketClaims, contestID, serviceID, victimTeamID int64) (string, error)
	MaxAge() int
}

type runtimeHTTPServiceAdapter struct {
	commandService                  runtimeHTTPCommandService
	queryService                    runtimeHTTPQueryService
	proxyTickets                    runtimeHTTPProxyTicketService
	proxyTicketReader               runtimeports.ProxyTicketInstanceReader
	defenseWorkbench                runtimeDefenseWorkbenchRuntime
	proxyBodyPreviewSize            int
	defenseSSHEnabled               bool
	defenseSSHHost                  string
	defenseSSHPort                  int
	defenseWorkbenchReadOnlyEnabled bool
	defenseWorkbenchRoot            string
}

func newRuntimeHTTPServiceAdapter(commandService runtimeHTTPCommandService, queryService runtimeHTTPQueryService, proxyTickets runtimeHTTPProxyTicketService, proxyTicketReader runtimeports.ProxyTicketInstanceReader, defenseWorkbench runtimeDefenseWorkbenchRuntime, proxyBodyPreviewSize int, defenseSSHEnabled bool, defenseSSHHost string, defenseSSHPort int, defenseWorkbenchReadOnlyEnabled bool, defenseWorkbenchRoot string) *runtimeHTTPServiceAdapter {
	return &runtimeHTTPServiceAdapter{
		commandService:                  commandService,
		queryService:                    queryService,
		proxyTickets:                    proxyTickets,
		proxyTicketReader:               proxyTicketReader,
		defenseWorkbench:                defenseWorkbench,
		proxyBodyPreviewSize:            proxyBodyPreviewSize,
		defenseSSHEnabled:               defenseSSHEnabled,
		defenseSSHHost:                  defenseSSHHost,
		defenseSSHPort:                  defenseSSHPort,
		defenseWorkbenchReadOnlyEnabled: defenseWorkbenchReadOnlyEnabled,
		defenseWorkbenchRoot:            defenseWorkbenchRoot,
	}
}

func (a *runtimeHTTPServiceAdapter) DestroyInstance(ctx context.Context, instanceID, userID int64) error {
	if a == nil || a.commandService == nil {
		return errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.commandService.DestroyInstance(ctx, instanceID, userID)
}

func (a *runtimeHTTPServiceAdapter) ExtendInstance(ctx context.Context, instanceID, userID int64) (*dto.InstanceResp, error) {
	if a == nil || a.commandService == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.commandService.ExtendInstance(ctx, instanceID, userID)
}

func (a *runtimeHTTPServiceAdapter) GetAccessURL(ctx context.Context, instanceID, userID int64) (string, error) {
	if a == nil || a.queryService == nil {
		return "", errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.queryService.GetAccessURL(ctx, instanceID, userID)
}

func (a *runtimeHTTPServiceAdapter) GetUserInstances(ctx context.Context, userID int64) ([]*dto.InstanceInfo, error) {
	if a == nil || a.queryService == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.queryService.GetUserInstances(ctx, userID)
}

func (a *runtimeHTTPServiceAdapter) ListTeacherInstances(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherInstanceQuery) ([]dto.TeacherInstanceItem, error) {
	if a == nil || a.queryService == nil {
		return nil, errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.queryService.ListTeacherInstances(ctx, requesterID, requesterRole, query)
}

func (a *runtimeHTTPServiceAdapter) DestroyTeacherInstance(ctx context.Context, instanceID, requesterID int64, requesterRole string) error {
	if a == nil || a.commandService == nil {
		return errRuntimeHTTPInstanceServiceUnavailable()
	}
	return a.commandService.DestroyTeacherInstance(ctx, instanceID, requesterID, requesterRole)
}

func (a *runtimeHTTPServiceAdapter) IssueProxyTicket(ctx context.Context, user authctx.CurrentUser, instanceID int64) (string, error) {
	if a == nil || a.proxyTickets == nil {
		return "", errRuntimeHTTPProxyTicketServiceUnavailable()
	}

	ticket, _, err := a.proxyTickets.IssueTicket(ctx, user, instanceID)
	return ticket, err
}

func (a *runtimeHTTPServiceAdapter) IssueAWDTargetProxyTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID, victimTeamID int64) (string, error) {
	if a == nil || a.proxyTickets == nil {
		return "", errRuntimeHTTPProxyTicketServiceUnavailable()
	}

	ticket, _, err := a.proxyTickets.IssueAWDTargetTicket(ctx, user, contestID, serviceID, victimTeamID)
	return ticket, err
}

func (a *runtimeHTTPServiceAdapter) IssueAWDDefenseSSHTicket(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64) (*dto.AWDDefenseSSHAccessResp, error) {
	if a == nil || a.proxyTickets == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	if !a.defenseSSHEnabled || a.defenseSSHHost == "" || a.defenseSSHPort <= 0 {
		return nil, errcode.ErrAWDDefenseSSHUnavailable.WithCause(fmt.Errorf("awd defense ssh gateway is not enabled"))
	}

	ticket, expiresAt, err := a.proxyTickets.IssueAWDDefenseSSHTicket(ctx, user, contestID, serviceID)
	if err != nil {
		return nil, err
	}
	username := fmt.Sprintf("%s+%d+%d", user.Username, contestID, serviceID)
	return &dto.AWDDefenseSSHAccessResp{
		Host:     a.defenseSSHHost,
		Port:     a.defenseSSHPort,
		Username: username,
		Password: ticket,
		Command:  fmt.Sprintf("ssh %s@%s -p %d", username, a.defenseSSHHost, a.defenseSSHPort),
		SSHProfile: &dto.SSHProfileResp{
			Alias:    buildAWDDefenseSSHProfileAlias(contestID, serviceID),
			HostName: a.defenseSSHHost,
			Port:     a.defenseSSHPort,
			User:     username,
		},
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}, nil
}

func buildAWDDefenseSSHProfileAlias(contestID, serviceID int64) string {
	return fmt.Sprintf("ctf-awd-%d-%d", contestID, serviceID)
}

const (
	awdDefenseMaxFileSize      = 256 * 1024
	awdDefenseMaxDirectoryList = 300
	awdDefenseMaxCommandSize   = 2000
	awdDefenseMaxCommandOutput = 64 * 1024
)

func (a *runtimeHTTPServiceAdapter) ReadAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, filePath string) (*dto.AWDDefenseFileResp, error) {
	if a == nil || a.proxyTicketReader == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	if !a.defenseWorkbenchReadOnlyEnabled {
		return nil, errcode.ErrForbidden
	}
	rootPath, err := normalizeAWDDefenseWorkbenchRoot(a.defenseWorkbenchRoot)
	if err != nil {
		return nil, err
	}
	cleanPath, err := normalizeAWDDefensePath(filePath)
	if err != nil {
		return nil, err
	}
	if isBlockedAWDDefensePath(cleanPath) {
		return nil, errcode.ErrForbidden
	}
	scope, err := a.proxyTicketReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.ContainerID == "" {
		return nil, errcode.ErrForbidden
	}

	containerPath := buildAWDDefenseWorkbenchContainerPath(rootPath, cleanPath)
	content, err := a.defenseWorkbench.ReadFileFromContainer(ctx, scope.ContainerID, containerPath, awdDefenseMaxFileSize)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.AWDDefenseFileResp{
		Path:    cleanPath,
		Content: string(content),
		Size:    len(content),
	}, nil
}

func (a *runtimeHTTPServiceAdapter) ListAWDDefenseDirectory(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, dirPath string) (*dto.AWDDefenseDirectoryResp, error) {
	if a == nil || a.proxyTicketReader == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	if !a.defenseWorkbenchReadOnlyEnabled {
		return nil, errcode.ErrForbidden
	}
	rootPath, err := normalizeAWDDefenseWorkbenchRoot(a.defenseWorkbenchRoot)
	if err != nil {
		return nil, err
	}
	cleanPath, err := normalizeAWDDefenseDirectoryPath(dirPath)
	if err != nil {
		return nil, err
	}
	if cleanPath != "." && isBlockedAWDDefensePath(cleanPath) {
		return nil, errcode.ErrForbidden
	}
	scope, err := a.proxyTicketReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.ContainerID == "" {
		return nil, errcode.ErrForbidden
	}

	containerPath := buildAWDDefenseWorkbenchContainerPath(rootPath, cleanPath)
	entries, err := a.defenseWorkbench.ListDirectoryFromContainer(ctx, scope.ContainerID, containerPath, awdDefenseMaxDirectoryList)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	resp := &dto.AWDDefenseDirectoryResp{
		Path:    cleanPath,
		Entries: make([]dto.AWDDefenseDirectoryEntryResp, 0, len(entries)),
	}
	for _, entry := range entries {
		entryPath := entry.Name
		if cleanPath != "." {
			entryPath = path.Join(cleanPath, entry.Name)
		}
		if isBlockedAWDDefensePath(entryPath) {
			continue
		}
		resp.Entries = append(resp.Entries, dto.AWDDefenseDirectoryEntryResp{
			Name: entry.Name,
			Path: entryPath,
			Type: entry.Type,
			Size: entry.Size,
		})
	}
	return resp, nil
}

func (a *runtimeHTTPServiceAdapter) SaveAWDDefenseFile(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseFileSaveReq) (*dto.AWDDefenseFileSaveResp, error) {
	if a == nil || a.proxyTicketReader == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	cleanPath, err := normalizeAWDDefensePath(req.Path)
	if err != nil {
		return nil, err
	}
	content := []byte(req.Content)
	if len(content) > awdDefenseMaxFileSize {
		return nil, errcode.ErrInvalidParams.WithCause(fmt.Errorf("awd defense file is too large"))
	}
	scope, err := a.proxyTicketReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.ContainerID == "" {
		return nil, errcode.ErrForbidden
	}

	backupPath := ""
	if req.Backup {
		existing, readErr := a.defenseWorkbench.ReadFileFromContainer(ctx, scope.ContainerID, cleanPath, awdDefenseMaxFileSize)
		if readErr == nil {
			backupPath = fmt.Sprintf("%s.bak.%d", cleanPath, time.Now().Unix())
			if err := a.defenseWorkbench.WriteFileToContainer(ctx, scope.ContainerID, backupPath, existing); err != nil {
				return nil, errcode.ErrInternal.WithCause(err)
			}
		}
	}
	if err := a.defenseWorkbench.WriteFileToContainer(ctx, scope.ContainerID, cleanPath, content); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return &dto.AWDDefenseFileSaveResp{
		Path:       cleanPath,
		Size:       len(content),
		BackupPath: backupPath,
	}, nil
}

func (a *runtimeHTTPServiceAdapter) RunAWDDefenseCommand(ctx context.Context, user authctx.CurrentUser, contestID, serviceID int64, req dto.AWDDefenseCommandReq) (*dto.AWDDefenseCommandResp, error) {
	if a == nil || a.proxyTicketReader == nil || a.defenseWorkbench == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	command := strings.TrimSpace(req.Command)
	if command == "" || len(command) > awdDefenseMaxCommandSize {
		return nil, errcode.ErrInvalidParams
	}
	scope, err := a.proxyTicketReader.FindAWDDefenseSSHScope(ctx, user.UserID, contestID, serviceID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if scope == nil || scope.ContainerID == "" {
		return nil, errcode.ErrForbidden
	}

	runCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	output, err := a.defenseWorkbench.ExecContainerCommand(runCtx, scope.ContainerID, []string{"/bin/sh", "-lc", command}, nil, awdDefenseMaxCommandOutput)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return &dto.AWDDefenseCommandResp{
		Command: command,
		Output:  string(output),
	}, nil
}

func normalizeAWDDefenseDirectoryPath(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" || trimmed == "." {
		return ".", nil
	}
	return normalizeAWDDefensePath(trimmed)
}

func normalizeAWDDefensePath(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", errcode.ErrInvalidParams
	}
	if strings.HasPrefix(trimmed, "/") {
		return "", errcode.ErrInvalidParams
	}
	cleaned := path.Clean(trimmed)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, "../") || strings.Contains(cleaned, "/../") {
		return "", errcode.ErrInvalidParams
	}
	return cleaned, nil
}

func normalizeAWDDefenseWorkbenchRoot(input string) (string, error) {
	trimmed := strings.TrimSpace(input)
	if trimmed == "" {
		return "", errcode.ErrForbidden
	}
	cleaned := path.Clean(trimmed)
	if !path.IsAbs(cleaned) || cleaned == "/" || strings.Contains(cleaned, "/../") {
		return "", errcode.ErrForbidden
	}
	return cleaned, nil
}

func buildAWDDefenseWorkbenchContainerPath(rootPath, cleanPath string) string {
	if cleanPath == "." {
		return rootPath
	}
	return path.Join(rootPath, cleanPath)
}

func isBlockedAWDDefensePath(cleanPath string) bool {
	trimmed := strings.TrimSpace(cleanPath)
	if trimmed == "" {
		return true
	}
	parts := strings.Split(path.Clean(trimmed), "/")
	for _, part := range parts {
		name := strings.ToLower(strings.TrimSpace(part))
		if name == "" || name == "." {
			continue
		}
		if name == ".ssh" || name == ".env" || strings.HasPrefix(name, ".env.") || strings.HasSuffix(name, ".env") {
			return true
		}
		if name == "proc" || name == "sys" || name == "dev" || name == "run" {
			return true
		}
		if name == "var" && len(parts) > 1 {
			for i := 0; i < len(parts)-1; i++ {
				if strings.EqualFold(parts[i], "var") && strings.EqualFold(parts[i+1], "run") {
					return true
				}
			}
		}
	}
	return false
}

func (a *runtimeHTTPServiceAdapter) ResolveProxyTicket(ctx context.Context, ticket string) (*runtimeports.ProxyTicketClaims, error) {
	if a == nil || a.proxyTickets == nil {
		return nil, errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	return a.proxyTickets.ResolveTicket(ctx, ticket)
}

func (a *runtimeHTTPServiceAdapter) ResolveAWDTargetAccessURL(ctx context.Context, claims *runtimeports.ProxyTicketClaims, contestID, serviceID, victimTeamID int64) (string, error) {
	if a == nil || a.proxyTickets == nil {
		return "", errRuntimeHTTPProxyTicketServiceUnavailable()
	}
	return a.proxyTickets.ResolveAWDTargetAccessURL(ctx, claims, contestID, serviceID, victimTeamID)
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
	cleaner     *runtimecmd.RuntimeCleanupService
	provisioner *runtimecmd.ProvisioningService
}

func newRuntimePracticeServiceAdapter(cleaner *runtimecmd.RuntimeCleanupService, provisioner *runtimecmd.ProvisioningService) practiceports.RuntimeInstanceService {
	if cleaner == nil && provisioner == nil {
		return nil
	}
	return &runtimePracticeServiceAdapter{
		cleaner:     cleaner,
		provisioner: provisioner,
	}
}

func (a *runtimePracticeServiceAdapter) CleanupRuntime(ctx context.Context, instance *model.Instance) error {
	if a == nil || a.cleaner == nil {
		return nil
	}
	return a.cleaner.CleanupRuntime(ctx, instance)
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

func toRuntimeTopologyCreateRequest(req *practiceports.TopologyCreateRequest) *runtimeports.TopologyCreateRequest {
	if req == nil {
		return nil
	}

	networks := make([]runtimeports.TopologyCreateNetwork, 0, len(req.Networks))
	for _, network := range req.Networks {
		networks = append(networks, runtimeports.TopologyCreateNetwork{
			Key:      network.Key,
			Name:     network.Name,
			Internal: network.Internal,
			Shared:   network.Shared,
		})
	}

	nodes := make([]runtimeports.TopologyCreateNode, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodes = append(nodes, runtimeports.TopologyCreateNode{
			Key:             node.Key,
			Image:           node.Image,
			Env:             cloneRuntimeStringMap(node.Env),
			ServicePort:     node.ServicePort,
			ServiceProtocol: node.ServiceProtocol,
			IsEntryPoint:    node.IsEntryPoint,
			NetworkKeys:     append([]string(nil), node.NetworkKeys...),
			NetworkAliases:  append([]string(nil), node.NetworkAliases...),
			Resources:       cloneRuntimeResourceLimits(node.Resources),
		})
	}

	return &runtimeports.TopologyCreateRequest{
		Networks:                   networks,
		Nodes:                      nodes,
		Policies:                   append([]model.TopologyTrafficPolicy(nil), req.Policies...),
		ReservedHostPort:           req.ReservedHostPort,
		DisableEntryPortPublishing: req.DisableEntryPortPublishing,
	}
}

func fromRuntimeTopologyCreateResult(result *runtimeports.TopologyCreateResult) *practiceports.TopologyCreateResult {
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

type runtimeChallengeServiceAdapter struct {
	cleaner     *runtimecmd.RuntimeCleanupService
	provisioner *runtimecmd.ProvisioningService
	publicHost  string
}

func newRuntimeChallengeServiceAdapter(cleaner *runtimecmd.RuntimeCleanupService, provisioner *runtimecmd.ProvisioningService, publicHost string) challengeports.ChallengeRuntimeProbe {
	if cleaner == nil && provisioner == nil {
		return nil
	}
	return &runtimeChallengeServiceAdapter{
		cleaner:     cleaner,
		provisioner: provisioner,
		publicHost:  publicHost,
	}
}

func (a *runtimeChallengeServiceAdapter) CreateTopology(ctx context.Context, req *challengeports.RuntimeTopologyCreateRequest) (*challengeports.RuntimeTopologyCreateResult, error) {
	if a == nil || a.provisioner == nil {
		return nil, fmt.Errorf("runtime provisioning service is not configured")
	}
	if req == nil {
		return nil, fmt.Errorf("runtime topology create request is nil")
	}
	result, err := a.provisioner.CreateTopology(ctx, toRuntimeChallengeTopologyCreateRequest(req))
	if err != nil {
		return nil, err
	}
	return &challengeports.RuntimeTopologyCreateResult{
		AccessURL:      result.AccessURL,
		RuntimeDetails: result.RuntimeDetails,
	}, nil
}

func (a *runtimeChallengeServiceAdapter) CreateContainer(ctx context.Context, imageName string, env map[string]string) (string, model.InstanceRuntimeDetails, error) {
	if a == nil || a.provisioner == nil {
		return "", model.InstanceRuntimeDetails{}, fmt.Errorf("runtime provisioning service is not configured")
	}

	containerID, networkID, hostPort, servicePort, err := a.provisioner.CreateContainer(ctx, imageName, env, 0)
	if err != nil {
		return "", model.InstanceRuntimeDetails{}, err
	}

	accessURL := fmt.Sprintf("http://%s:%d", a.publicHost, hostPort)
	return accessURL, model.InstanceRuntimeDetails{
		Networks: []model.InstanceRuntimeNetwork{
			{
				Key:       model.TopologyDefaultNetworkKey,
				Name:      model.TopologyDefaultNetworkKey,
				NetworkID: networkID,
			},
		},
		Containers: []model.InstanceRuntimeContainer{
			{
				NodeKey:         "default",
				ContainerID:     containerID,
				ServicePort:     servicePort,
				ServiceProtocol: model.ChallengeTargetProtocolHTTP,
				HostPort:        hostPort,
				IsEntryPoint:    true,
				NetworkKeys:     []string{model.TopologyDefaultNetworkKey},
			},
		},
	}, nil
}

func (a *runtimeChallengeServiceAdapter) CleanupRuntimeDetails(ctx context.Context, details model.InstanceRuntimeDetails) error {
	if a == nil || a.cleaner == nil {
		return nil
	}

	rawDetails, err := model.EncodeInstanceRuntimeDetails(details)
	if err != nil {
		return err
	}
	instance := &model.Instance{
		RuntimeDetails: rawDetails,
	}
	return a.cleaner.CleanupRuntime(ctx, instance)
}

func toRuntimeChallengeTopologyCreateRequest(req *challengeports.RuntimeTopologyCreateRequest) *runtimeports.TopologyCreateRequest {
	if req == nil {
		return nil
	}
	networks := make([]runtimeports.TopologyCreateNetwork, 0, len(req.Networks))
	for _, network := range req.Networks {
		networks = append(networks, runtimeports.TopologyCreateNetwork{
			Key:      network.Key,
			Internal: network.Internal,
		})
	}

	nodes := make([]runtimeports.TopologyCreateNode, 0, len(req.Nodes))
	for _, node := range req.Nodes {
		nodes = append(nodes, runtimeports.TopologyCreateNode{
			Key:             node.Key,
			Image:           node.Image,
			Env:             cloneRuntimeStringMap(node.Env),
			ServicePort:     node.ServicePort,
			ServiceProtocol: node.ServiceProtocol,
			IsEntryPoint:    node.IsEntryPoint,
			NetworkKeys:     append([]string(nil), node.NetworkKeys...),
			Resources:       cloneRuntimeResourceLimits(node.Resources),
		})
	}

	return &runtimeports.TopologyCreateRequest{
		Networks: networks,
		Nodes:    nodes,
		Policies: append([]model.TopologyTrafficPolicy(nil), req.Policies...),
	}
}
