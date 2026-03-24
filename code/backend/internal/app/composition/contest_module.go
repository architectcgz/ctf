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

func BuildContestModule(root *Root, challenge *ChallengeModule, runtime *RuntimeModule) *ContestModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	repo := contestinfra.NewRepository(db)
	awdRepo := contestinfra.NewAWDRepository(db)
	scoreboardCommands := contestcmd.NewScoreboardAdminService(repo, cache, &cfg.Contest)
	scoreboardQueries := contestqry.NewScoreboardService(repo, cache, &cfg.Contest, log.Named("contest_scoreboard_service"))
	contestCommands := contestcmd.NewContestService(repo, log.Named("contest_service"))
	contestQueries := contestqry.NewContestService(repo, log.Named("contest_service"))
	awdUpdater := contestjobs.NewAWDRoundUpdater(
		awdRepo,
		cache,
		cfg.Contest.AWD,
		cfg.Container.FlagGlobalSecret,
		contestinfra.NewDockerAWDFlagInjector(db, runtime.contest.containerFiles, log.Named("awd_flag_injector")),
		log.Named("awd_round_updater"),
	)
	awdCommands := contestcmd.NewAWDService(
		awdRepo,
		repo,
		cache,
		cfg.Container.FlagGlobalSecret,
		cfg.Contest.AWD,
		log.Named("contest_awd_service"),
		awdUpdater,
	)
	awdQueries := contestqry.NewAWDService(awdRepo, repo)
	contestChallengeRepo := contestinfra.NewChallengeRepository(db)
	contestChallengeCommands := contestcmd.NewChallengeService(contestChallengeRepo, challenge.Catalog, repo)
	contestChallengeQueries := contestqry.NewChallengeService(contestChallengeRepo, challenge.Catalog, repo)
	teamRepo := contestinfra.NewTeamRepository(db)
	teamCommands := contestcmd.NewTeamService(teamRepo, repo)
	teamQueries := contestqry.NewTeamService(teamRepo, repo)
	participationRepo := contestinfra.NewParticipationRepository(db)
	participationCommands := contestcmd.NewParticipationService(repo, participationRepo, teamRepo)
	participationQueries := contestqry.NewParticipationService(repo, participationRepo, teamRepo)
	submissionRepo := contestinfra.NewSubmissionRepository(db)
	submissionService := contestcmd.NewSubmissionService(repo, submissionRepo, cache, challenge.FlagValidator, teamRepo, scoreboardCommands, cfg)
	statusUpdater := contestjobs.NewStatusUpdater(
		repo,
		cache,
		cfg.Contest.StatusUpdateInterval,
		cfg.Contest.StatusUpdateBatchSize,
		cfg.Contest.StatusUpdateLockTTL,
		log.Named("contest_status_updater"),
	)
	root.RegisterBackgroundJob(NewLoopBackgroundJob("contest_status_updater", statusUpdater.Start))
	root.RegisterBackgroundJob(NewLoopBackgroundJob("awd_round_updater", awdUpdater.Start))

	return &ContestModule{
		AWDHandler:           contesthttp.NewAWDHandler(awdCommands, awdQueries),
		ChallengeHandler:     contesthttp.NewChallengeHandler(contestChallengeCommands, contestChallengeQueries),
		Handler:              contesthttp.NewHandler(contestCommands, contestQueries, scoreboardQueries, scoreboardCommands),
		ParticipationHandler: contesthttp.NewParticipationHandler(participationCommands, participationQueries),
		SubmissionHandler:    contesthttp.NewSubmissionHandler(submissionService),
		TeamHandler:          contesthttp.NewTeamHandler(teamCommands, teamQueries),
	}
}
