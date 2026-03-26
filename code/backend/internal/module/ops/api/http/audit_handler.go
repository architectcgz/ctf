package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type auditQueryService interface {
	ListAuditLogs(ctx context.Context, query *dto.AuditLogQuery) ([]dto.AuditLogItem, int64, int, int, error)
}

type AuditHandler struct {
	service auditQueryService
}

func NewAuditHandler(service auditQueryService) *AuditHandler {
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
