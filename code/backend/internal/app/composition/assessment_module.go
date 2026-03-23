package composition

import (
	"context"

	assessmenthttp "ctf-platform/internal/module/assessment/api/http"
	assessmentapp "ctf-platform/internal/module/assessment/application"
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
	service := assessmentapp.NewService(repo, cache, cfg.Assessment, log.Named("assessment_service"))
	service.RegisterPracticeEventConsumers(root.Events)
	recommendationService := assessmentapp.NewRecommendationService(
		repo,
		challenge.Catalog,
		cache,
		cfg.Recommendation,
		log.Named("recommendation_service"),
	)
	recommendationService.RegisterPracticeEventConsumers(root.Events)
	reportRepo := assessmentinfra.NewReportRepository(db)
	reportService := assessmentapp.NewReportService(reportRepo, service, cfg.Report, log.Named("report_service"))
	cleaner := assessmentapp.NewCleaner(service, log.Named("assessment_cleaner"))
	root.RegisterBackgroundJob(NewBackgroundJob(
		"assessment_cleaner",
		func(context.Context) error {
			return cleaner.Start(cfg.Assessment.FullRebuildCron, cfg.Assessment.FullRebuildTimeout)
		},
		cleaner.Stop,
	))

	return &AssessmentModule{
		BackgroundCloser: reportService,
		Handler:          assessmenthttp.NewHandler(service, recommendationService),
		ProfileService:   service,
		Recommendations:  recommendationService,
		ReportHandler:    assessmenthttp.NewReportHandler(reportService),
	}
}
