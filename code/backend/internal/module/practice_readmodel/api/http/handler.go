package http

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	practicereadmodelqueries "ctf-platform/internal/module/practice_readmodel/application/queries"
	"ctf-platform/pkg/response"
)

type Handler struct {
	query practicereadmodelqueries.Service
}

func NewHandler(query practicereadmodelqueries.Service) *Handler {
	return &Handler{query: query}
}

// GetProgress 获取个人解题进度
// GET /api/v1/users/me/progress
func (h *Handler) GetProgress(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID

	resp, err := h.query.GetProgress(c.Request.Context(), userID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

// GetTimeline 获取解题时间线
// GET /api/v1/users/me/timeline
func (h *Handler) GetTimeline(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID

	var req struct {
		Limit  int `form:"limit" binding:"omitempty,min=1,max=100"`
		Offset int `form:"offset" binding:"omitempty,min=0"`
	}
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if req.Limit == 0 {
		req.Limit = 100
	}

	resp, err := h.query.GetTimeline(c.Request.Context(), userID, req.Limit, req.Offset)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}
