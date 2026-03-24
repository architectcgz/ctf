package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type awdCommandService interface {
	CreateRound(ctx context.Context, contestID int64, req *dto.CreateAWDRoundReq) (*dto.AWDRoundResp, error)
	RunCurrentRoundChecks(ctx context.Context, contestID int64) (*dto.AWDCheckerRunResp, error)
	RunRoundChecks(ctx context.Context, contestID, roundID int64) (*dto.AWDCheckerRunResp, error)
	UpsertServiceCheck(ctx context.Context, contestID, roundID int64, req *dto.UpsertAWDServiceCheckReq) (*dto.AWDTeamServiceResp, error)
	CreateAttackLog(ctx context.Context, contestID, roundID int64, req *dto.CreateAWDAttackLogReq) (*dto.AWDAttackLogResp, error)
	SubmitAttack(ctx context.Context, userID, contestID, challengeID int64, req *dto.SubmitAWDAttackReq) (*dto.AWDAttackLogResp, error)
}

type awdQueryService interface {
	ListRounds(ctx context.Context, contestID int64) ([]*dto.AWDRoundResp, error)
	ListServices(ctx context.Context, contestID, roundID int64) ([]*dto.AWDTeamServiceResp, error)
	ListAttackLogs(ctx context.Context, contestID, roundID int64) ([]*dto.AWDAttackLogResp, error)
	GetRoundSummary(ctx context.Context, contestID, roundID int64) (*dto.AWDRoundSummaryResp, error)
}

type AWDHandler struct {
	commands awdCommandService
	queries  awdQueryService
}

func NewAWDHandler(commands awdCommandService, queries awdQueryService) *AWDHandler {
	return &AWDHandler{commands: commands, queries: queries}
}

func (h *AWDHandler) CreateRound(c *gin.Context) {
	contestID := c.GetInt64("id")
	var req dto.CreateAWDRoundReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.CreateRound(c.Request.Context(), contestID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) ListRounds(c *gin.Context) {
	contestID := c.GetInt64("id")
	resp, err := h.queries.ListRounds(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) RunCurrentRoundChecks(c *gin.Context) {
	contestID := c.GetInt64("id")
	resp, err := h.commands.RunCurrentRoundChecks(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) RunRoundChecks(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.commands.RunRoundChecks(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) UpsertServiceCheck(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	var req dto.UpsertAWDServiceCheckReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.UpsertServiceCheck(c.Request.Context(), contestID, roundID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) ListServices(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.queries.ListServices(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) CreateAttackLog(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	var req dto.CreateAWDAttackLogReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.CreateAttackLog(c.Request.Context(), contestID, roundID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) SubmitAttack(c *gin.Context) {
	userID := authctx.MustCurrentUser(c).UserID
	contestID := c.GetInt64("id")
	challengeID := c.GetInt64("cid")

	var req dto.SubmitAWDAttackReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.SubmitAttack(c.Request.Context(), userID, contestID, challengeID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) ListAttackLogs(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.queries.ListAttackLogs(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) GetRoundSummary(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.queries.GetRoundSummary(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}
