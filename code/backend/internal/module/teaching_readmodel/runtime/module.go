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
	Query   readmodelqueries.Service
}

type Deps struct {
	Config          *config.Config
	Logger          *zap.Logger
	DB              *gorm.DB
	Recommendations assessmentcontracts.RecommendationProvider
}

type moduleDeps struct {
	input Deps
	// repo            readmodelports.Repository
	repo interface {
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
		Handler: teachinghttp.NewHandler(service),
		Query:   service,
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
	log := deps.input.Logger
	return readmodelqueries.NewQueryService(
		deps.repo,
		deps.recommendations,
		cfg.Pagination,
		log.Named("teaching_readmodel_query_service"),
	)
}
