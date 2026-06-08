package runtime

import (
	contesthttp "ctf-platform/internal/module/contest/api/http"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestjobs "ctf-platform/internal/module/contest/application/jobs"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
)

func buildCoreHandler(deps *moduleDeps) (*contesthttp.Handler, *contestcmd.ScoreboardAdminService, *contestjobs.StatusUpdater) {
	cfg := deps.input.Config
	log := deps.input.Logger
	cache := deps.input.Cache
	statusSideEffects := contestinfra.NewContestStatusSideEffectStore(cache, deps.endedRuntimeCleaner)
	statusUpdateLockStore := contestinfra.NewContestStatusUpdateLockStore(cache)
	scoreboardStateStore := contestinfra.NewContestScoreboardStateStore(cache)

	scoreboardCommands := contestcmd.NewScoreboardAdminService(deps.contestAdmin, scoreboardStateStore, &cfg.Contest)
	scoreboardCommands.SetStatusSideEffectStore(statusSideEffects)
	scoreboardCommands.SetEventBus(deps.input.Events)
	scoreboardQueries := contestqry.NewScoreboardService(deps.contestScoreboard, scoreboardStateStore, &cfg.Contest, log.Named("contest_scoreboard_service"))
	contestCommands := contestcmd.NewContestService(deps.contestCommands, deps.awdRepo, log.Named("contest_service"))
	contestCommands.SetStatusSideEffectStore(statusSideEffects)
	contestQueries := contestqry.NewContestService(deps.contestList, log.Named("contest_service"))
	readinessQueries := contestqry.NewAWDService(deps.awdQuery, deps.contestLookup)
	statusUpdater := contestjobs.NewStatusUpdater(
		deps.contestStatus,
		cfg.Contest.StatusUpdateInterval,
		cfg.Contest.StatusUpdateBatchSize,
		cfg.Contest.StatusUpdateLockTTL,
		log.Named("contest_status_updater"),
		deps.awdRepo,
	)
	statusUpdater.SetStatusSideEffectStore(statusSideEffects)
	statusUpdater.SetStatusUpdateLockStore(statusUpdateLockStore)

	return contesthttp.NewHandler(contestCommands, contestQueries, readinessQueries, scoreboardQueries, scoreboardCommands), scoreboardCommands, statusUpdater
}

func buildAWDHandler(deps *moduleDeps) (*contesthttp.AWDHandler, *contestjobs.AWDRoundUpdater) {
	cfg := deps.input.Config
	log := deps.input.Logger
	cache := deps.input.Cache
	db := deps.input.DB
	scoreboardCache := contestinfra.NewScoreboardCache(db, cache)
	awdRoundStateStore := contestinfra.NewAWDRoundStateStore(cache)
	previewTokenStore := contestinfra.NewAWDCheckerPreviewTokenStore(cache)
	awdJobRepo := contestinfra.NewAWDJobRepository(deps.awdRepo)

	awdUpdater := contestjobs.NewAWDRoundUpdater(
		awdJobRepo,
		awdRoundStateStore,
		cfg.Contest.AWD,
		cfg.Container.FlagGlobalSecret,
		contestinfra.NewDockerAWDFlagInjector(db, deps.containerFiles, log.Named("awd_flag_injector")),
		log.Named("awd_round_updater"),
		scoreboardCache,
	)
	awdUpdater.SetHTTPRuntime(contestinfra.NewAWDHTTPRuntimeAdapter(nil, cfg.Contest.AWD.CheckerTimeout))
	if deps.checkerRunner != nil {
		awdUpdater.SetCheckerRunner(deps.checkerRunner)
	}
	awdCommandRepo := contestinfra.NewAWDCommandRepository(deps.awdRepo)
	awdCommandRoundManager := contestinfra.NewAWDRoundManagerAdapter(awdUpdater)
	awdCommands := contestcmd.NewAWDService(
		awdCommandRepo,
		deps.contestLookup,
		awdRoundStateStore,
		previewTokenStore,
		cfg.Container.FlagGlobalSecret,
		cfg.Contest.AWD,
		log.Named("contest_awd_service"),
		awdCommandRoundManager,
		deps.previewImageRepo,
		deps.previewChallengeRepo,
		deps.runtimeProbe,
		scoreboardCache,
	)
	awdCommands.SetFlagInjector(contestinfra.NewDockerAWDFlagInjector(db, deps.containerFiles, log.Named("awd_flag_injector")))
	awdCommands.SetEventBus(deps.input.Events)
	awdQueries := contestqry.NewAWDService(deps.awdQuery, deps.contestLookup)
	awdServiceCommands := contestcmd.NewContestAWDServiceService(
		awdCommandRepo,
		deps.contestLookup,
		deps.challengeRepo,
		deps.challengeCatalogCmd,
		deps.awdChallengeQueryCmd,
		previewTokenStore,
	)
	awdServiceQueries := contestqry.NewContestAWDServiceQueryService(deps.awdRepo, deps.contestLookup)

	return contesthttp.NewAWDHandler(awdCommands, awdQueries, awdServiceCommands, awdServiceQueries), awdUpdater
}

func buildChallengeHandler(deps *moduleDeps) *contesthttp.ChallengeHandler {
	contestChallengeCommands := contestcmd.NewChallengeService(deps.challengeRepo, deps.challengeCatalogCmd, deps.contestLookup, deps.awdRepo)
	contestChallengeQueries := contestqry.NewChallengeService(deps.challengeRepo, deps.challengeCatalog, deps.contestLookup, deps.awdRepo)
	return contesthttp.NewChallengeHandler(contestChallengeCommands, contestChallengeQueries)
}

func buildParticipationHandler(deps *moduleDeps) *contesthttp.ParticipationHandler {
	participationCommands := contestcmd.NewParticipationService(deps.contestLookup, deps.participationLookup, deps.teamFinder)
	participationCommands.SetEventBus(deps.input.Events)
	participationQueries := contestqry.NewParticipationService(deps.contestLookup, deps.participationLookup, deps.teamFinder)
	return contesthttp.NewParticipationHandler(participationCommands, participationQueries)
}

func buildTeamHandler(deps *moduleDeps) *contesthttp.TeamHandler {
	teamCommands := contestcmd.NewTeamService(deps.teamCommand, deps.contestLookup)
	teamQueries := contestqry.NewTeamService(deps.teamQuery, deps.contestLookup)
	return contesthttp.NewTeamHandler(teamCommands, teamQueries)
}

func buildSubmissionHandler(deps *moduleDeps, scoreboardCommands *contestcmd.ScoreboardAdminService) *contesthttp.SubmissionHandler {
	cfg := deps.input.Config
	rateLimitStore := contestinfra.NewContestSubmissionRateLimitStore(deps.input.Cache, cfg.RateLimit.RedisKeyPrefix)

	submissionService := contestcmd.NewSubmissionService(
		deps.contestLookup,
		deps.submissionLookup,
		rateLimitStore,
		deps.flagValidator,
		deps.teamFinder,
		scoreboardCommands,
		cfg,
	)
	submissionService.SetEventBus(deps.input.Events)
	return contesthttp.NewSubmissionHandler(submissionService)
}
