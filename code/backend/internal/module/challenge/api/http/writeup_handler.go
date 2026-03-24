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
	Delete(challengeID int64) error
}

type writeupQueryService interface {
	GetAdmin(challengeID int64) (*dto.AdminChallengeWriteupResp, error)
	GetPublished(userID, challengeID int64) (*dto.ChallengeWriteupResp, error)
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
