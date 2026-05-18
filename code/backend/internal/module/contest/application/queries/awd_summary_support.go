package queries

import contestentity "ctf-platform/internal/module/contest/entity"

func buildAWDRoundSummary(
	teams map[int64]*contestentity.Team,
	services []contestentity.AWDTeamService,
	attackLogs []contestentity.AWDAttackLog,
) (*AWDRoundMetricsResult, []*AWDRoundSummaryItemResult) {
	items := make(map[int64]*AWDRoundSummaryItemResult, len(teams))
	metrics := &AWDRoundMetricsResult{}
	for teamID, team := range teams {
		items[teamID] = &AWDRoundSummaryItemResult{
			TeamID:   teamID,
			TeamName: team.Name,
		}
	}

	accumulateAWDRoundServiceSummary(items, metrics, services)
	uniqueAttackersAgainst := accumulateAWDRoundAttackSummary(items, metrics, attackLogs)

	respItems := make([]*AWDRoundSummaryItemResult, 0, len(items))
	for teamID, item := range items {
		item.UniqueAttackersAgainst = len(uniqueAttackersAgainst[teamID])
		item.TotalScore = item.AttackScore + item.DefenseScore + item.SLAScore
		respItems = append(respItems, item)
	}
	sortAWDRoundSummaryItems(respItems)
	return metrics, respItems
}
