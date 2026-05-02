package runtime

import (
	"context"

	redislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	"ctf-platform/internal/dto"
	assessmenthttp "ctf-platform/internal/module/assessment/api/http"
	assessmentcmd "ctf-platform/internal/module/assessment/application/commands"
	assessmentqry "ctf-platform/internal/module/assessment/application/queries"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	assessmentinfra "ctf-platform/internal/module/assessment/infrastructure"
	assessmentports "ctf-platform/internal/module/assessment/ports"
	platformevents "ctf-platform/internal/platform/events"
)

type BackgroundTaskCloser interface {
	Close(ctx context.Context) error
}

type BackgroundJob struct {
	Name  string
	Start func(context.Context) error
	Stop  func(context.Context) error
}

type Module struct {
	BackgroundJobs          []BackgroundJob
	BackgroundTasks         BackgroundTaskCloser
	Handler                 *assessmenthttp.Handler
	ProfileService          assessmentcontracts.ProfileService
	Recommendations         assessmentcontracts.RecommendationProvider
	ReportHandler           *assessmenthttp.ReportHandler
	TeacherAWDReviewHandler *assessmenthttp.TeacherAWDReviewHandler
}

type Deps struct {
	AppContext    context.Context
	Config        *config.Config
	Logger        *zap.Logger
	DB            *gorm.DB
	Cache         *redislib.Client
	Events        platformevents.Bus
	ChallengeRepo assessmentports.ChallengeRepository
}

type moduleDeps struct {
	input              Deps
	profileRepo        assessmentports.ProfileRepository
	recommendationRepo assessmentports.RecommendationRepository
	reportRepo         assessmentports.ReportRepository
	awdReviewRepo      assessmentports.TeacherAWDReviewRepository
	challengeRepo      assessmentports.ChallengeRepository
}

func Build(deps Deps) *Module {
	internalDeps := newModuleDeps(deps)

	profileCommandService, profileQueryService := buildProfileHandler(internalDeps)
	profileCommandService.RegisterPracticeEventConsumers(deps.Events)
	profileCommandService.RegisterContestEventConsumers(deps.Events)

	recommendationService := buildRecommendationHandler(internalDeps)
	recommendationService.RegisterPracticeEventConsumers(deps.Events)
	recommendationService.RegisterContestEventConsumers(deps.Events)

	reportService, reportHandler := buildReportHandler(internalDeps, profileQueryService)
	teacherAWDReviewHandler := buildTeacherAWDReviewHandler(internalDeps, reportService)
	cleaner := assessmentcmd.NewCleaner(profileCommandService, deps.Logger.Named("assessment_cleaner"))

	return &Module{
		BackgroundJobs: []BackgroundJob{
			{
				Name: "assessment_cleaner",
				Start: func(ctx context.Context) error {
					return cleaner.Start(ctx, deps.Config.Assessment.FullRebuildCron, deps.Config.Assessment.FullRebuildTimeout)
				},
				Stop: cleaner.Stop,
			},
		},
		BackgroundTasks:         reportService,
		Handler:                 assessmenthttp.NewHandler(profileQueryService, recommendationService),
		ProfileService:          profileCommandService,
		Recommendations:         recommendationService,
		ReportHandler:           reportHandler,
		TeacherAWDReviewHandler: teacherAWDReviewHandler,
	}
}

func newModuleDeps(deps Deps) moduleDeps {
	repo := assessmentinfra.NewRepository(deps.DB)
	return moduleDeps{
		input:              deps,
		profileRepo:        repo,
		recommendationRepo: repo,
		reportRepo:         assessmentinfra.NewReportRepository(deps.DB),
		awdReviewRepo:      assessmentinfra.NewTeacherAWDReviewRepository(deps.DB),
		challengeRepo:      deps.ChallengeRepo,
	}
}

func buildProfileHandler(deps moduleDeps) (*assessmentcmd.Service, *assessmentqry.ProfileService) {
	profileCommandService := assessmentcmd.NewProfileService(
		deps.profileRepo,
		deps.input.Cache,
		deps.input.Config.Assessment,
		deps.input.Logger.Named("assessment_service"),
	)
	profileQueryService := assessmentqry.NewProfileService(deps.profileRepo)
	return profileCommandService, profileQueryService
}

func buildRecommendationHandler(deps moduleDeps) *assessmentqry.RecommendationService {
	return assessmentqry.NewRecommendationService(
		deps.recommendationRepo,
		deps.challengeRepo,
		deps.input.Cache,
		deps.input.Config.Recommendation,
		deps.input.Logger.Named("recommendation_service"),
	)
}

func buildReportHandler(deps moduleDeps, profileQueryService assessmentports.AssessmentProfileReader) (*assessmentcmd.ReportService, *assessmenthttp.ReportHandler) {
	reportService := assessmentcmd.NewReportService(
		deps.reportRepo,
		profileQueryService,
		deps.input.Config.Report,
		deps.input.Logger.Named("report_service"),
	)
	reportService.StartBackgroundTasks(deps.input.AppContext)
	reportService.SetAWDReviewExportBuilder(
		assessmentcmd.NewAWDReviewExportBuilder(
			assessmentqry.NewTeacherAWDReviewService(deps.awdReviewRepo),
		),
	)
	return reportService, assessmenthttp.NewReportHandler(reportService)
}

func buildTeacherAWDReviewHandler(deps moduleDeps, reportService *assessmentcmd.ReportService) *assessmenthttp.TeacherAWDReviewHandler {
	service := assessmentqry.NewTeacherAWDReviewService(deps.awdReviewRepo)
	return assessmenthttp.NewTeacherAWDReviewHandler(&teacherAWDReviewHandlerService{
		queryService:  service,
		reportService: reportService,
	})
}

type teacherAWDReviewHandlerService struct {
	queryService  *assessmentqry.TeacherAWDReviewService
	reportService *assessmentcmd.ReportService
}

func (s *teacherAWDReviewHandlerService) ListContests(ctx context.Context, requesterID int64) (*dto.TeacherAWDReviewContestListResp, error) {
	return s.queryService.ListContests(ctx, requesterID)
}

func (s *teacherAWDReviewHandlerService) GetContestArchive(ctx context.Context, requesterID, contestID int64, req assessmentqry.GetTeacherAWDReviewArchiveInput) (*dto.TeacherAWDReviewArchiveResp, error) {
	return s.queryService.GetContestArchive(ctx, requesterID, contestID, req)
}

func (s *teacherAWDReviewHandlerService) CreateTeacherAWDReviewArchive(ctx context.Context, requesterID, contestID int64, req assessmentcmd.CreateTeacherAWDReviewExportInput) (*dto.ReportExportData, error) {
	return s.reportService.CreateTeacherAWDReviewArchive(ctx, requesterID, contestID, req)
}

func (s *teacherAWDReviewHandlerService) CreateTeacherAWDReviewReport(ctx context.Context, requesterID, contestID int64, req assessmentcmd.CreateTeacherAWDReviewExportInput) (*dto.ReportExportData, error) {
	return s.reportService.CreateTeacherAWDReviewReport(ctx, requesterID, contestID, req)
}
