package http

import (
	"context"
	"strconv"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

type challengeCommandService interface {
	AddChallengeToContest(ctx context.Context, contestID int64, req *dto.AddContestChallengeReq) (*dto.ContestChallengeResp, error)
	RemoveChallengeFromContest(ctx context.Context, contestID, challengeID int64) error
	UpdateChallenge(ctx context.Context, contestID, challengeID int64, req *dto.UpdateContestChallengeReq) error
}

type challengeQueryService interface {
	GetContestChallenges(ctx context.Context, userID, contestID int64) ([]*dto.ContestChallengeInfo, error)
	ListAdminChallenges(ctx context.Context, contestID int64) ([]*dto.ContestChallengeResp, error)
}

type ChallengeHandler struct {
	commands challengeCommandService
	queries  challengeQueryService
}

func NewChallengeHandler(commands challengeCommandService, queries challengeQueryService) *ChallengeHandler {
	return &ChallengeHandler{commands: commands, queries: queries}
}

func (h *ChallengeHandler) AddChallenge(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛 ID")
		return
	}

	var req dto.AddContestChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.commands.AddChallengeToContest(c.Request.Context(), contestID, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *ChallengeHandler) RemoveChallenge(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛 ID")
		return
	}

	challengeID, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的题目 ID")
		return
	}

	if err := h.commands.RemoveChallengeFromContest(c.Request.Context(), contestID, challengeID); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *ChallengeHandler) UpdatePoints(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛 ID")
		return
	}

	challengeID, err := strconv.ParseInt(c.Param("cid"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的题目 ID")
		return
	}

	var req dto.UpdateContestChallengeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.commands.UpdateChallenge(c.Request.Context(), contestID, challengeID, &req); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, nil)
}

func (h *ChallengeHandler) ListChallenges(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛 ID")
		return
	}

	payload, err := h.queries.GetContestChallenges(c.Request.Context(), authctx.MustCurrentUser(c).UserID, contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, payload)
}

func (h *ChallengeHandler) ListAdminChallenges(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛 ID")
		return
	}

	payload, err := h.queries.ListAdminChallenges(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, payload)
}
