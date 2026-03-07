package system

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/pkg/response"
)

type RiskHandler struct {
	service *RiskService
}

func NewRiskHandler(service *RiskService) *RiskHandler {
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
