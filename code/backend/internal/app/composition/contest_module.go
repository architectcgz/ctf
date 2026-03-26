package composition

import (
	contesthttp "ctf-platform/internal/module/contest/api/http"
	contestcmd "ctf-platform/internal/module/contest/application/commands"
	contestjobs "ctf-platform/internal/module/contest/application/jobs"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	contestinfra "ctf-platform/internal/module/contest/infrastructure"
)

type ContestModule struct {
	AWDHandler           *contesthttp.AWDHandler
	ChallengeHandler     *contesthttp.ChallengeHandler
	Handler              *contesthttp.Handler
	ParticipationHandler *contesthttp.ParticipationHandler
	SubmissionHandler    *contesthttp.SubmissionHandler
	TeamHandler          *contesthttp.TeamHandler
}

type contestModuleDeps struct {
	root              *Root
	challenge         *ChallengeModule
	runtime           *RuntimeModule
	contestRepo       *contestinfra.Repository
	awdRepo           *contestinfra.AWDRepository
	challengeRepo     *contestinfra.ChallengeRepository
	teamRepo          *contestinfra.TeamRepository
	participationRepo *contestinfra.ParticipationRepository
	submissionRepo    *contestinfra.SubmissionRepository
}

func BuildContestModule(root *Root, challenge *ChallengeModule, runtime *RuntimeModule) *ContestModule {
	db := root.DB()

	deps := &contestModuleDeps{
		root:              root,
		challenge:         challenge,
		runtime:           runtime,
		contestRepo:       contestinfra.NewRepository(db),
		awdRepo:           contestinfra.NewAWDRepository(db),
		challengeRepo:     contestinfra.NewChallengeRepository(db),
		teamRepo:          contestinfra.NewTeamRepository(db),
		participationRepo: contestinfra.NewParticipationRepository(db),
		submissionRepo:    contestinfra.NewSubmissionRepository(db),
	}

	handler, scoreboardCommands, statusUpdater := buildContestCoreHandler(deps)
	awdHandler, awdUpdater := buildContestAWDHandler(deps)
	challengeHandler := buildContestChallengeHandler(deps)
	participationHandler := buildContestParticipationHandler(deps)
	teamHandler := buildContestTeamHandler(deps)
	submissionHandler := buildContestSubmissionHandler(deps, scoreboardCommands)

	root.RegisterBackgroundJob(NewLoopBackgroundJob("contest_status_updater", statusUpdater.Start))
	root.RegisterBackgroundJob(NewLoopBackgroundJob("awd_round_updater", awdUpdater.Start))

	return &ContestModule{
		AWDHandler:           awdHandler,
		ChallengeHandler:     challengeHandler,
		Handler:              handler,
		ParticipationHandler: participationHandler,
		SubmissionHandler:    submissionHandler,
		TeamHandler:          teamHandler,
	}
}

func buildContestCoreHandler(deps *contestModuleDeps) (*contesthttp.Handler, *contestcmd.ScoreboardAdminService, *contestjobs.StatusUpdater) {
	cfg := deps.root.Config()
	log := deps.root.Logger()
	cache := deps.root.Cache()

	scoreboardCommands := contestcmd.NewScoreboardAdminService(deps.contestRepo, cache, &cfg.Contest)
	scoreboardQueries := contestqry.NewScoreboardService(deps.contestRepo, cache, &cfg.Contest, log.Named("contest_scoreboard_service"))
	contestCommands := contestcmd.NewContestService(deps.contestRepo, log.Named("contest_service"))
	contestQueries := contestqry.NewContestService(deps.contestRepo, log.Named("contest_service"))
	statusUpdater := contestjobs.NewStatusUpdater(
		deps.contestRepo,
		cache,
		cfg.Contest.StatusUpdateInterval,
		cfg.Contest.StatusUpdateBatchSize,
		cfg.Contest.StatusUpdateLockTTL,
		log.Named("contest_status_updater"),
	)

	return contesthttp.NewHandler(contestCommands, contestQueries, scoreboardQueries, scoreboardCommands), scoreboardCommands, statusUpdater
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
		contestinfra.NewDockerAWDFlagInjector(db, deps.runtime.contest.containerFiles, log.Named("awd_flag_injector")),
		log.Named("awd_round_updater"),
	)
	awdCommands := contestcmd.NewAWDService(
		deps.awdRepo,
		deps.contestRepo,
		cache,
		cfg.Container.FlagGlobalSecret,
		cfg.Contest.AWD,
		log.Named("contest_awd_service"),
		awdUpdater,
	)
	awdQueries := contestqry.NewAWDService(deps.awdRepo, deps.contestRepo)

	return contesthttp.NewAWDHandler(awdCommands, awdQueries), awdUpdater
}

func buildContestChallengeHandler(deps *contestModuleDeps) *contesthttp.ChallengeHandler {
	contestChallengeCommands := contestcmd.NewChallengeService(deps.challengeRepo, deps.challenge.Catalog, deps.contestRepo)
	contestChallengeQueries := contestqry.NewChallengeService(deps.challengeRepo, deps.challenge.Catalog, deps.contestRepo)
	return contesthttp.NewChallengeHandler(contestChallengeCommands, contestChallengeQueries)
}

func buildContestParticipationHandler(deps *contestModuleDeps) *contesthttp.ParticipationHandler {
	participationCommands := contestcmd.NewParticipationService(deps.contestRepo, deps.participationRepo, deps.teamRepo)
	participationQueries := contestqry.NewParticipationService(deps.contestRepo, deps.participationRepo, deps.teamRepo)
	return contesthttp.NewParticipationHandler(participationCommands, participationQueries)
}

func buildContestTeamHandler(deps *contestModuleDeps) *contesthttp.TeamHandler {
	teamCommands := contestcmd.NewTeamService(deps.teamRepo, deps.contestRepo)
	teamQueries := contestqry.NewTeamService(deps.teamRepo, deps.contestRepo)
	return contesthttp.NewTeamHandler(teamCommands, teamQueries)
}

func buildContestSubmissionHandler(deps *contestModuleDeps, scoreboardCommands *contestcmd.ScoreboardAdminService) *contesthttp.SubmissionHandler {
	cache := deps.root.Cache()
	cfg := deps.root.Config()

	submissionService := contestcmd.NewSubmissionService(
		deps.contestRepo,
		deps.submissionRepo,
		cache,
		deps.challenge.FlagValidator,
		deps.teamRepo,
		scoreboardCommands,
		cfg,
	)
	return contesthttp.NewSubmissionHandler(submissionService)
}
