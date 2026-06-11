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
	input            Deps
	users            queryports.TeachingUserLookupRepository
	classQuery       queryports.TeachingClassQueryRepository
	studentDirectory queryports.TeachingStudentDirectoryRepository
	studentProfile   queryports.TeachingStudentProfileRepository
	studentActivity  queryports.TeachingStudentActivityRepository
	classInsight     queryports.TeachingClassInsightRepository
	overview         queryports.TeachingOverviewRepository
	recommendations  assessmentcontracts.RecommendationProvider
}

type queryServiceRepository struct {
	queryports.TeachingClassQueryRepository
	queryports.TeachingStudentDirectoryRepository
}

type overviewServiceRepository struct {
	queryports.TeachingClassQueryRepository
	queryports.TeachingStudentDirectoryRepository
	queryports.TeachingClassInsightRepository
	queryports.TeachingOverviewRepository
}

type studentReviewServiceRepository struct {
	queryports.TeachingStudentProfileRepository
	queryports.TeachingStudentActivityRepository
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
		input:            deps,
		users:            deps.Users,
		classQuery:       queryinfra.NewClassQueryRepository(deps.DB),
		studentDirectory: queryinfra.NewStudentDirectoryRepository(deps.DB),
		studentProfile:   queryinfra.NewStudentProfileRepository(deps.DB),
		studentActivity:  queryinfra.NewStudentActivityRepository(deps.DB),
		classInsight:     queryinfra.NewClassInsightRepository(deps.DB),
		overview:         queryinfra.NewOverviewRepository(deps.DB),
		recommendations:  deps.Recommendations,
	}
}

func buildQueryService(deps moduleDeps) teachingqueries.Service {
	cfg := deps.input.Config
	return teachingqueries.NewQueryService(
		deps.users,
		queryServiceRepository{
			TeachingClassQueryRepository:       deps.classQuery,
			TeachingStudentDirectoryRepository: deps.studentDirectory,
		},
		cfg.Pagination,
	)
}

func buildOverviewService(deps moduleDeps) teachingqueries.OverviewService {
	return teachingqueries.NewOverviewService(deps.users, overviewServiceRepository{
		TeachingClassQueryRepository:       deps.classQuery,
		TeachingStudentDirectoryRepository: deps.studentDirectory,
		TeachingClassInsightRepository:     deps.classInsight,
		TeachingOverviewRepository:         deps.overview,
	})
}

func buildClassInsightService(deps moduleDeps) teachingqueries.ClassInsightService {
	return teachingqueries.NewClassInsightService(
		deps.users,
		deps.classInsight,
		deps.recommendations,
		deps.input.Logger.Named("teaching_query_class_insight_service"),
	)
}

func buildStudentReviewService(deps moduleDeps) teachingqueries.StudentReviewService {
	return teachingqueries.NewStudentReviewService(
		deps.users,
		studentReviewServiceRepository{
			TeachingStudentProfileRepository:  deps.studentProfile,
			TeachingStudentActivityRepository: deps.studentActivity,
		},
		deps.recommendations,
	)
}
