package composition

import (
	"context"

	"ctf-platform/internal/config"
	assessmenthttp "ctf-platform/internal/module/assessment/api/http"
	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
	assessmentports "ctf-platform/internal/module/assessment/ports"
)

type AssessmentModule struct {
	BackgroundCloser asyncTaskCloser
	Handler          *assessmenthttp.Handler
	ProfileService   assessmentcontracts.ProfileService
	Recommendations  assessmentcontracts.RecommendationProvider
	ReportHandler    *assessmenthttp.ReportHandler
}

type assessmentModuleDeps struct {
	profileRepo        assessmentports.ProfileRepository
	recommendationRepo assessmentports.RecommendationRepository
	challengeRepo      assessmentports.ChallengeRepository
	reportRepo         assessmentports.ReportRepository
}

func BuildAssessmentModule(root *Root, challenge *ChallengeModule) *AssessmentModule {
	cfg := root.Config()
	deps := buildAssessmentModuleDeps(root, challenge)

	profileCommandService, profileQueryService := buildAssessmentProfileHandler(root, deps)
	profileCommandService.RegisterPracticeEventConsumers(root.Events)
	recommendationService := buildAssessmentRecommendationHandler(root, cfg, deps)
	recommendationService.RegisterPracticeEventConsumers(root.Events)
	reportService, reportHandler := buildAssessmentReportHandler(root, cfg, deps, profileQueryService)
	cleaner := assessmentcmd.NewCleaner(profileCommandService, root.Logger().Named("assessment_cleaner"))
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
		ReportHandler:    reportHandler,
	}
}

func buildAssessmentModuleDeps(root *Root, challenge *ChallengeModule) assessmentModuleDeps {
	repo := assessmentinfra.NewRepository(root.DB())
	return assessmentModuleDeps{
		profileRepo:        repo,
		recommendationRepo: repo,
		challengeRepo:      challenge.Catalog,
		reportRepo:         assessmentinfra.NewReportRepository(root.DB()),
	}
}

func buildAssessmentProfileHandler(root *Root, deps assessmentModuleDeps) (*assessmentcmd.Service, *assessmentqry.ProfileService) {
	profileCommandService := assessmentcmd.NewProfileService(
		deps.profileRepo,
		root.Cache(),
		root.Config().Assessment,
		root.Logger().Named("assessment_service"),
	)
	profileQueryService := assessmentqry.NewProfileService(deps.profileRepo)
	return profileCommandService, profileQueryService
}

func buildAssessmentRecommendationHandler(root *Root, cfg *config.Config, deps assessmentModuleDeps) *assessmentqry.RecommendationService {
	return assessmentqry.NewRecommendationService(
		deps.recommendationRepo,
		deps.challengeRepo,
		root.Cache(),
		cfg.Recommendation,
		root.Logger().Named("recommendation_service"),
	)
}

func buildAssessmentReportHandler(root *Root, cfg *config.Config, deps assessmentModuleDeps, profileQueryService assessmentports.AssessmentProfileReader) (*assessmentcmd.ReportService, *assessmenthttp.ReportHandler) {
	reportService := assessmentcmd.NewReportService(
		deps.reportRepo,
		profileQueryService,
		cfg.Report,
		root.Logger().Named("report_service"),
	)
	return reportService, assessmenthttp.NewReportHandler(reportService)
}
