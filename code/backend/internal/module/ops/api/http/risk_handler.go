package http

import (
	"github.com/gin-gonic/gin"

	opsmodule "ctf-platform/internal/module/ops"
	"ctf-platform/pkg/response"
)

type RiskHandler struct {
	service opsmodule.RiskService
}

func NewRiskHandler(service opsmodule.RiskService) *RiskHandler {
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
