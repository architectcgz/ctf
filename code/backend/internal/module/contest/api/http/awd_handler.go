package http

import (
	"context"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type awdService interface {
	CreateRound(ctx context.Context, contestID int64, req *dto.CreateAWDRoundReq) (*dto.AWDRoundResp, error)
	ListRounds(ctx context.Context, contestID int64) ([]*dto.AWDRoundResp, error)
	RunCurrentRoundChecks(ctx context.Context, contestID int64) (*dto.AWDCheckerRunResp, error)
	RunRoundChecks(ctx context.Context, contestID, roundID int64) (*dto.AWDCheckerRunResp, error)
	UpsertServiceCheck(ctx context.Context, contestID, roundID int64, req *dto.UpsertAWDServiceCheckReq) (*dto.AWDTeamServiceResp, error)
	ListServices(ctx context.Context, contestID, roundID int64) ([]*dto.AWDTeamServiceResp, error)
	CreateAttackLog(ctx context.Context, contestID, roundID int64, req *dto.CreateAWDAttackLogReq) (*dto.AWDAttackLogResp, error)
	SubmitAttack(ctx context.Context, userID, contestID, challengeID int64, req *dto.SubmitAWDAttackReq) (*dto.AWDAttackLogResp, error)
	ListAttackLogs(ctx context.Context, contestID, roundID int64) ([]*dto.AWDAttackLogResp, error)
	GetRoundSummary(ctx context.Context, contestID, roundID int64) (*dto.AWDRoundSummaryResp, error)
}

type AWDHandler struct {
	service awdService
}

func NewAWDHandler(service awdService) *AWDHandler {
	return &AWDHandler{service: service}
}

func (h *AWDHandler) CreateRound(c *gin.Context) {
	contestID := c.GetInt64("id")
	var req dto.CreateAWDRoundReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.CreateRound(c.Request.Context(), contestID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) ListRounds(c *gin.Context) {
	contestID := c.GetInt64("id")
	resp, err := h.service.ListRounds(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) RunCurrentRoundChecks(c *gin.Context) {
	contestID := c.GetInt64("id")
	resp, err := h.service.RunCurrentRoundChecks(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) RunRoundChecks(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.service.RunRoundChecks(c.Request.Context(), contestID, roundID)
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

	resp, err := h.service.UpsertServiceCheck(c.Request.Context(), contestID, roundID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) ListServices(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.service.ListServices(c.Request.Context(), contestID, roundID)
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

	resp, err := h.service.CreateAttackLog(c.Request.Context(), contestID, roundID, &req)
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

	resp, err := h.service.SubmitAttack(c.Request.Context(), userID, contestID, challengeID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) ListAttackLogs(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.service.ListAttackLogs(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *AWDHandler) GetRoundSummary(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.service.GetRoundSummary(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}
