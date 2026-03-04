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
	status := h.service.Check(c.Request.Context())
	response.SuccessWithStatus(c, status.HTTPStatus(), status.HealthStatus)
}

func (h *Handler) GetDB(c *gin.Context) {
	if err := h.service.CheckDB(c.Request.Context()); err != nil {
		response.SuccessWithStatus(c, http.StatusServiceUnavailable, gin.H{
			"status": "down",
			"target": "postgres",
		})
		return
	}

	response.Success(c, gin.H{
		"status": "ok",
		"target": "postgres",
	})
}

func (h *Handler) GetRedis(c *gin.Context) {
	if err := h.service.CheckRedis(c.Request.Context()); err != nil {
		response.SuccessWithStatus(c, http.StatusServiceUnavailable, gin.H{
			"status": "down",
			"target": "redis",
		})
		return
	}

	response.Success(c, gin.H{
		"status": "ok",
		"target": "redis",
	})
}
