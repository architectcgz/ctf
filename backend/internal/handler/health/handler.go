package health

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/service/health"
	"ctf-platform/pkg/response"
)

type Handler struct {
	service health.Service
}

func NewHandler(service health.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(c *gin.Context) {
	response.SuccessWithStatus(c, http.StatusOK, h.service.Check(c.Request.Context()))
}
