package http

import (
	"strconv"

	"ctf-platform/internal/authctx"
	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

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

func contestChallengeResultsToDTO(items []*contestqry.ContestChallengeResult) []*dto.ContestChallengeResp {
	result := make([]*dto.ContestChallengeResp, 0, len(items))
	for _, item := range items {
		if item == nil {
			result = append(result, nil)
			continue
		}
		result = append(result, &dto.ContestChallengeResp{
			ID:          item.ID,
			ContestID:   item.ContestID,
			ChallengeID: item.ChallengeID,
			Title:       item.Title,
			Category:    item.Category,
			Difficulty:  item.Difficulty,
			Points:      item.Points,
			Order:       item.Order,
			IsVisible:   item.IsVisible,
			CreatedAt:   item.CreatedAt,
		})
	}
	return result
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

	response.Success(c, contestChallengeResultsToDTO(payload))
}
