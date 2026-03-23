package http

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/errcode"
	"ctf-platform/pkg/response"
)

type Handler struct {
	service practiceService
}

type practiceService interface {
	StartChallengeWithContext(ctx context.Context, userID, challengeID int64) (*dto.InstanceResp, error)
	StartContestChallenge(ctx context.Context, userID, contestID, challengeID int64) (*dto.InstanceResp, error)
	SubmitFlagWithContext(ctx context.Context, userID, challengeID int64, flag string) (*dto.SubmissionResp, error)
	UnlockHint(userID, challengeID int64, level int) (*dto.UnlockHintResp, error)
}

func NewHandler(service practiceService) *Handler {
	return &Handler{service: service}
}

// StartChallenge 启动靶机实例
// POST /api/v1/challenges/:id/instances
func (h *Handler) StartChallenge(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	instance, err := h.service.StartChallengeWithContext(c.Request.Context(), userID, challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, instance)
}

// StartContestChallenge 启动竞赛靶机实例
// POST /api/v1/contests/:id/challenges/:cid/instances
func (h *Handler) StartContestChallenge(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}
	challengeID, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	instance, err := h.service.StartContestChallenge(c.Request.Context(), userID, contestID, challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, instance)
}

// SubmitFlag 提交 Flag
// POST /api/v1/challenges/:id/submit
func (h *Handler) SubmitFlag(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	var req dto.SubmitFlagReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.SubmitFlagWithContext(c.Request.Context(), userID, challengeID, req.Flag)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

// UnlockHint 解锁题目提示
// POST /api/v1/challenges/:id/hints/:level/unlock
func (h *Handler) UnlockHint(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}
	level, err := strconv.Atoi(c.Param("level"))
	if err != nil || level <= 0 {
		response.Error(c, errcode.ErrInvalidParams)
		return
	}

	resp, err := h.service.UnlockHint(userID, challengeID, level)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}
