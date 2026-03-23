package ops

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/pkg/response"
)

type DashboardHTTPHandler struct {
	service *DashboardService
}

func NewDashboardHandler(service *DashboardService) *DashboardHTTPHandler {
	return &DashboardHTTPHandler{
		service: service,
	}
}

// GetDashboard 获取仪表盘数据
// @Summary 获取仪表盘数据
// @Tags 系统管理
// @Security BearerAuth
// @Success 200 {object} dto.DashboardStats
// @Router /api/v1/admin/dashboard [get]
func (h *DashboardHTTPHandler) GetDashboard(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(c.Request.Context())
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, stats)
}
