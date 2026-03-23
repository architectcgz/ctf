package ops

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/pkg/response"
)

type RiskHTTPHandler struct {
	service *RiskService
}

func NewRiskHandler(service *RiskService) *RiskHTTPHandler {
	return &RiskHTTPHandler{service: service}
}

func (h *RiskHTTPHandler) GetCheatDetection(c *gin.Context) {
	result, err := h.service.GetCheatDetection(c.Request.Context())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, result)
}
