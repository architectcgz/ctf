package composition

import contestModule "ctf-platform/internal/module/contest"

type ContestModule struct {
	AWDHandler           *contestModule.AWDHandler
	ChallengeHandler     *contestModule.ChallengeHandler
	Handler              *contestModule.Handler
	ParticipationHandler *contestModule.ParticipationHandler
	SubmissionHandler    *contestModule.SubmissionHandler
	TeamHandler          *contestModule.TeamHandler
}

func BuildContestModule(root *Root, challenge *ChallengeModule, runtime *RuntimeModule) *ContestModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	repo := contestModule.NewRepository(db)
	scoreboardService := contestModule.NewScoreboardService(repo, cache, &cfg.Contest, log.Named("contest_scoreboard_service"))
	contestService := contestModule.NewService(repo, log.Named("contest_service"))
	awdService := contestModule.NewAWDService(
		contestModule.NewAWDRepository(db),
		repo,
		cache,
		cfg.Container.FlagGlobalSecret,
		cfg.Contest.AWD,
		log.Named("contest_awd_service"),
	)
	contestChallengeRepo := contestModule.NewChallengeRepository(db)
	contestChallengeService := contestModule.NewChallengeService(contestChallengeRepo, challenge.Catalog, repo)
	teamRepo := contestModule.NewTeamRepository(db)
	teamService := contestModule.NewTeamService(teamRepo, repo)
	participationRepo := contestModule.NewParticipationRepository(db)
	participationService := contestModule.NewParticipationService(repo, participationRepo, teamRepo)
	submissionRepo := contestModule.NewSubmissionRepository(db)
	submissionService := contestModule.NewSubmissionService(repo, submissionRepo, cache, challenge.FlagValidator, teamRepo, scoreboardService, cfg)
	statusUpdater := contestModule.NewStatusUpdater(
		repo,
		cache,
		cfg.Contest.StatusUpdateInterval,
		cfg.Contest.StatusUpdateBatchSize,
		cfg.Contest.StatusUpdateLockTTL,
		log.Named("contest_status_updater"),
	)
	awdUpdater := contestModule.NewAWDRoundUpdater(
		db,
		cache,
		cfg.Contest.AWD,
		cfg.Container.FlagGlobalSecret,
		contestModule.NewDockerAWDFlagInjector(db, runtime.contest.containerFiles, log.Named("awd_flag_injector")),
		log.Named("awd_round_updater"),
	)
	root.RegisterBackgroundJob(NewLoopBackgroundJob("contest_status_updater", statusUpdater.Start))
	root.RegisterBackgroundJob(NewLoopBackgroundJob("awd_round_updater", awdUpdater.Start))

	return &ContestModule{
		AWDHandler:           contestModule.NewAWDHandler(awdService),
		ChallengeHandler:     contestModule.NewChallengeHandler(contestChallengeService),
		Handler:              contestModule.NewHandler(contestService, scoreboardService),
		ParticipationHandler: contestModule.NewParticipationHandler(participationService),
		SubmissionHandler:    contestModule.NewSubmissionHandler(submissionService),
		TeamHandler:          contestModule.NewTeamHandler(teamService),
	}
}
