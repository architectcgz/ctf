package http

import (
	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"

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
	response.Success(c, awdWorkspaceResultToDTO(resp))
}

func awdWorkspaceResultToDTO(item *contestqry.AWDWorkspaceResult) *dto.ContestAWDWorkspaceResp {
	if item == nil {
		return nil
	}
	mapped := contestRequestMapper.ToAWDWorkspaceResp(*item)
	return &mapped
}
