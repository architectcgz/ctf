package http

import (
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
	Upsert(challengeID, actorUserID int64, req *dto.UpsertChallengeWriteupReq) (*dto.AdminChallengeWriteupResp, error)
	UpsertSubmission(challengeID, actorUserID int64, req *dto.UpsertSubmissionWriteupReq) (*dto.SubmissionWriteupResp, error)
	ReviewSubmission(submissionID, reviewerID int64, reviewerRole string, req *dto.ReviewSubmissionWriteupReq) (*dto.SubmissionWriteupResp, error)
	Delete(challengeID int64) error
}

type writeupQueryService interface {
	GetAdmin(challengeID int64) (*dto.AdminChallengeWriteupResp, error)
	GetPublished(userID, challengeID int64) (*dto.ChallengeWriteupResp, error)
	GetMySubmission(userID, challengeID int64) (*dto.SubmissionWriteupResp, error)
	ListTeacherSubmissions(requesterID int64, requesterRole string, query *dto.TeacherSubmissionWriteupQuery) (*dto.PageResult, error)
	GetTeacherSubmission(submissionID, requesterID int64, requesterRole string) (*dto.TeacherSubmissionWriteupDetailResp, error)
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
	resp, err := h.commands.Upsert(challengeID, authctx.MustCurrentUser(c).UserID, &req)
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
	resp, err := h.queries.GetAdmin(challengeID)
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
	if err := h.commands.Delete(challengeID); err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *WriteupHandler) GetPublished(c *gin.Context) {
	challengeID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 challenge id")
		return
	}
	resp, err := h.queries.GetPublished(authctx.MustCurrentUser(c).UserID, challengeID)
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
	resp, err := h.commands.UpsertSubmission(challengeID, authctx.MustCurrentUser(c).UserID, &req)
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
	resp, err := h.queries.GetMySubmission(authctx.MustCurrentUser(c).UserID, challengeID)
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
	resp, err := h.queries.ListTeacherSubmissions(currentUser.UserID, currentUser.Role, &query)
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
	resp, err := h.queries.GetTeacherSubmission(submissionID, currentUser.UserID, currentUser.Role)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *WriteupHandler) ReviewSubmission(c *gin.Context) {
	currentUser := authctx.MustCurrentUser(c)
	submissionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的 submission id")
		return
	}
	var req dto.ReviewSubmissionWriteupReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}
	resp, err := h.commands.ReviewSubmission(submissionID, currentUser.UserID, currentUser.Role, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}
