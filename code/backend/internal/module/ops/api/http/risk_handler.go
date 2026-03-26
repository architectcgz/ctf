package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type riskQueryService interface {
	GetCheatDetection(ctx context.Context) (*dto.CheatDetectionResp, error)
}

type RiskHandler struct {
	service riskQueryService
}

func NewRiskHandler(service riskQueryService) *RiskHandler {
	return &RiskHandler{service: service}
}

func (h *RiskHandler) GetCheatDetection(c *gin.Context) {
	result, err := h.service.GetCheatDetection(c.Request.Context())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, result)
}
