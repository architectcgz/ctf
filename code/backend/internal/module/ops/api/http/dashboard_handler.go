package http

import (
	"github.com/gin-gonic/gin"

	opsmodule "ctf-platform/internal/module/ops"
	"ctf-platform/pkg/response"
)

type DashboardHandler struct {
	service opsmodule.DashboardService
}

func NewDashboardHandler(service opsmodule.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		service: service,
	}
}

// GetDashboard 获取仪表盘数据
// @Summary 获取仪表盘数据
// @Tags 系统管理
// @Security BearerAuth
// @Success 200 {object} dto.DashboardStats
// @Router /api/v1/admin/dashboard [get]
func (h *DashboardHandler) GetDashboard(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, stats)
}
