package runtime

import (
	"context"
	"io"
	"time"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contestports "ctf-platform/internal/module/contest/ports"
	opsports "ctf-platform/internal/module/ops/ports"
	practiceports "ctf-platform/internal/module/practice/ports"
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

	PracticeInstanceRepository practiceInstanceRepository
	PracticeRuntimeService     practiceports.RuntimeInstanceService
	ChallengeImageRuntime      challengeports.ImageRuntime
	ChallengeRuntimeProbe      challengeports.ChallengeRuntimeProbe
	OpsRuntimeQuery            opsports.RuntimeQuery
	OpsRuntimeStatsProvider    opsports.RuntimeStatsProvider
	ContestContainerFiles      contestports.AWDContainerFileWriter

	Engine Engine
}

type Deps struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *gorm.DB
	Cache  *redislib.Client
	Engine Engine
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
	input                 Deps
	repo                  runtimeInstanceRepository
	practiceInstanceRepo  practiceInstanceRepository
	countRunningQuery     opsports.RuntimeQuery
	cleanupService        *runtimecmd.RuntimeCleanupService
	provisioningService   *runtimecmd.ProvisioningService
	containerStatsService *runtimeapp.ContainerStatsService
	imageRuntime          challengeports.ImageRuntime
	containerFiles        contestports.AWDContainerFileWriter
	containerPublicHost   string
}

type runtimeDefenseWorkbenchRuntime interface {
	ReadFileFromContainer(ctx context.Context, containerID, filePath string, limit int64) ([]byte, error)
	ListDirectoryFromContainer(ctx context.Context, containerID, dirPath string, limit int) ([]runtimeports.ContainerDirectoryEntry, error)
	WriteFileToContainer(ctx context.Context, containerID, filePath string, content []byte) error
	ExecContainerCommand(ctx context.Context, containerID string, command []string, stdin []byte, limit int64) ([]byte, error)
}

func Build(deps Deps) *Module {
	internalDeps := buildRuntimeModuleDeps(deps)
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
		Engine:                     deps.Engine,
	}
}

func buildRuntimeModuleDeps(deps Deps) runtimeModuleDeps {
	cfg := deps.Config
	log := deps.Logger
	if cfg == nil {
		cfg = &config.Config{}
	}
	if log == nil {
		log = zap.NewNop()
	}
	repo := runtimeinfra.NewRepository(deps.DB)
	cleanupService := runtimecmd.NewRuntimeCleanupService(deps.Engine, repo, log.Named("runtime_cleanup_service"))
	provisioningService := runtimecmd.NewProvisioningService(repo, deps.Engine, &cfg.Container, log.Named("runtime_provisioning_service"))
	var containerStatsService *runtimeapp.ContainerStatsService
	if deps.Engine != nil {
		containerStatsService = runtimeapp.NewContainerStatsService(deps.Engine)
	}

	return runtimeModuleDeps{
		input:                 deps,
		repo:                  repo,
		practiceInstanceRepo:  repo,
		countRunningQuery:     runtimeqry.NewCountRunningService(repo),
		cleanupService:        cleanupService,
		provisioningService:   provisioningService,
		containerStatsService: containerStatsService,
		imageRuntime:          runtimeapp.NewImageRuntimeService(deps.Engine),
		containerFiles:        runtimeapp.NewContainerFileService(deps.Engine, log.Named("runtime_container_file_service")),
		containerPublicHost:   cfg.Container.PublicHost,
	}
}

func buildBackgroundJobs(deps runtimeModuleDeps) []BackgroundJob {
	_ = deps
	return nil
}

type runtimePracticeDeps struct {
	instanceRepository practiceInstanceRepository
	runtimeService     practiceports.RuntimeInstanceService
}

func buildRuntimePracticeDeps(deps runtimeModuleDeps) runtimePracticeDeps {
	return runtimePracticeDeps{
		instanceRepository: deps.practiceInstanceRepo,
		runtimeService:     newRuntimePracticeServiceAdapter(deps.cleanupService, deps.provisioningService, deps.input.Engine),
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
