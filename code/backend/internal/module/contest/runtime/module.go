package runtime

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	challengeports "ctf-platform/internal/module/challenge/ports"
	contesthttp "ctf-platform/internal/module/contest/api/http"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestjobs "ctf-platform/internal/module/contest/application/jobs"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contestports "ctf-platform/internal/module/contest/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type BackgroundJob struct {
	Name string
	Run  func(context.Context)
}

type Module struct {
	AWDHandler           *contesthttp.AWDHandler
	ChallengeHandler     *contesthttp.ChallengeHandler
	Handler              *contesthttp.Handler
	ParticipationHandler *contesthttp.ParticipationHandler
	SubmissionHandler    *contesthttp.SubmissionHandler
	TeamHandler          *contesthttp.TeamHandler

	BackgroundJobs []BackgroundJob
	bindRealtime   func(contestports.RealtimeBroadcaster)
}

type Deps struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *gorm.DB
	Cache  *redislib.Client
	Events platformevents.Bus

	ChallengeCatalog      challengecontracts.ContestChallengeContract
	AWDChallengeQueryRepo challengeports.AWDChallengeQueryRepository
	ImageRepo             challengecontracts.ImageStore
	FlagValidator         challengecontracts.FlagValidator
	ContainerFiles        contestports.AWDContainerFileWriter
	RuntimeProbe          challengeports.ChallengeRuntimeProbe
}

type moduleDeps struct {
	input                 Deps
	contestCommands       *contestinfra.Repository
	contestLookup         contestports.ContestLookupRepository
	contestList           contestports.ContestListRepository
	contestScoreboard     contestports.ContestScoreboardRepository
	contestAdmin          contestports.ContestScoreboardAdminRepository
	contestStatus         contestports.ContestStatusRepository
	awdRepo               *contestinfra.AWDRepository
	challengeRepo         *contestinfra.ChallengeRepository
	teamRepo              *contestinfra.TeamRepository
	teamFinder            contestports.ContestTeamFinder
	participationRepo     *contestinfra.ParticipationRepository
	submissionRepo        *contestinfra.SubmissionRepository
	challengeCatalog      challengecontracts.ContestChallengeContract
	awdChallengeQueryRepo challengeports.AWDChallengeQueryRepository
	imageRepo             challengecontracts.ImageStore
	flagValidator         challengecontracts.FlagValidator
	containerFiles        contestports.AWDContainerFileWriter
	runtimeProbe          challengeports.ChallengeRuntimeProbe
}

func Build(deps Deps) *Module {
	internalDeps := newModuleDeps(deps)

	handler, scoreboardCommands, statusUpdater := buildCoreHandler(internalDeps)
	awdHandler, awdUpdater, awdCommands := buildAWDHandler(internalDeps)
	challengeHandler := buildChallengeHandler(internalDeps)
	participationHandler, participationCommands := buildParticipationHandler(internalDeps)
	teamHandler := buildTeamHandler(internalDeps)
	submissionHandler, submissionService := buildSubmissionHandler(internalDeps, scoreboardCommands)

	return &Module{
		AWDHandler:           awdHandler,
		ChallengeHandler:     challengeHandler,
		Handler:              handler,
		ParticipationHandler: participationHandler,
		SubmissionHandler:    submissionHandler,
		TeamHandler:          teamHandler,
		BackgroundJobs: []BackgroundJob{
			{Name: "contest_status_updater", Run: statusUpdater.Start},
			{Name: "awd_round_updater", Run: awdUpdater.Start},
		},
		bindRealtime: func(broadcaster contestports.RealtimeBroadcaster) {
			scoreboardCommands.SetRealtimeBroadcaster(broadcaster)
			participationCommands.SetRealtimeBroadcaster(broadcaster)
			submissionService.SetRealtimeBroadcaster(broadcaster)
			awdCommands.SetRealtimeBroadcaster(broadcaster)
		},
	}
}

func (m *Module) BindRealtimeBroadcaster(broadcaster contestports.RealtimeBroadcaster) {
	if m == nil || m.bindRealtime == nil {
		return
	}
	m.bindRealtime(broadcaster)
}

func newModuleDeps(deps Deps) *moduleDeps {
	contestRepo := contestinfra.NewRepository(deps.DB)
	awdRepo := contestinfra.NewAWDRepository(deps.DB)
	challengeRepo := contestinfra.NewChallengeRepository(deps.DB)
	teamRepo := contestinfra.NewTeamRepository(deps.DB)
	participationRepo := contestinfra.NewParticipationRepository(deps.DB)
	submissionRepo := contestinfra.NewSubmissionRepository(deps.DB)

	return &moduleDeps{
		input:                 deps,
		contestCommands:       contestRepo,
		contestLookup:         contestRepo,
		contestList:           contestRepo,
		contestScoreboard:     contestRepo,
		contestAdmin:          contestRepo,
		contestStatus:         contestRepo,
		awdRepo:               awdRepo,
		challengeRepo:         challengeRepo,
		teamRepo:              teamRepo,
		teamFinder:            teamRepo,
		participationRepo:     participationRepo,
		submissionRepo:        submissionRepo,
		challengeCatalog:      deps.ChallengeCatalog,
		awdChallengeQueryRepo: deps.AWDChallengeQueryRepo,
		imageRepo:             deps.ImageRepo,
		flagValidator:         deps.FlagValidator,
		containerFiles:        deps.ContainerFiles,
		runtimeProbe:          deps.RuntimeProbe,
	}
}

func buildCoreHandler(deps *moduleDeps) (*contesthttp.Handler, *contestcmd.ScoreboardAdminService, *contestjobs.StatusUpdater) {
	cfg := deps.input.Config
	log := deps.input.Logger
	cache := deps.input.Cache

	scoreboardCommands := contestcmd.NewScoreboardAdminService(deps.contestAdmin, cache, &cfg.Contest)
	scoreboardQueries := contestqry.NewScoreboardService(deps.contestScoreboard, cache, &cfg.Contest, log.Named("contest_scoreboard_service"))
	contestCommands := contestcmd.NewContestService(deps.contestCommands, deps.awdRepo, log.Named("contest_service"))
	contestQueries := contestqry.NewContestService(deps.contestList, log.Named("contest_service"))
	readinessQueries := contestqry.NewAWDService(deps.awdRepo, deps.contestLookup)
	statusUpdater := contestjobs.NewStatusUpdater(
		deps.contestStatus,
		cache,
		cfg.Contest.StatusUpdateInterval,
		cfg.Contest.StatusUpdateBatchSize,
		cfg.Contest.StatusUpdateLockTTL,
		log.Named("contest_status_updater"),
		deps.awdRepo,
	)

	return contesthttp.NewHandler(contestCommands, contestQueries, readinessQueries, scoreboardQueries, scoreboardCommands), scoreboardCommands, statusUpdater
}

func buildAWDHandler(deps *moduleDeps) (*contesthttp.AWDHandler, *contestjobs.AWDRoundUpdater, *contestcmd.AWDService) {
	cfg := deps.input.Config
	log := deps.input.Logger
	cache := deps.input.Cache
	db := deps.input.DB
	scoreboardCache := contestinfra.NewScoreboardCache(db, cache)

	awdUpdater := contestjobs.NewAWDRoundUpdater(
		deps.awdRepo,
		cache,
		cfg.Contest.AWD,
		cfg.Container.FlagGlobalSecret,
		contestinfra.NewDockerAWDFlagInjector(db, deps.containerFiles, log.Named("awd_flag_injector")),
		log.Named("awd_round_updater"),
		scoreboardCache,
	)
	if checkerRunner, err := contestinfra.NewDockerCheckerRunner(cfg.Contest.AWD.CheckerSandbox); err == nil {
		awdUpdater.SetCheckerRunner(checkerRunner)
	} else {
		log.Named("awd_round_updater").Warn("checker_sandbox_runner_unavailable", zap.Error(err))
	}
	awdCommands := contestcmd.NewAWDService(
		deps.awdRepo,
		deps.contestLookup,
		cache,
		cfg.Container.FlagGlobalSecret,
		cfg.Contest.AWD,
		log.Named("contest_awd_service"),
		awdUpdater,
		deps.imageRepo,
		deps.awdChallengeQueryRepo,
		deps.runtimeProbe,
		scoreboardCache,
	)
	awdCommands.SetEventBus(deps.input.Events)
	awdQueries := contestqry.NewAWDService(deps.awdRepo, deps.contestLookup)
	awdServiceCommands := contestcmd.NewContestAWDServiceService(
		deps.awdRepo,
		deps.contestLookup,
		deps.challengeRepo,
		deps.challengeCatalog,
		deps.awdChallengeQueryRepo,
		deps.input.Cache,
	)
	awdServiceQueries := contestqry.NewContestAWDServiceQueryService(deps.awdRepo, deps.contestLookup)

	return contesthttp.NewAWDHandler(awdCommands, awdQueries, awdServiceCommands, awdServiceQueries), awdUpdater, awdCommands
}

func buildChallengeHandler(deps *moduleDeps) *contesthttp.ChallengeHandler {
	contestChallengeCommands := contestcmd.NewChallengeService(deps.challengeRepo, deps.challengeCatalog, deps.contestLookup, deps.awdRepo, deps.input.Cache)
	contestChallengeQueries := contestqry.NewChallengeService(deps.challengeRepo, deps.challengeCatalog, deps.contestLookup, deps.awdRepo)
	return contesthttp.NewChallengeHandler(contestChallengeCommands, contestChallengeQueries)
}

func buildParticipationHandler(deps *moduleDeps) (*contesthttp.ParticipationHandler, *contestcmd.ParticipationService) {
	participationCommands := contestcmd.NewParticipationService(deps.contestLookup, deps.participationRepo, deps.teamFinder)
	participationQueries := contestqry.NewParticipationService(deps.contestLookup, deps.participationRepo, deps.teamFinder)
	return contesthttp.NewParticipationHandler(participationCommands, participationQueries), participationCommands
}

func buildTeamHandler(deps *moduleDeps) *contesthttp.TeamHandler {
	teamCommands := contestcmd.NewTeamService(deps.teamRepo, deps.contestLookup)
	teamQueries := contestqry.NewTeamService(deps.teamRepo, deps.contestLookup)
	return contesthttp.NewTeamHandler(teamCommands, teamQueries)
}

func buildSubmissionHandler(deps *moduleDeps, scoreboardCommands *contestcmd.ScoreboardAdminService) (*contesthttp.SubmissionHandler, *contestcmd.SubmissionService) {
	cfg := deps.input.Config

	submissionService := contestcmd.NewSubmissionService(
		deps.contestLookup,
		deps.submissionRepo,
		deps.input.Cache,
		deps.flagValidator,
		deps.teamFinder,
		scoreboardCommands,
		cfg,
	)
	return contesthttp.NewSubmissionHandler(submissionService), submissionService
}
