package composition

import (
	"context"

	assessmenthttp "ctf-platform/internal/module/assessment/api/http"
	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
)

type AssessmentModule struct {
	BackgroundCloser asyncTaskCloser
	Handler          *assessmenthttp.Handler
	ProfileService   assessmentcontracts.ProfileService
	Recommendations  assessmentcontracts.RecommendationProvider
	ReportHandler    *assessmenthttp.ReportHandler
}

func BuildAssessmentModule(root *Root, challenge *ChallengeModule) *AssessmentModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	repo := assessmentinfra.NewRepository(db)
	profileCommandService := assessmentcmd.NewProfileService(repo, cache, cfg.Assessment, log.Named("assessment_service"))
	profileCommandService.RegisterPracticeEventConsumers(root.Events)
	profileQueryService := assessmentqry.NewProfileService(repo)
	recommendationService := assessmentqry.NewRecommendationService(
		repo,
		challenge.Catalog,
		cache,
		cfg.Recommendation,
		log.Named("recommendation_service"),
	)
	recommendationService.RegisterPracticeEventConsumers(root.Events)
	reportRepo := assessmentinfra.NewReportRepository(db)
	reportService := assessmentcmd.NewReportService(reportRepo, profileQueryService, cfg.Report, log.Named("report_service"))
	cleaner := assessmentcmd.NewCleaner(profileCommandService, log.Named("assessment_cleaner"))
	root.RegisterBackgroundJob(NewBackgroundJob(
		"assessment_cleaner",
		func(context.Context) error {
			return cleaner.Start(cfg.Assessment.FullRebuildCron, cfg.Assessment.FullRebuildTimeout)
		},
		cleaner.Stop,
	))

	return &AssessmentModule{
		BackgroundCloser: reportService,
		Handler:          assessmenthttp.NewHandler(profileQueryService, recommendationService),
		ProfileService:   profileCommandService,
		Recommendations:  recommendationService,
		ReportHandler:    assessmenthttp.NewReportHandler(reportService),
	}
}
