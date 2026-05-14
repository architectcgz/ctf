package runtime

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	teachinghttp "ctf-platform/internal/module/teaching_query/api/http"
	teachingqueries "ctf-platform/internal/module/teaching_query/application/queries"
	queryinfra "ctf-platform/internal/module/teaching_query/infrastructure"
	queryports "ctf-platform/internal/module/teaching_query/ports"
)

type Module struct {
	Handler *teachinghttp.Handler
}

type Deps struct {
	Config          *config.Config
	Logger          *zap.Logger
	DB              *gorm.DB
	Users           queryports.TeachingUserLookupRepository
	Recommendations assessmentcontracts.RecommendationProvider
}

type moduleDeps struct {
	input Deps
	users queryports.TeachingUserLookupRepository
	repo  interface {
		queryports.TeachingClassQueryRepository
		queryports.TeachingStudentDirectoryRepository
		queryports.TeachingStudentProfileRepository
		queryports.TeachingStudentActivityRepository
		queryports.TeachingClassInsightRepository
		queryports.TeachingOverviewRepository
	}
	recommendations assessmentcontracts.RecommendationProvider
}

func Build(deps Deps) *Module {
	internalDeps := newModuleDeps(deps)
	service := buildQueryService(internalDeps)

	return &Module{
		Handler: teachinghttp.NewHandler(
			service,
			buildOverviewService(internalDeps),
			buildClassInsightService(internalDeps),
			buildStudentReviewService(internalDeps),
		),
	}
}

func newModuleDeps(deps Deps) moduleDeps {
	return moduleDeps{
		input:           deps,
		users:           deps.Users,
		repo:            queryinfra.NewRepository(deps.DB),
		recommendations: deps.Recommendations,
	}
}

func buildQueryService(deps moduleDeps) teachingqueries.Service {
	cfg := deps.input.Config
	return teachingqueries.NewQueryService(
		deps.users,
		deps.repo,
		cfg.Pagination,
	)
}

func buildOverviewService(deps moduleDeps) teachingqueries.OverviewService {
	return teachingqueries.NewOverviewService(deps.users, deps.repo)
}

func buildClassInsightService(deps moduleDeps) teachingqueries.ClassInsightService {
	return teachingqueries.NewClassInsightService(
		deps.users,
		deps.repo,
		deps.recommendations,
		deps.input.Logger.Named("teaching_query_class_insight_service"),
	)
}

func buildStudentReviewService(deps moduleDeps) teachingqueries.StudentReviewService {
	return teachingqueries.NewStudentReviewService(
		deps.users,
		deps.repo,
		deps.recommendations,
	)
}
