package composition

import (
	challengecontracts "ctf-platform/internal/module/challenge/contracts"
	contesthttp "ctf-platform/internal/module/contest/api/http"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestjobs "ctf-platform/internal/module/contest/application/jobs"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
	contestports "ctf-platform/internal/module/contest/ports"
)

type ContestModule struct {
	AWDHandler           *contesthttp.AWDHandler
	ChallengeHandler     *contesthttp.ChallengeHandler
	Handler              *contesthttp.Handler
	ParticipationHandler *contesthttp.ParticipationHandler
	SubmissionHandler    *contesthttp.SubmissionHandler
	TeamHandler          *contesthttp.TeamHandler
	bindRealtime         func(contestports.RealtimeBroadcaster)
}

type contestModuleDeps struct {
	root              *Root
	contestCommands   contestports.ContestCommandRepository
	contestLookup     contestports.ContestLookupRepository
	contestList       contestports.ContestListRepository
	contestScoreboard contestports.ContestScoreboardRepository
	contestAdmin      contestports.ContestScoreboardAdminRepository
	contestStatus     contestports.ContestStatusRepository
	awdRepo           contestports.AWDRepository
	challengeRepo     contestports.ContestChallengeRepository
	teamRepo          contestports.ContestTeamRepository
	teamFinder        contestports.ContestTeamFinder
	participationRepo contestports.ContestParticipationRepository
	submissionRepo    contestports.ContestSubmissionRepository
	challengeCatalog  challengecontracts.ContestChallengeContract
	flagValidator     challengecontracts.FlagValidator
	containerFiles    contestports.AWDContainerFileWriter
}

func BuildContestModule(root *Root, challenge *ChallengeModule, runtime *RuntimeModule) *ContestModule {
	deps := newContestModuleDeps(root, challenge, runtime)

	handler, scoreboardCommands, statusUpdater := buildContestCoreHandler(deps)
	awdHandler, awdUpdater := buildContestAWDHandler(deps)
	challengeHandler := buildContestChallengeHandler(deps)
	participationHandler, participationCommands := buildContestParticipationHandler(deps)
	teamHandler := buildContestTeamHandler(deps)
	submissionHandler, submissionService := buildContestSubmissionHandler(deps, scoreboardCommands)

	root.RegisterBackgroundJob(NewLoopBackgroundJob("contest_status_updater", statusUpdater.Start))
	root.RegisterBackgroundJob(NewLoopBackgroundJob("awd_round_updater", awdUpdater.Start))

	return &ContestModule{
		AWDHandler:           awdHandler,
		ChallengeHandler:     challengeHandler,
		Handler:              handler,
		ParticipationHandler: participationHandler,
		SubmissionHandler:    submissionHandler,
		TeamHandler:          teamHandler,
		bindRealtime: func(broadcaster contestports.RealtimeBroadcaster) {
			scoreboardCommands.SetRealtimeBroadcaster(broadcaster)
			participationCommands.SetRealtimeBroadcaster(broadcaster)
			submissionService.SetRealtimeBroadcaster(broadcaster)
		},
	}
}

func (m *ContestModule) BindRealtimeBroadcaster(broadcaster contestports.RealtimeBroadcaster) {
	if m == nil || m.bindRealtime == nil {
		return
	}
	m.bindRealtime(broadcaster)
}

func newContestModuleDeps(root *Root, challenge *ChallengeModule, runtime *RuntimeModule) *contestModuleDeps {
	db := root.DB()

	contestRepo := contestinfra.NewRepository(db)
	awdRepo := contestinfra.NewAWDRepository(db)
	challengeRepo := contestinfra.NewChallengeRepository(db)
	teamRepo := contestinfra.NewTeamRepository(db)
	participationRepo := contestinfra.NewParticipationRepository(db)
	submissionRepo := contestinfra.NewSubmissionRepository(db)

	return &contestModuleDeps{
		root:              root,
		contestCommands:   contestRepo,
		contestLookup:     contestRepo,
		contestList:       contestRepo,
		contestScoreboard: contestRepo,
		contestAdmin:      contestRepo,
		contestStatus:     contestRepo,
		awdRepo:           awdRepo,
		challengeRepo:     challengeRepo,
		teamRepo:          teamRepo,
		teamFinder:        teamRepo,
		participationRepo: participationRepo,
		submissionRepo:    submissionRepo,
		challengeCatalog:  challenge.Catalog,
		flagValidator:     challenge.FlagValidator,
		containerFiles:    runtime.ContestContainerFiles,
	}
}

func buildContestCoreHandler(deps *contestModuleDeps) (*contesthttp.Handler, *contestcmd.ScoreboardAdminService, *contestjobs.StatusUpdater) {
	cfg := deps.root.Config()
	log := deps.root.Logger()
	cache := deps.root.Cache()

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
	)

	return contesthttp.NewHandler(contestCommands, contestQueries, readinessQueries, scoreboardQueries, scoreboardCommands), scoreboardCommands, statusUpdater
}

func buildContestAWDHandler(deps *contestModuleDeps) (*contesthttp.AWDHandler, *contestjobs.AWDRoundUpdater) {
	cfg := deps.root.Config()
	log := deps.root.Logger()
	cache := deps.root.Cache()
	db := deps.root.DB()

	awdUpdater := contestjobs.NewAWDRoundUpdater(
		deps.awdRepo,
		cache,
		cfg.Contest.AWD,
		cfg.Container.FlagGlobalSecret,
		contestinfra.NewDockerAWDFlagInjector(db, deps.containerFiles, log.Named("awd_flag_injector")),
		log.Named("awd_round_updater"),
	)
	awdCommands := contestcmd.NewAWDService(
		deps.awdRepo,
		deps.contestLookup,
		cache,
		cfg.Container.FlagGlobalSecret,
		cfg.Contest.AWD,
		log.Named("contest_awd_service"),
		awdUpdater,
	)
	awdQueries := contestqry.NewAWDService(deps.awdRepo, deps.contestLookup)

	return contesthttp.NewAWDHandler(awdCommands, awdQueries), awdUpdater
}

func buildContestChallengeHandler(deps *contestModuleDeps) *contesthttp.ChallengeHandler {
	contestChallengeCommands := contestcmd.NewChallengeService(deps.challengeRepo, deps.challengeCatalog, deps.contestLookup, deps.root.Cache())
	contestChallengeQueries := contestqry.NewChallengeService(deps.challengeRepo, deps.challengeCatalog, deps.contestLookup)
	return contesthttp.NewChallengeHandler(contestChallengeCommands, contestChallengeQueries)
}

func buildContestParticipationHandler(deps *contestModuleDeps) (*contesthttp.ParticipationHandler, *contestcmd.ParticipationService) {
	participationCommands := contestcmd.NewParticipationService(deps.contestLookup, deps.participationRepo, deps.teamFinder)
	participationQueries := contestqry.NewParticipationService(deps.contestLookup, deps.participationRepo, deps.teamFinder)
	return contesthttp.NewParticipationHandler(participationCommands, participationQueries), participationCommands
}

func buildContestTeamHandler(deps *contestModuleDeps) *contesthttp.TeamHandler {
	teamCommands := contestcmd.NewTeamService(deps.teamRepo, deps.contestLookup)
	teamQueries := contestqry.NewTeamService(deps.teamRepo, deps.contestLookup)
	return contesthttp.NewTeamHandler(teamCommands, teamQueries)
}

func buildContestSubmissionHandler(deps *contestModuleDeps, scoreboardCommands *contestcmd.ScoreboardAdminService) (*contesthttp.SubmissionHandler, *contestcmd.SubmissionService) {
	cache := deps.root.Cache()
	cfg := deps.root.Config()

	submissionService := contestcmd.NewSubmissionService(
		deps.contestLookup,
		deps.submissionRepo,
		cache,
		deps.flagValidator,
		deps.teamFinder,
		scoreboardCommands,
		cfg,
	)
	return contesthttp.NewSubmissionHandler(submissionService), submissionService
}
