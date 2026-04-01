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
	UnlockHintWithContext(ctx context.Context, userID, challengeID int64, level int) (*dto.UnlockHintResp, error)
	ListTeacherManualReviewSubmissions(requesterID int64, requesterRole string, query *dto.TeacherManualReviewSubmissionQuery) (*dto.PageResult, error)
	GetTeacherManualReviewSubmission(submissionID, requesterID int64, requesterRole string) (*dto.TeacherManualReviewSubmissionDetailResp, error)
	ReviewManualReviewSubmissionWithContext(ctx context.Context, submissionID, reviewerID int64, reviewerRole string, req *dto.ReviewManualReviewSubmissionReq) (*dto.TeacherManualReviewSubmissionDetailResp, error)
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

	resp, err := h.service.UnlockHintWithContext(c.Request.Context(), userID, challengeID, level)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) ListTeacherManualReviewSubmissions(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	var query dto.TeacherManualReviewSubmissionQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.service.ListTeacherManualReviewSubmissions(currentUser.UserID, currentUser.Role, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) GetTeacherManualReviewSubmission(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	submissionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 submission id")
		return
	}
	resp, err := h.service.GetTeacherManualReviewSubmission(submissionID, currentUser.UserID, currentUser.Role)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) ReviewManualReviewSubmission(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	submissionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 submission id")
		return
	}
	var req dto.ReviewManualReviewSubmissionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.service.ReviewManualReviewSubmissionWithContext(c.Request.Context(), submissionID, currentUser.UserID, currentUser.Role, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}
