package contest

import (
	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
	if err != nil {
		response.InvalidParams(c, "无效的竞赛ID")
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
	if err != nil {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}

	var req dto.FreezeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	contest, err := h.repo.FindByID(contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	freezeTime := contest.EndTime.Add(-time.Duration(req.MinutesBeforeEnd) * time.Minute)
	contest.FreezeTime = &freezeTime

	if err := h.repo.Update(contest); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{"freeze_time": freezeTime})
}

// UnfreezeScoreboard 解冻排行榜
func (h *Handler) UnfreezeScoreboard(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}

	contest, err := h.repo.FindByID(contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}

	contest.FreezeTime = nil
	if err := h.repo.Update(contest); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "排行榜已解冻"})
}
