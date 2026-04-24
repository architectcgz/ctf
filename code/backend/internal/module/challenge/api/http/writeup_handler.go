package http

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type WriteupHandler struct {
	commands writeupCommandService
	queries  writeupQueryService
}

type writeupCommandService interface {
	Upsert(ctx context.Context, challengeID, actorUserID int64, req *dto.UpsertChallengeWriteupReq) (*dto.AdminChallengeWriteupResp, error)
	UpsertSubmission(ctx context.Context, challengeID, actorUserID int64, req *dto.UpsertSubmissionWriteupReq) (*dto.SubmissionWriteupResp, error)
	RecommendOfficial(ctx context.Context, challengeID, actorUserID int64) (*dto.AdminChallengeWriteupResp, error)
	UnrecommendOfficial(ctx context.Context, challengeID, actorUserID int64) (*dto.AdminChallengeWriteupResp, error)
	RecommendCommunity(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*dto.SubmissionWriteupResp, error)
	UnrecommendCommunity(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*dto.SubmissionWriteupResp, error)
	HideCommunity(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*dto.SubmissionWriteupResp, error)
	RestoreCommunity(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*dto.SubmissionWriteupResp, error)
	Delete(ctx context.Context, challengeID int64) error
}

type writeupQueryService interface {
	GetAdmin(ctx context.Context, challengeID int64) (*dto.AdminChallengeWriteupResp, error)
	GetPublished(ctx context.Context, userID, challengeID int64) (*dto.ChallengeWriteupResp, error)
	GetMySubmission(ctx context.Context, userID, challengeID int64) (*dto.SubmissionWriteupResp, error)
	ListRecommendedSolutions(ctx context.Context, userID, challengeID int64) (*dto.PageResult, error)
	ListCommunitySolutions(ctx context.Context, userID, challengeID int64, query *dto.CommunityChallengeSolutionQuery) (*dto.PageResult, error)
	ListTeacherSubmissions(ctx context.Context, requesterID int64, requesterRole string, query *dto.TeacherSubmissionWriteupQuery) (*dto.PageResult, error)
	GetTeacherSubmission(ctx context.Context, submissionID, requesterID int64, requesterRole string) (*dto.TeacherSubmissionWriteupDetailResp, error)
}

func NewWriteupHandler(commands writeupCommandService, queries writeupQueryService) *WriteupHandler {
	return &WriteupHandler{commands: commands, queries: queries}
}

func (h *WriteupHandler) Upsert(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	var req dto.UpsertChallengeWriteupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.commands.Upsert(c.Request.Context(), challengeID, authctx.MustCurrentUser(c).UserID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *WriteupHandler) GetAdmin(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	resp, err := h.queries.GetAdmin(c.Request.Context(), challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *WriteupHandler) Delete(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	if err := h.commands.Delete(c.Request.Context(), challengeID); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *WriteupHandler) RecommendOfficial(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	resp, err := h.commands.RecommendOfficial(c.Request.Context(), challengeID, authctx.MustCurrentUser(c).UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *WriteupHandler) UnrecommendOfficial(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	resp, err := h.commands.UnrecommendOfficial(c.Request.Context(), challengeID, authctx.MustCurrentUser(c).UserID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *WriteupHandler) GetPublished(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	resp, err := h.queries.GetPublished(c.Request.Context(), authctx.MustCurrentUser(c).UserID, challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *WriteupHandler) UpsertSubmission(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	var req dto.UpsertSubmissionWriteupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.commands.UpsertSubmission(c.Request.Context(), challengeID, authctx.MustCurrentUser(c).UserID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *WriteupHandler) GetMySubmission(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	resp, err := h.queries.GetMySubmission(c.Request.Context(), authctx.MustCurrentUser(c).UserID, challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *WriteupHandler) ListRecommendedSolutions(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	resp, err := h.queries.ListRecommendedSolutions(c.Request.Context(), authctx.MustCurrentUser(c).UserID, challengeID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *WriteupHandler) ListCommunitySolutions(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	var query dto.CommunityChallengeSolutionQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.queries.ListCommunitySolutions(c.Request.Context(), authctx.MustCurrentUser(c).UserID, challengeID, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *WriteupHandler) ListTeacherSubmissions(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	var query dto.TeacherSubmissionWriteupQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.queries.ListTeacherSubmissions(c.Request.Context(), currentUser.UserID, currentUser.Role, &query)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *WriteupHandler) GetTeacherSubmission(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	submissionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 submission id")
		return
	}
	resp, err := h.queries.GetTeacherSubmission(c.Request.Context(), submissionID, currentUser.UserID, currentUser.Role)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *WriteupHandler) RecommendCommunity(c *gin.Context) {
	h.respondCommunityModeration(c, func(submissionID int64, currentUser authctx.CurrentUser) (*dto.SubmissionWriteupResp, error) {
		return h.commands.RecommendCommunity(c.Request.Context(), submissionID, currentUser.UserID, currentUser.Role)
	})
}

func (h *WriteupHandler) UnrecommendCommunity(c *gin.Context) {
	h.respondCommunityModeration(c, func(submissionID int64, currentUser authctx.CurrentUser) (*dto.SubmissionWriteupResp, error) {
		return h.commands.UnrecommendCommunity(c.Request.Context(), submissionID, currentUser.UserID, currentUser.Role)
	})
}

func (h *WriteupHandler) HideCommunity(c *gin.Context) {
	h.respondCommunityModeration(c, func(submissionID int64, currentUser authctx.CurrentUser) (*dto.SubmissionWriteupResp, error) {
		return h.commands.HideCommunity(c.Request.Context(), submissionID, currentUser.UserID, currentUser.Role)
	})
}

func (h *WriteupHandler) RestoreCommunity(c *gin.Context) {
	h.respondCommunityModeration(c, func(submissionID int64, currentUser authctx.CurrentUser) (*dto.SubmissionWriteupResp, error) {
		return h.commands.RestoreCommunity(c.Request.Context(), submissionID, currentUser.UserID, currentUser.Role)
	})
}

func (h *WriteupHandler) respondCommunityModeration(
	c *gin.Context,
	action func(submissionID int64, currentUser authctx.CurrentUser) (*dto.SubmissionWriteupResp, error),
) {
	currentUser := authctx.MustCurrentUser(c)
	submissionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 submission id")
		return
	}
	resp, err := action(submissionID, currentUser)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}
