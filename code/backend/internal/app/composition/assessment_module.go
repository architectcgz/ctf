package composition

import "context"

import assessmentModule "ctf-platform/internal/module/assessment"
import assessmentcontracts "ctf-platform/internal/module/assessment/contracts"

type AssessmentModule struct {
	BackgroundCloser asyncTaskCloser
	Handler          *assessmentModule.Handler
	ProfileService   assessmentcontracts.ProfileService
	Recommendations  assessmentcontracts.RecommendationProvider
	ReportHandler    *assessmentModule.ReportHandler
}

func BuildAssessmentModule(root *Root, challenge *ChallengeModule) *AssessmentModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	repo := assessmentModule.NewRepository(db)
	service := assessmentModule.NewService(repo, cache, cfg.Assessment, log.Named("assessment_service"))
	service.RegisterPracticeEventConsumers(root.Events)
	recommendationService := assessmentModule.NewRecommendationService(
		repo,
		challenge.Catalog,
		cache,
		cfg.Recommendation,
		log.Named("recommendation_service"),
	)
	recommendationService.RegisterPracticeEventConsumers(root.Events)
	reportRepo := assessmentModule.NewReportRepository(db)
	reportService := assessmentModule.NewReportService(reportRepo, service, cfg.Report, log.Named("report_service"))
	cleaner := assessmentModule.NewCleaner(service, log.Named("assessment_cleaner"))
	root.RegisterBackgroundJob(NewBackgroundJob(
		"assessment_cleaner",
		func(context.Context) error {
			return cleaner.Start(cfg.Assessment.FullRebuildCron, cfg.Assessment.FullRebuildTimeout)
		},
		cleaner.Stop,
	))

	return &AssessmentModule{
		BackgroundCloser: reportService,
		Handler:          assessmentModule.NewHandler(service, recommendationService),
		ProfileService:   service,
		Recommendations:  recommendationService,
		ReportHandler:    assessmentModule.NewReportHandler(reportService),
	}
}
