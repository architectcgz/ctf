package runtime

import (
	"context"
	"io"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestports "ctf-platform/internal/module/contest/ports"
	opsports "ctf-platform/internal/module/ops/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
	runtimehttp "ctf-platform/internal/module/runtime/api/http"
	runtimeapp "ctf-platform/internal/module/runtime/application"
	runtimecmd "ctf-platform/internal/module/runtime/application/commands"
	runtimeqry "ctf-platform/internal/module/runtime/application/queries"
	runtimeinfra "ctf-platform/internal/module/runtime/infrastructure"
	runtimeports "ctf-platform/internal/module/runtime/ports"
)

type Engine interface {
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

type ContainerInteractiveExecutor interface {
	ExecContainerInteractive(ctx context.Context, containerID string, command []string, stdin io.Reader, stdout io.Writer) error
}

type BackgroundJob struct {
	Name  string
	Start func(context.Context) error
	Stop  func(context.Context) error
}

type Module struct {
	BackgroundJobs []BackgroundJob
	Handler        *runtimehttp.Handler

	PracticeInstanceRepository practiceInstanceRepository
	PracticeRuntimeService     practiceports.RuntimeInstanceService
	ChallengeImageRuntime      challengeports.ImageRuntime
	ChallengeRuntimeProbe      challengeports.ChallengeRuntimeProbe
	OpsRuntimeQuery            opsports.RuntimeQuery
	OpsRuntimeStatsProvider    opsports.RuntimeStatsProvider
	ContestContainerFiles      contestports.AWDContainerFileWriter

	ProxyTicketReader  runtimeports.ProxyTicketInstanceReader
	ProxyTicketService runtimeHTTPProxyTicketService
	SSHExecutor        ContainerInteractiveExecutor

	http runtimeHTTPDeps
}

type Deps struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *gorm.DB
	Cache  *redislib.Client
	Engine Engine
}

type runtimeHTTPDeps struct {
	service              runtimeHTTPService
	proxyTrafficRecorder runtimeports.ProxyTrafficEventRecorder
}

type runtimeInstanceRepository interface {
	runtimeports.InstanceLookupRepository
	runtimeports.InstanceUserLookupRepository
	runtimeports.InstanceAccessRepository
	runtimeports.UserVisibleInstanceRepository
	runtimeports.TeacherInstanceQueryRepository
	runtimeports.InstanceExtendRepository
	runtimeports.InstanceStatusRepository
	runtimeports.ProxyTicketInstanceReader
	runtimeports.CountRunningRepository
}

type practiceInstanceRepository interface {
	practiceports.PracticeInstanceLookupRepository
	practiceports.PracticeInstanceRuntimeWriteRepository
	practiceports.PracticeInstanceAWDOperationRepository
	practiceports.PracticeInstanceStatusRepository
	practiceports.PracticePendingInstanceRepository
	practiceports.PracticeInstanceStatsRepository
}

type runtimeModuleDeps struct {
	input                           Deps
	repo                            runtimeInstanceRepository
	proxyTicketReader               runtimeports.ProxyTicketInstanceReader
	practiceInstanceRepo            practiceInstanceRepository
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
	sshExecutor                     ContainerInteractiveExecutor
	defenseWorkbench                runtimeDefenseWorkbenchRuntime
	defenseWorkbenchReadOnlyEnabled bool
	defenseWorkbenchRoot            string
	defenseSSHEnabled               bool
	defenseSSHHost                  string
	defenseSSHPort                  int
}

type runtimeDefenseWorkbenchRuntime interface {
	ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error)
	ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]runtimeports.ContainerDirectoryEntry, error)
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
	ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error)
}

func Build(deps Deps) *Module {
	internalDeps := buildRuntimeModuleDeps(deps)
	httpDeps := buildRuntimeHTTPDeps(internalDeps)
	practiceDeps := buildRuntimePracticeDeps(internalDeps)
	challengeDeps := buildRuntimeChallengeDeps(internalDeps)
	opsDeps := buildRuntimeOpsDeps(internalDeps)
	contestDeps := buildRuntimeContestDeps(internalDeps)

	return &Module{
		BackgroundJobs:             buildBackgroundJobs(internalDeps),
		PracticeInstanceRepository: practiceDeps.instanceRepository,
		PracticeRuntimeService:     practiceDeps.runtimeService,
		ChallengeImageRuntime:      challengeDeps.imageRuntime,
		ChallengeRuntimeProbe:      challengeDeps.runtimeProbe,
		OpsRuntimeQuery:            opsDeps.query,
		OpsRuntimeStatsProvider:    opsDeps.statsProvider,
		ContestContainerFiles:      contestDeps.containerFiles,
		ProxyTicketReader:          internalDeps.proxyTicketReader,
		ProxyTicketService:         internalDeps.proxyTicketService,
		SSHExecutor:                internalDeps.sshExecutor,
		http:                       httpDeps,
	}
}

func (m *Module) BuildHandler(auditRecorder auditlog.Recorder, cookieConfig runtimehttp.CookieConfig) {
	if m == nil {
		return
	}
	m.Handler = runtimehttp.NewHandler(
		m.http.service,
		auditRecorder,
		cookieConfig,
		m.http.proxyTrafficRecorder,
	)
}

func buildRuntimeModuleDeps(deps Deps) runtimeModuleDeps {
	cfg := deps.Config
	log := deps.Logger
	repo := runtimeinfra.NewRepository(deps.DB)
	cleanupService := runtimecmd.NewRuntimeCleanupService(deps.Engine, repo, log.Named("runtime_cleanup_service"))
	maintenanceService := runtimecmd.NewRuntimeMaintenanceService(repo, deps.Engine, cleanupService, &cfg.Container, log.Named("runtime_maintenance_service"))
	provisioningService := runtimecmd.NewProvisioningService(repo, deps.Engine, &cfg.Container, log.Named("runtime_provisioning_service"))
	var containerStatsService *runtimeapp.ContainerStatsService
	if deps.Engine != nil {
		containerStatsService = runtimeapp.NewContainerStatsService(deps.Engine)
	}
	proxyTicketStore := runtimeinfra.NewProxyTicketStore(deps.Cache)
	proxyTicketService := runtimeqry.NewProxyTicketService(proxyTicketStore, repo, cfg.Container.ProxyTicketTTL)

	return runtimeModuleDeps{
		input:                           deps,
		repo:                            repo,
		proxyTicketReader:               repo,
		practiceInstanceRepo:            repo,
		instanceCommands:                runtimecmd.NewInstanceService(repo, cleanupService, &cfg.Container, log.Named("runtime_instance_service")),
		instanceQueries:                 runtimeqry.NewInstanceService(repo),
		countRunningQuery:               runtimeqry.NewCountRunningService(repo),
		proxyTicketService:              proxyTicketService,
		cleanupService:                  cleanupService,
		maintenanceService:              maintenanceService,
		provisioningService:             provisioningService,
		containerStatsService:           containerStatsService,
		proxyTrafficRecorder:            runtimeinfra.NewProxyTrafficEventRecorder(deps.DB),
		imageRuntime:                    runtimeapp.NewImageRuntimeService(deps.Engine),
		containerFiles:                  runtimeapp.NewContainerFileService(deps.Engine, log.Named("runtime_container_file_service")),
		containerPublicHost:             cfg.Container.PublicHost,
		sshExecutor:                     deps.Engine,
		defenseWorkbench:                deps.Engine,
		defenseWorkbenchReadOnlyEnabled: cfg.Container.DefenseWorkbenchReadOnlyEnabled && deps.Engine != nil,
		defenseWorkbenchRoot:            cfg.Container.DefenseWorkbenchRoot,
		defenseSSHEnabled:               cfg.Container.DefenseSSHEnabled && deps.Engine != nil,
		defenseSSHHost:                  cfg.Container.DefenseSSHHost,
		defenseSSHPort:                  cfg.Container.DefenseSSHPort,
	}
}

func buildBackgroundJobs(deps runtimeModuleDeps) []BackgroundJob {
	cfg := deps.input.Config
	log := deps.input.Logger
	cleaner := runtimeinfra.NewCleaner(deps.maintenanceService, deps.input.Cache, cfg.Container.CleanupLockTTL, log.Named("runtime_cleaner"))
	return []BackgroundJob{
		{
			Name: "runtime_cleaner",
			Start: func(ctx context.Context) error {
				return cleaner.Start(ctx, cfg.Container.CleanupInterval)
			},
			Stop: cleaner.Stop,
		},
	}
}

func buildRuntimeHTTPDeps(deps runtimeModuleDeps) runtimeHTTPDeps {
	return runtimeHTTPDeps{
		service: newRuntimeHTTPServiceAdapter(
			deps.instanceCommands,
			deps.instanceQueries,
			deps.proxyTicketService,
			deps.proxyTicketReader,
			deps.defenseWorkbench,
			deps.input.Config.Container.ProxyBodyPreviewSize,
			deps.defenseSSHEnabled,
			deps.defenseSSHHost,
			deps.defenseSSHPort,
			deps.defenseWorkbenchReadOnlyEnabled,
			deps.defenseWorkbenchRoot,
		),
		proxyTrafficRecorder: deps.proxyTrafficRecorder,
	}
}

type runtimePracticeDeps struct {
	instanceRepository practiceInstanceRepository
	runtimeService     practiceports.RuntimeInstanceService
}

func buildRuntimePracticeDeps(deps runtimeModuleDeps) runtimePracticeDeps {
	return runtimePracticeDeps{
		instanceRepository: deps.practiceInstanceRepo,
		runtimeService:     newRuntimePracticeServiceAdapter(deps.cleanupService, deps.provisioningService),
	}
}

type runtimeChallengeDeps struct {
	imageRuntime challengeports.ImageRuntime
	runtimeProbe challengeports.ChallengeRuntimeProbe
}

func buildRuntimeChallengeDeps(deps runtimeModuleDeps) runtimeChallengeDeps {
	return runtimeChallengeDeps{
		imageRuntime: deps.imageRuntime,
		runtimeProbe: newRuntimeChallengeServiceAdapter(deps.cleanupService, deps.provisioningService, deps.containerPublicHost),
	}
}

type runtimeOpsDeps struct {
	query         opsports.RuntimeQuery
	statsProvider opsports.RuntimeStatsProvider
}

func buildRuntimeOpsDeps(deps runtimeModuleDeps) runtimeOpsDeps {
	return runtimeOpsDeps{
		query:         deps.countRunningQuery,
		statsProvider: newRuntimeOpsStatsProvider(deps.containerStatsService),
	}
}

type runtimeContestDeps struct {
	containerFiles contestports.AWDContainerFileWriter
}

func buildRuntimeContestDeps(deps runtimeModuleDeps) runtimeContestDeps {
	return runtimeContestDeps{
		containerFiles: deps.containerFiles,
	}
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
