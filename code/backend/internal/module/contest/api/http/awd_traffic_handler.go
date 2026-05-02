package http

import (
	"github.com/gin-gonic/gin"

	"ctf-platform/internal/dto"
	contestqry "ctf-platform/internal/module/contest/application/queries"
	"ctf-platform/pkg/response"
)

func (h *AWDHandler) GetTrafficSummary(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	resp, err := h.queries.GetTrafficSummary(c.Request.Context(), contestID, roundID)
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, awdTrafficSummaryResultToDTO(resp))
}

func (h *AWDHandler) ListTrafficEvents(c *gin.Context) {
	contestID := c.GetInt64("id")
	roundID := c.GetInt64("rid")
	var req dto.ListAWDTrafficEventsReq
	if err := c.ShouldBindQuery(&req); err != nil {
		response.ValidationError(c, err)
		return
	}

	resp, err := h.queries.ListTrafficEvents(c.Request.Context(), contestID, roundID, contestRequestMapper.ToListAWDTrafficEventsInput(req))
	if err != nil {
		response.FromError(c, err)
		return
	}
	response.Success(c, awdTrafficEventPageResultToDTO(resp))
}

func awdTrafficEventPageResultToDTO(result *contestqry.AWDTrafficEventPageResult) *dto.AWDTrafficEventPageResp {
	if result == nil {
		return nil
	}
	items := make([]*dto.AWDTrafficEventResp, 0, len(result.List))
	for i := range result.List {
		item := result.List[i]
		items = append(items, &dto.AWDTrafficEventResp{
			ID:                item.ID,
			ContestID:         item.ContestID,
			RoundID:           item.RoundID,
			AttackerTeamID:    item.AttackerTeamID,
			AttackerTeam:      item.AttackerTeam,
			AttackerTeamName:  item.AttackerTeamName,
			VictimTeamID:      item.VictimTeamID,
			VictimTeam:        item.VictimTeam,
			VictimTeamName:    item.VictimTeamName,
			ServiceID:         item.ServiceID,
			AWDChallengeID:    item.AWDChallengeID,
			AWDChallengeTitle: item.AWDChallengeTitle,
			Method:            item.Method,
			Path:              item.Path,
			StatusCode:        item.StatusCode,
			StatusGroup:       item.StatusGroup,
			IsError:           item.IsError,
			Source:            item.Source,
			OccurredAt:        item.OccurredAt,
		})
	}
	return &dto.AWDTrafficEventPageResp{
		List:     items,
		Total:    result.Total,
		Page:     result.Page,
		PageSize: result.PageSize,
	}
}

func awdTrafficSummaryResultToDTO(item *contestqry.AWDTrafficSummaryResult) *dto.AWDTrafficSummaryResp {
	if item == nil {
		return nil
	}
	result := &dto.AWDTrafficSummaryResp{
		ContestID:           item.ContestID,
		RoundID:             item.RoundID,
		TotalRequests:       item.TotalRequests,
		ActiveAttackerTeams: item.ActiveAttackerTeams,
		TargetedTeams:       item.TargetedTeams,
		ErrorRequests:       item.ErrorRequests,
		UniquePathCount:     item.UniquePathCount,
		LatestEventAt:       item.LatestEventAt,
		Trend:               make([]*dto.AWDTrafficTrendBucketResp, 0, len(item.Trend)),
		TopAttackers:        make([]*dto.AWDTrafficTopTeamResp, 0, len(item.TopAttackers)),
		TopVictims:          make([]*dto.AWDTrafficTopTeamResp, 0, len(item.TopVictims)),
		TopChallenges:       make([]*dto.AWDTrafficTopChallengeResp, 0, len(item.TopChallenges)),
		TopPaths:            make([]*dto.AWDTrafficTopPathResp, 0, len(item.TopPaths)),
		TopErrorPaths:       make([]*dto.AWDTrafficTopPathResp, 0, len(item.TopErrorPaths)),
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
	for _, trend := range item.Trend {
		if trend == nil {
			result.Trend = append(result.Trend, nil)
			continue
		}
		result.Trend = append(result.Trend, &dto.AWDTrafficTrendBucketResp{
			BucketStart:  trend.BucketStart,
			RequestCount: trend.RequestCount,
			ErrorCount:   trend.ErrorCount,
		})
	}
	for _, attacker := range item.TopAttackers {
		if attacker == nil {
			result.TopAttackers = append(result.TopAttackers, nil)
			continue
		}
		result.TopAttackers = append(result.TopAttackers, &dto.AWDTrafficTopTeamResp{
			TeamID:       attacker.TeamID,
			TeamName:     attacker.TeamName,
			RequestCount: attacker.RequestCount,
			ErrorCount:   attacker.ErrorCount,
		})
	}
	for _, victim := range item.TopVictims {
		if victim == nil {
			result.TopVictims = append(result.TopVictims, nil)
			continue
		}
		result.TopVictims = append(result.TopVictims, &dto.AWDTrafficTopTeamResp{
			TeamID:       victim.TeamID,
			TeamName:     victim.TeamName,
			RequestCount: victim.RequestCount,
			ErrorCount:   victim.ErrorCount,
		})
	}
	for _, challenge := range item.TopChallenges {
		if challenge == nil {
			result.TopChallenges = append(result.TopChallenges, nil)
			continue
		}
		result.TopChallenges = append(result.TopChallenges, &dto.AWDTrafficTopChallengeResp{
			AWDChallengeID:    challenge.AWDChallengeID,
			AWDChallengeTitle: challenge.AWDChallengeTitle,
			RequestCount:      challenge.RequestCount,
			ErrorCount:        challenge.ErrorCount,
		})
	}
	for _, path := range item.TopPaths {
		if path == nil {
			result.TopPaths = append(result.TopPaths, nil)
			continue
		}
		result.TopPaths = append(result.TopPaths, &dto.AWDTrafficTopPathResp{
			Path:           path.Path,
			RequestCount:   path.RequestCount,
			ErrorCount:     path.ErrorCount,
			LastStatusCode: path.LastStatusCode,
		})
	}
	for _, path := range item.TopErrorPaths {
		if path == nil {
			result.TopErrorPaths = append(result.TopErrorPaths, nil)
			continue
		}
		result.TopErrorPaths = append(result.TopErrorPaths, &dto.AWDTrafficTopPathResp{
			Path:           path.Path,
			RequestCount:   path.RequestCount,
			ErrorCount:     path.ErrorCount,
			LastStatusCode: path.LastStatusCode,
		})
	}
	return result
}
