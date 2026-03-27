package http

import (
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetContest(c *gin.Context) {
	id := c.GetInt64("id")
	resp, err := h.queries.GetContest(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) ListContests(c *gin.Context) {
	var req dto.ListContestsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	contests, total, err := h.queries.ListContests(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 20
	}

	response.Page(c, contests, total, page, size)
}
