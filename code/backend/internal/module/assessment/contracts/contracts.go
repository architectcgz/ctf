package contracts

import (
	"context"

	"ctf-platform/internal/dto"
)

type ProfileService interface {
	UpdateSkillProfileForDimension(ctx context.Context, userID int64, dimension string) error
}

type RecommendationProvider interface {
	Recommend(userID int64, limit int) (*dto.RecommendationResp, error)
	RecommendWithContext(ctx context.Context, userID int64, limit int) (*dto.RecommendationResp, error)
}
