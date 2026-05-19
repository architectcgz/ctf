package http

import (
	"context"

	"github.com/gin-gonic/gin"

	response "ctf-platform/internal/httpresponse"
	opsqry "ctf-platform/internal/module/ops/application/queries"
)

type auditQueryService interface {
	ListAuditLogs(ctx context.Context, query *opsqry.AuditLogQuery) ([]opsqry.AuditLogItem, int64, int, int, error)
}

type AuditHandler struct {
	service auditQueryService
}

func NewAuditHandler(service auditQueryService) *AuditHandler {
	return &AuditHandler{service: service}
}

func (h *AuditHandler) ListAuditLogs(c *gin.Context) {
	var query opsqry.AuditLogQuery
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
