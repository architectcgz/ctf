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
	BackgroundCloser        asyncTaskCloser
	Handler                 *assessmenthttp.Handler
	ProfileService          assessmentcontracts.ProfileService
	Recommendations         assessmentcontracts.RecommendationProvider
	ReportHandler           *assessmenthttp.ReportHandler
	TeacherAWDReviewHandler *assessmenthttp.TeacherAWDReviewHandler
}

type assessmentModuleDeps struct {
	profileRepo        assessmentports.ProfileRepository
	recommendationRepo assessmentports.RecommendationRepository
	reportRepo         assessmentports.ReportRepository
	awdReviewRepo      assessmentports.TeacherAWDReviewRepository
}

type assessmentModuleExternalDeps struct {
	challengeRepo assessmentports.ChallengeRepository
}

func BuildAssessmentModule(root *Root, challenge *ChallengeModule) *AssessmentModule {
	cfg := root.Config()
	deps := buildAssessmentModuleDeps(root)
	externalDeps := buildAssessmentModuleExternalDeps(challenge)

	profileCommandService, profileQueryService := buildAssessmentProfileHandler(root, deps)
	profileCommandService.RegisterPracticeEventConsumers(root.Events)
	recommendationService := buildAssessmentRecommendationHandler(root, cfg, deps, externalDeps)
	recommendationService.RegisterPracticeEventConsumers(root.Events)
	reportService, reportHandler := buildAssessmentReportHandler(root, cfg, deps, profileQueryService)
	teacherAWDReviewHandler := buildAssessmentTeacherAWDReviewHandler(deps)
	cleaner := assessmentcmd.NewCleaner(profileCommandService, root.Logger().Named("assessment_cleaner"))
	root.RegisterBackgroundJob(NewBackgroundJob(
		"assessment_cleaner",
		func(context.Context) error {
			return cleaner.Start(cfg.Assessment.FullRebuildCron, cfg.Assessment.FullRebuildTimeout)
		},
		cleaner.Stop,
	))

	return &AssessmentModule{
		BackgroundCloser:        reportService,
		Handler:                 assessmenthttp.NewHandler(profileQueryService, recommendationService),
		ProfileService:          profileCommandService,
		Recommendations:         recommendationService,
		ReportHandler:           reportHandler,
		TeacherAWDReviewHandler: teacherAWDReviewHandler,
	}
}

func buildAssessmentModuleDeps(root *Root) assessmentModuleDeps {
	repo := assessmentinfra.NewRepository(root.DB())
	return assessmentModuleDeps{
		profileRepo:        repo,
		recommendationRepo: repo,
		reportRepo:         assessmentinfra.NewReportRepository(root.DB()),
		awdReviewRepo:      assessmentinfra.NewTeacherAWDReviewRepository(root.DB()),
	}
}

func buildAssessmentModuleExternalDeps(challenge *ChallengeModule) assessmentModuleExternalDeps {
	return assessmentModuleExternalDeps{
		challengeRepo: challenge.Catalog,
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

func buildAssessmentRecommendationHandler(root *Root, cfg *config.Config, deps assessmentModuleDeps, externalDeps assessmentModuleExternalDeps) *assessmentqry.RecommendationService {
	return assessmentqry.NewRecommendationService(
		deps.recommendationRepo,
		externalDeps.challengeRepo,
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

func buildAssessmentTeacherAWDReviewHandler(deps assessmentModuleDeps) *assessmenthttp.TeacherAWDReviewHandler {
	service := assessmentqry.NewTeacherAWDReviewService(deps.awdReviewRepo)
	return assessmenthttp.NewTeacherAWDReviewHandler(service)
}
