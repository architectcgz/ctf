package http

import (
	"strconv"

	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

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

	var scoreboard *contestqry.ScoreboardResult
	if live {
		scoreboard, err = h.scoreboardQueries.GetLiveScoreboard(c.Request.Context(), contestID, page, pageSize)
	} else {
		scoreboard, err = h.scoreboardQueries.GetScoreboard(c.Request.Context(), contestID, page, pageSize)
	}
	if err != nil {
		response.FromError(c, err)
		return
	}

	response.Success(c, scoreboardResultToDTO(scoreboard))
}

func scoreboardResultToDTO(item *contestqry.ScoreboardResult) *dto.ScoreboardResp {
	if item == nil {
		return nil
	}
	result := &dto.ScoreboardResp{
		Frozen: item.Frozen,
	}
	if item.Contest != nil {
		result.Contest = &dto.ScoreboardContestInfo{
			ID:        item.Contest.ID,
			Title:     item.Contest.Title,
			Status:    item.Contest.Status,
			StartedAt: item.Contest.StartedAt,
			EndsAt:    item.Contest.EndsAt,
		}
	}
	if item.Scoreboard != nil {
		result.Scoreboard = &dto.ScoreboardPage{
			List:     make([]*dto.ScoreboardItem, 0, len(item.Scoreboard.List)),
			Total:    item.Scoreboard.Total,
			Page:     item.Scoreboard.Page,
			PageSize: item.Scoreboard.PageSize,
		}
		for _, scoreboardItem := range item.Scoreboard.List {
			if scoreboardItem == nil {
				result.Scoreboard.List = append(result.Scoreboard.List, nil)
				continue
			}
			result.Scoreboard.List = append(result.Scoreboard.List, &dto.ScoreboardItem{
				Rank:             scoreboardItem.Rank,
				TeamID:           scoreboardItem.TeamID,
				TeamName:         scoreboardItem.TeamName,
				Score:            scoreboardItem.Score,
				SolvedCount:      scoreboardItem.SolvedCount,
				LastSubmissionAt: scoreboardItem.LastSubmissionAt,
			})
		}
	}
	return result
}
