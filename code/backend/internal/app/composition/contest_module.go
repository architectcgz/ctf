package composition

import contestModule "ctf-platform/internal/module/contest"

type ContestModule struct {
	AWDHandler           *contestModule.AWDHandler
	ChallengeHandler     *contestModule.ChallengeHandler
	Handler              *contestModule.Handler
	ParticipationHandler *contestModule.ParticipationHandler
	Repository           contestModule.Repository
	SubmissionHandler    *contestModule.SubmissionHandler
	TeamHandler          *contestModule.TeamHandler
}

func BuildContestModule(root *Root, challenge *ChallengeModule) *ContestModule {
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
	contestChallengeService := contestModule.NewChallengeService(contestChallengeRepo, challenge.Repository, repo)
	teamRepo := contestModule.NewTeamRepository(db)
	teamService := contestModule.NewTeamService(teamRepo, repo)
	participationRepo := contestModule.NewParticipationRepository(db)
	participationService := contestModule.NewParticipationService(repo, participationRepo, teamRepo)
	submissionRepo := contestModule.NewSubmissionRepository(db)
	submissionService := contestModule.NewSubmissionService(repo, submissionRepo, cache, challenge.FlagService, teamRepo, scoreboardService, cfg)

	return &ContestModule{
		AWDHandler:           contestModule.NewAWDHandler(awdService),
		ChallengeHandler:     contestModule.NewChallengeHandler(contestChallengeService),
		Handler:              contestModule.NewHandler(contestService, scoreboardService),
		ParticipationHandler: contestModule.NewParticipationHandler(participationService),
		Repository:           repo,
		SubmissionHandler:    contestModule.NewSubmissionHandler(submissionService),
		TeamHandler:          contestModule.NewTeamHandler(teamService),
	}
}
