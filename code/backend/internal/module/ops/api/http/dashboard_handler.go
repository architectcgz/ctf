package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type dashboardQueryService interface {
	GetDashboardStats(ctx context.Context) (*dto.DashboardStats, error)
}

type DashboardHandler struct {
	service dashboardQueryService
}

func NewDashboardHandler(service dashboardQueryService) *DashboardHandler {
	return &DashboardHandler{
		service: service,
	}
}

// GetDashboard 获取仪表盘数据
// @Summary 获取仪表盘数据
// @Tags 系统管理
// @Security sessionCookieAuth
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
