package assessment

import (
	"ctf-platform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	recommendationService *RecommendationService
}

func NewHandler(recommendationService *RecommendationService) *Handler {
	return &Handler{
		recommendationService: recommendationService,
	}
}

func (h *Handler) GetRecommendations(c *gin.Context) {
	userID := c.GetInt64("user_id")

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	weakDimensions, err := h.recommendationService.GetWeakDimensions(userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	challenges, err := h.recommendationService.RecommendChallenges(userID, limit)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{
		"weak_dimensions": weakDimensions,
		"challenges":      challenges,
	})
}
