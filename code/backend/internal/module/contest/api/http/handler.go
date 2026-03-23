package http

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	"ctf-platform/pkg/response"
)

type contestService interface {
	CreateContest(ctx context.Context, req *dto.CreateContestReq) (*dto.ContestResp, error)
	UpdateContest(ctx context.Context, id int64, req *dto.UpdateContestReq) (*dto.ContestResp, error)
	GetContest(ctx context.Context, id int64) (*dto.ContestResp, error)
	ListContests(ctx context.Context, req *dto.ListContestsReq) ([]*dto.ContestResp, int64, error)
}

type scoreboardService interface {
	GetScoreboard(ctx context.Context, contestID int64, page, pageSize int) (*dto.ScoreboardResp, error)
	GetLiveScoreboard(ctx context.Context, contestID int64, page, pageSize int) (*dto.ScoreboardResp, error)
	FreezeScoreboard(ctx context.Context, contestID int64, minutesBeforeEnd int) error
	UnfreezeScoreboard(ctx context.Context, contestID int64) error
}

type Handler struct {
	service           contestService
	scoreboardService scoreboardService
}

func NewHandler(service contestService, scoreboardServices ...scoreboardService) *Handler {
	var sb scoreboardService
	if len(scoreboardServices) > 0 {
		sb = scoreboardServices[0]
	}
	return &Handler{
		service:           service,
		scoreboardService: sb,
	}
}

func (h *Handler) CreateContest(c *gin.Context) {
	var req dto.CreateContestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.CreateContest(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) UpdateContest(c *gin.Context) {
	id := c.GetInt64("id")
	var req dto.UpdateContestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.service.UpdateContest(c.Request.Context(), id, &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, resp)
}

func (h *Handler) GetContest(c *gin.Context) {
	id := c.GetInt64("id")
	resp, err := h.service.GetContest(c.Request.Context(), id)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, resp)
}

func (h *Handler) ListContests(c *gin.Context) {
	var req dto.ListContestsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	contests, total, err := h.service.ListContests(c.Request.Context(), &req)
	if err != nil {
		response.FromError(c, err)
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	size := req.Size
	if size < 1 {
		size = 20
	}

	response.Page(c, contests, total, page, size)
}

func (h *Handler) GetScoreboard(c *gin.Context) {
	h.getScoreboard(c, false)
}

func (h *Handler) GetLiveScoreboard(c *gin.Context) {
	h.getScoreboard(c, true)
}

func (h *Handler) getScoreboard(c *gin.Context, live bool) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}

	page := 1
	if value := c.Query("page"); value != "" {
		if parsed, parseErr := strconv.Atoi(value); parseErr == nil && parsed > 0 {
			page = parsed
		}
	}
	pageSize := 20
	if value := c.Query("page_size"); value != "" {
		if parsed, parseErr := strconv.Atoi(value); parseErr == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	var scoreboard *dto.ScoreboardResp
	if live {
		scoreboard, err = h.scoreboardService.GetLiveScoreboard(c.Request.Context(), contestID, page, pageSize)
	} else {
		scoreboard, err = h.scoreboardService.GetScoreboard(c.Request.Context(), contestID, page, pageSize)
	}
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, scoreboard)
}

func (h *Handler) FreezeScoreboard(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
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

func (h *Handler) UnfreezeScoreboard(c *gin.Context) {
	contestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || contestID <= 0 {
		response.InvalidParams(c, "无效的竞赛ID")
		return
	}

	if err := h.scoreboardService.UnfreezeScoreboard(c.Request.Context(), contestID); err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, gin.H{"message": "排行榜已解冻"})
}
