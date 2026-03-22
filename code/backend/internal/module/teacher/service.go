package teacher

import (
	"go.uber.org/zap"

	readmodelapp "ctf-platform/internal/module/teaching_readmodel/application"
)

type RecommendationProvider = readmodelapp.RecommendationProvider
type Service = readmodelapp.QueryService

func NewService(repo *Repository, recommendationService RecommendationProvider, logger *zap.Logger) *Service {
	return readmodelapp.NewQueryService(repo, recommendationService, logger)
}
