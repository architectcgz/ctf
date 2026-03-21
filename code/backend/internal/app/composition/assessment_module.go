package composition

import assessmentModule "ctf-platform/internal/module/assessment"

type AssessmentModule struct {
	Handler               *assessmentModule.Handler
	RecommendationService *assessmentModule.RecommendationService
	ReportHandler         *assessmentModule.ReportHandler
	ReportService         *assessmentModule.ReportService
	Service               *assessmentModule.Service
}

func BuildAssessmentModule(root *Root, challenge *ChallengeModule) *AssessmentModule {
	cfg := root.Config()
	log := root.Logger()
	db := root.DB()
	cache := root.Cache()

	repo := assessmentModule.NewRepository(db)
	service := assessmentModule.NewService(repo, cache, cfg.Assessment, log.Named("assessment_service"))
	recommendationService := assessmentModule.NewRecommendationService(
		repo,
		challenge.Repository,
		cache,
		cfg.Recommendation,
		log.Named("recommendation_service"),
	)
	reportRepo := assessmentModule.NewReportRepository(db)
	reportService := assessmentModule.NewReportService(reportRepo, service, cfg.Report, log.Named("report_service"))

	return &AssessmentModule{
		Handler:               assessmentModule.NewHandler(service, recommendationService),
		RecommendationService: recommendationService,
		ReportHandler:         assessmentModule.NewReportHandler(reportService),
		ReportService:         reportService,
		Service:               service,
	}
}
