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
	CheckerRunner         contestports.CheckerRunner
	RuntimeProbe          challengeports.ChallengeRuntimeProbe
	EndedRuntimeCleaner   contestports.ContestEndedRuntimeCleaner
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
	awdQuery              *contestinfra.AWDQueryRepository
	challengeRepo         *contestinfra.ChallengeRepository
	teamRepo              *contestinfra.TeamRepository
	teamFinder            contestports.ContestTeamFinder
	teamCommand           *contestinfra.TeamCommandAdapter
	teamQuery             *contestinfra.TeamQueryAdapter
	participationLookup   *contestinfra.ParticipationRegistrationRepository
	submissionLookup      *contestinfra.SubmissionRegistrationRepository
	challengeCatalog      contestports.ContestChallengeCatalog
	challengeCatalogCmd   contestports.ContestChallengeCatalog
	awdChallengeQueryRepo challengeports.AWDChallengeQueryRepository
	awdChallengeQueryCmd  challengeports.AWDChallengeQueryRepository
	previewChallengeRepo  challengeports.AWDChallengeQueryRepository
	imageRepo             challengecontracts.ImageStore
	previewImageRepo      challengecontracts.ImageStore
	flagValidator         challengecontracts.FlagValidator
	containerFiles        contestports.AWDContainerFileWriter
	checkerRunner         contestports.CheckerRunner
	runtimeProbe          challengeports.ChallengeRuntimeProbe
	endedRuntimeCleaner   contestports.ContestEndedRuntimeCleaner
}

func Build(deps Deps) *Module {
	internalDeps := newModuleDeps(deps)

	handler, scoreboardCommands, statusUpdater := buildCoreHandler(internalDeps)
	awdHandler, awdUpdater := buildAWDHandler(internalDeps)
	challengeHandler := buildChallengeHandler(internalDeps)
	participationHandler := buildParticipationHandler(internalDeps)
	teamHandler := buildTeamHandler(internalDeps)
	submissionHandler := buildSubmissionHandler(internalDeps, scoreboardCommands)

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
	}
}

func newModuleDeps(deps Deps) *moduleDeps {
	contestRepo := contestinfra.NewRepository(deps.DB)
	awdRepo := contestinfra.NewAWDRepository(deps.DB)
	challengeRepo := contestinfra.NewChallengeRepository(deps.DB)
	teamRepo := contestinfra.NewTeamRepository(deps.DB)
	participationRepo := contestinfra.NewParticipationRepository(deps.DB)
	submissionRepo := contestinfra.NewSubmissionRepository(deps.DB)
	teamFinder := contestinfra.NewTeamFinderRepository(teamRepo)
	teamCommand := contestinfra.NewTeamCommandAdapter(teamRepo)
	teamQuery := contestinfra.NewTeamQueryAdapter(teamRepo)
	participationLookup := contestinfra.NewParticipationRegistrationRepository(participationRepo)
	submissionLookup := contestinfra.NewSubmissionRegistrationRepository(submissionRepo)
	previewRuntimeChallengeLookup := contestinfra.NewAWDPreviewRuntimeChallengeRepository(deps.AWDChallengeQueryRepo)
	challengeCatalogCmd := contestinfra.NewContestChallengeLookupAdapter(deps.ChallengeCatalog)
	awdChallengeQueryCmd := contestinfra.NewContestAWDChallengeLookupAdapter(deps.AWDChallengeQueryRepo)

	return &moduleDeps{
		input:                 deps,
		contestCommands:       contestRepo,
		contestLookup:         contestRepo,
		contestList:           contestRepo,
		contestScoreboard:     contestRepo,
		contestAdmin:          contestRepo,
		contestStatus:         contestRepo,
		awdRepo:               awdRepo,
		awdQuery:              contestinfra.NewAWDQueryRepository(awdRepo),
		challengeRepo:         challengeRepo,
		teamRepo:              teamRepo,
		teamFinder:            teamFinder,
		teamCommand:           teamCommand,
		teamQuery:             teamQuery,
		participationLookup:   participationLookup,
		submissionLookup:      submissionLookup,
		challengeCatalog:      challengeCatalogCmd,
		challengeCatalogCmd:   challengeCatalogCmd,
		awdChallengeQueryRepo: deps.AWDChallengeQueryRepo,
		awdChallengeQueryCmd:  awdChallengeQueryCmd,
		previewChallengeRepo:  previewRuntimeChallengeLookup,
		imageRepo:             deps.ImageRepo,
		previewImageRepo:      deps.ImageRepo,
		flagValidator:         deps.FlagValidator,
		containerFiles:        deps.ContainerFiles,
		checkerRunner:         deps.CheckerRunner,
		runtimeProbe:          deps.RuntimeProbe,
		endedRuntimeCleaner:   deps.EndedRuntimeCleaner,
	}
}
