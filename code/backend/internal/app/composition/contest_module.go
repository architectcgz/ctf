package composition

import (
	contestModule "ctf-platform/internal/module/contest"
	contesthttp "ctf-platform/internal/module/contest/api/http"
	contestapp "ctf-platform/internal/module/contest/application"
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

func BuildContestModule(root *Root, challenge *ChallengeModule, runtime *RuntimeModule) *ContestModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	repo := contestinfra.NewRepository(db)
	scoreboardService := contestapp.NewScoreboardService(repo, cache, &cfg.Contest, log.Named("contest_scoreboard_service"))
	contestService := contestapp.NewService(repo, log.Named("contest_service"))
	awdService := contestModule.NewAWDService(
		contestModule.NewAWDRepository(db),
		repo,
		cache,
		cfg.Container.FlagGlobalSecret,
		cfg.Contest.AWD,
		log.Named("contest_awd_service"),
	)
	contestChallengeRepo := contestinfra.NewChallengeRepository(db)
	contestChallengeService := contestapp.NewChallengeService(contestChallengeRepo, challenge.Catalog, repo)
	teamRepo := contestModule.NewTeamRepository(db)
	teamService := contestModule.NewTeamService(teamRepo, repo)
	participationRepo := contestinfra.NewParticipationRepository(db)
	participationService := contestapp.NewParticipationService(repo, participationRepo, teamRepo)
	submissionRepo := contestModule.NewSubmissionRepository(db)
	submissionService := contestModule.NewSubmissionService(repo, submissionRepo, cache, challenge.FlagValidator, teamRepo, scoreboardService, cfg)
	statusUpdater := contestapp.NewStatusUpdater(
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
		AWDHandler:           contesthttp.NewAWDHandler(awdService),
		ChallengeHandler:     contesthttp.NewChallengeHandler(contestChallengeService),
		Handler:              contesthttp.NewHandler(contestService, scoreboardService),
		ParticipationHandler: contesthttp.NewParticipationHandler(participationService),
		SubmissionHandler:    contesthttp.NewSubmissionHandler(submissionService),
		TeamHandler:          contesthttp.NewTeamHandler(teamService),
	}
}
