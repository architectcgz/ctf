package http

import (
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *AWDHandler) ListContestAWDServices(c *gin.Context) {
	contestID := c.GetInt64("id")
	resp, err := h.serviceQueries.ListContestAWDServices(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) CreateContestAWDService(c *gin.Context) {
	contestID := c.GetInt64("id")
	var req dto.CreateContestAWDServiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.serviceCommands.CreateContestAWDService(c.Request.Context(), contestID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) UpdateContestAWDService(c *gin.Context) {
	contestID := c.GetInt64("id")
	serviceID := c.GetInt64("sid")
	var req dto.UpdateContestAWDServiceReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.serviceCommands.UpdateContestAWDService(c.Request.Context(), contestID, serviceID, &req); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *AWDHandler) DeleteContestAWDService(c *gin.Context) {
	contestID := c.GetInt64("id")
	serviceID := c.GetInt64("sid")
	if err := h.serviceCommands.DeleteContestAWDService(c.Request.Context(), contestID, serviceID); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}
