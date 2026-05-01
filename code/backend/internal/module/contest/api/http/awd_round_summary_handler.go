package http

import (
	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"

	"github.com/gin-gonic/gin"
)

func (h *AWDHandler) GetRoundSummary(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.queries.GetRoundSummary(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, awdRoundSummaryResultToDTO(resp))
}

func awdRoundSummaryResultToDTO(item *contestqry.AWDRoundSummaryResult) *dto.AWDRoundSummaryResp {
	if item == nil {
		return nil
	}
	result := &dto.AWDRoundSummaryResp{
		Items: make([]*dto.AWDRoundSummaryItem, 0, len(item.Items)),
	}
	if item.Round != nil {
		result.Round = &dto.AWDRoundResp{
			ID:           item.Round.ID,
			ContestID:    item.Round.ContestID,
			RoundNumber:  item.Round.RoundNumber,
			Status:       item.Round.Status,
			StartedAt:    item.Round.StartedAt,
			EndedAt:      item.Round.EndedAt,
			AttackScore:  item.Round.AttackScore,
			DefenseScore: item.Round.DefenseScore,
			CreatedAt:    item.Round.CreatedAt,
			UpdatedAt:    item.Round.UpdatedAt,
		}
	}
	if item.Metrics != nil {
		result.Metrics = &dto.AWDRoundMetrics{
			TotalServiceCount:         item.Metrics.TotalServiceCount,
			ServiceUpCount:            item.Metrics.ServiceUpCount,
			ServiceDownCount:          item.Metrics.ServiceDownCount,
			ServiceCompromisedCount:   item.Metrics.ServiceCompromisedCount,
			AttackedServiceCount:      item.Metrics.AttackedServiceCount,
			DefenseSuccessCount:       item.Metrics.DefenseSuccessCount,
			TotalAttackCount:          item.Metrics.TotalAttackCount,
			SuccessfulAttackCount:     item.Metrics.SuccessfulAttackCount,
			FailedAttackCount:         item.Metrics.FailedAttackCount,
			SchedulerCheckCount:       item.Metrics.SchedulerCheckCount,
			ManualCurrentRoundChecks:  item.Metrics.ManualCurrentRoundChecks,
			ManualSelectedRoundChecks: item.Metrics.ManualSelectedRoundChecks,
			ManualServiceCheckCount:   item.Metrics.ManualServiceCheckCount,
			SubmissionAttackCount:     item.Metrics.SubmissionAttackCount,
			ManualAttackLogCount:      item.Metrics.ManualAttackLogCount,
			LegacyAttackLogCount:      item.Metrics.LegacyAttackLogCount,
		}
	}
	for _, summaryItem := range item.Items {
		if summaryItem == nil {
			result.Items = append(result.Items, nil)
			continue
		}
		result.Items = append(result.Items, &dto.AWDRoundSummaryItem{
			TeamID:                  summaryItem.TeamID,
			TeamName:                summaryItem.TeamName,
			ServiceUpCount:          summaryItem.ServiceUpCount,
			ServiceDownCount:        summaryItem.ServiceDownCount,
			ServiceCompromisedCount: summaryItem.ServiceCompromisedCount,
			AttackScore:             summaryItem.AttackScore,
			DefenseScore:            summaryItem.DefenseScore,
			SLAScore:                summaryItem.SLAScore,
			TotalScore:              summaryItem.TotalScore,
			SuccessfulAttackCount:   summaryItem.SuccessfulAttackCount,
			SuccessfulBreachCount:   summaryItem.SuccessfulBreachCount,
			UniqueAttackersAgainst:  summaryItem.UniqueAttackersAgainst,
		})
	}
	return result
}
