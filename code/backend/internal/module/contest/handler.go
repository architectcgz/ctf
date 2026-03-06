package contest

import (
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	ErrMsgInvalidContestID = "无效的竞赛ID"
)

type Handler struct {
	scoreboardService *ScoreboardService
	repo              *Repository
}

func NewHandler(scoreboardService *ScoreboardService, repo *Repository) *Handler {
	return &Handler{
		scoreboardService: scoreboardService,
		repo:              repo,
	}
}

// GetScoreboard 获取排行榜
func (h *Handler) GetScoreboard(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, ErrMsgInvalidContestID)
		return
	}

	scoreboard, err := h.scoreboardService.GetScoreboard(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, scoreboard)
}

// FreezeScoreboard 冻结排行榜
func (h *Handler) FreezeScoreboard(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, ErrMsgInvalidContestID)
		return
	}

	var req dto.FreezeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	if err := h.scoreboardService.FreezeScoreboard(c.Request.Context(), contestID, req.MinutesBeforeEnd); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "排行榜已冻结"})
}

// UnfreezeScoreboard 解冻排行榜
func (h *Handler) UnfreezeScoreboard(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, ErrMsgInvalidContestID)
		return
	}

	if err := h.scoreboardService.UnfreezeScoreboard(c.Request.Context(), contestID); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "排行榜已解冻"})
}
