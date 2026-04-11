package queries

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func buildAWDRoundSummary(
	teams map[int64]*model.Team,
	services []model.AWDTeamService,
	attackLogs []model.AWDAttackLog,
) (*dto.AWDRoundMetrics, []*dto.AWDRoundSummaryItem) {
	items := make(map[int64]*dto.AWDRoundSummaryItem, len(teams))
	metrics := &dto.AWDRoundMetrics{}
	for teamID, team := range teams {
		items[teamID] = &dto.AWDRoundSummaryItem{
			TeamID:   teamID,
			TeamName: team.Name,
		}
	}

	accumulateAWDRoundServiceSummary(items, metrics, services)
	uniqueAttackersAgainst := accumulateAWDRoundAttackSummary(items, metrics, attackLogs)

	respItems := make([]*dto.AWDRoundSummaryItem, 0, len(items))
	for teamID, item := range items {
		item.UniqueAttackersAgainst = len(uniqueAttackersAgainst[teamID])
		item.TotalScore = item.AttackScore + item.DefenseScore + item.SLAScore
		respItems = append(respItems, item)
	}
	contestdomain.SortAWDSummaryItems(respItems)
	return metrics, respItems
}
