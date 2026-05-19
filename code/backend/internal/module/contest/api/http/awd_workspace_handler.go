package http

import (
	"ctf-platform/internal/authctx"
	response "ctf-platform/internal/httpresponse"

	"github.com/gin-gonic/gin"
)

func (h *AWDHandler) GetUserWorkspace(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	contestID := c.GetInt64("id")
	resp, err := h.queries.GetUserWorkspace(c.Request.Context(), currentUser.UserID, contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, contestResponseMapper.ToAWDWorkspaceRespPtr(resp))
}
