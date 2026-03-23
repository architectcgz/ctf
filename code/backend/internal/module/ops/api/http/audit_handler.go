package http

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	opsmodule "ctf-platform/internal/module/ops"
	"ctf-platform/pkg/response"
)

type AuditHandler struct {
	service opsmodule.AuditLogService
}

func NewAuditHandler(service opsmodule.AuditLogService) *AuditHandler {
	return &AuditHandler{service: service}
}

func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	var query dto.AuditLogQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}

	items, total, page, pageSize, err := h.service.ListAuditLogs(c.Request.Context(), &query)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Page(c, items, total, page, pageSize)
}
