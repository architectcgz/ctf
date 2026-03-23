package http

import (
	"context"
	"strconv"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

type submissionService interface {
	SubmitFlagInContest(ctx context.Context, userID, contestID, challengeID int64, flag string) (*dto.SubmissionResp, error)
}

type SubmissionHandler struct {
	service submissionService
}

func NewSubmissionHandler(service submissionService) *SubmissionHandler {
	return &SubmissionHandler{service: service}
}

// SubmitFlag 竞赛中提交 Flag
func (h *SubmissionHandler) SubmitFlag(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}

	challengeID, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil || challengeID <= 0 {
		response.InvalidParams(c, "无效的题目ID")
		return
	}

	userID := authctx.MustCurrentUser(c).UserID

	var req dto.SubmitFlagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.SubmitFlagInContest(c.Request.Context(), userID, contestID, challengeID, req.Flag)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}
