package http

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *AWDHandler) GetReadiness(c *gin.Context) {
	contestID := c.GetInt64("id")
	resp, err := h.queries.GetReadiness(c.Request.Context(), contestID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, awdReadinessResultToDTO(resp))
}

func awdReadinessResultToDTO(result *contestqry.AWDReadinessResult) *dto.AWDReadinessResp {
	if result == nil {
		return nil
	}
	items := make([]*dto.AWDReadinessItemResp, 0, len(result.Items))
	for i := range result.Items {
		item := result.Items[i]
		items = append(items, &dto.AWDReadinessItemResp{
			ServiceID:       item.ServiceID,
			AWDChallengeID:  item.AWDChallengeID,
			Title:           item.Title,
			CheckerType:     model.AWDCheckerType(item.CheckerType),
			ValidationState: item.ValidationState,
			LastPreviewAt:   item.LastPreviewAt,
			LastAccessURL:   item.LastAccessURL,
			BlockingReason:  item.BlockingReason,
		})
	}
	return &dto.AWDReadinessResp{
		ContestID:                result.ContestID,
		Ready:                    result.Ready,
		TotalChallenges:          result.TotalChallenges,
		PassedChallenges:         result.PassedChallenges,
		PendingChallenges:        result.PendingChallenges,
		FailedChallenges:         result.FailedChallenges,
		StaleChallenges:          result.StaleChallenges,
		MissingCheckerChallenges: result.MissingCheckerChallenges,
		BlockingCount:            result.BlockingCount,
		BlockingActions:          append([]string(nil), result.BlockingActions...),
		GlobalBlockingReasons:    append([]string(nil), result.GlobalBlockingReasons...),
		Items:                    items,
	}
}
