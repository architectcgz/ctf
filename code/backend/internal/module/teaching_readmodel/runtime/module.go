package runtime

import (
	"go.uber.org/zap"
	"gorm.io/gorm"

	"ctf-platform/internal/config"
	assessmentcontracts "ctf-platform/internal/module/assessment/contracts"
	teachinghttp "ctf-platform/internal/module/teaching_readmodel/api/http"
	readmodelqueries "ctf-platform/internal/module/teaching_readmodel/application/queries"
	readmodelinfra "ctf-platform/internal/module/teaching_readmodel/infrastructure"
	readmodelports "ctf-platform/internal/module/teaching_readmodel/ports"
)

type Module struct {
	Handler *teachinghttp.Handler
}

type Deps struct {
	Config          *config.Config
	Logger          *zap.Logger
	DB              *gorm.DB
	Recommendations assessmentcontracts.RecommendationProvider
}

type moduleDeps struct {
	input Deps
	repo  interface {
		readmodelports.TeachingUserLookupRepository
		readmodelports.TeachingClassQueryRepository
		readmodelports.TeachingStudentDirectoryRepository
		readmodelports.TeachingStudentProfileRepository
		readmodelports.TeachingStudentActivityRepository
		readmodelports.TeachingClassInsightRepository
		readmodelports.TeachingOverviewRepository
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
		repo:            readmodelinfra.NewRepository(deps.DB),
		recommendations: deps.Recommendations,
	}
}

func buildQueryService(deps moduleDeps) readmodelqueries.Service {
	cfg := deps.input.Config
	return readmodelqueries.NewQueryService(
		deps.repo,
		cfg.Pagination,
	)
}

func buildOverviewService(deps moduleDeps) readmodelqueries.OverviewService {
	return readmodelqueries.NewOverviewService(deps.repo)
}

func buildClassInsightService(deps moduleDeps) readmodelqueries.ClassInsightService {
	return readmodelqueries.NewClassInsightService(
		deps.repo,
		deps.recommendations,
		deps.input.Logger.Named("teaching_readmodel_class_insight_service"),
	)
}

func buildStudentReviewService(deps moduleDeps) readmodelqueries.StudentReviewService {
	return readmodelqueries.NewStudentReviewService(
		deps.repo,
		deps.recommendations,
	)
}
