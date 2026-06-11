package http

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/apperror"
	"ctf-platform/internal/auditlog"
	"ctf-platform/internal/authctx"
	response "ctf-platform/internal/httpresponse"
	"ctf-platform/internal/middleware"
)

// SubmitFlag 提交 Flag
// POST /api/v1/challenges/:id/submit
func (h *Handler) SubmitFlag(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	var req SubmitFlagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	auditControl := &auditlog.Control{}
	ctx := auditlog.WithControl(c.Request.Context(), auditControl)

	resp, err := h.submissions.SubmitFlag(ctx, userID, challengeID, req.Flag)
	if err != nil {
		response.FromError(c, err)
		return
	}
	if auditControl.Skip {
		middleware.SetSkipAudit(c)
	}

	response.Success(c, resp)
}

func (h *Handler) ListMyChallengeSubmissions(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, apperror.ErrInvalidParams)
		return
	}

	resp, err := h.submissions.ListMyChallengeSubmissions(c.Request.Context(), userID, challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}
