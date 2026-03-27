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
		item.TotalScore = item.AttackScore + item.DefenseScore
		respItems = append(respItems, item)
	}
	contestdomain.SortAWDSummaryItems(respItems)
	return metrics, respItems
}

func accumulateAWDRoundServiceSummary(
	items map[int64]*dto.AWDRoundSummaryItem,
	metrics *dto.AWDRoundMetrics,
	services []model.AWDTeamService,
) {
	for _, service := range services {
		metrics.TotalServiceCount++
		if service.AttackReceived > 0 {
			metrics.AttackedServiceCount++
		}
		switch contestdomain.NormalizeAWDCheckSource(contestdomain.ParseAWDCheckResult(service.CheckResult)["check_source"]) {
		case contestdomain.AWDCheckSourceScheduler:
			metrics.SchedulerCheckCount++
		case contestdomain.AWDCheckSourceManualCurrent:
			metrics.ManualCurrentRoundChecks++
		case contestdomain.AWDCheckSourceManualSelected:
			metrics.ManualSelectedRoundChecks++
		case contestdomain.AWDCheckSourceManualService:
			metrics.ManualServiceCheckCount++
		}

		item := items[service.TeamID]
		if item == nil {
			continue
		}
		switch service.ServiceStatus {
		case model.AWDServiceStatusUp:
			metrics.ServiceUpCount++
			if service.AttackReceived > 0 {
				metrics.DefenseSuccessCount++
			}
			item.ServiceUpCount++
		case model.AWDServiceStatusDown:
			metrics.ServiceDownCount++
			item.ServiceDownCount++
		case model.AWDServiceStatusCompromised:
			metrics.ServiceCompromisedCount++
			item.ServiceCompromisedCount++
		}
		item.DefenseScore += service.DefenseScore
	}
}

func accumulateAWDRoundAttackSummary(
	items map[int64]*dto.AWDRoundSummaryItem,
	metrics *dto.AWDRoundMetrics,
	attackLogs []model.AWDAttackLog,
) map[int64]map[int64]struct{} {
	uniqueAttackersAgainst := make(map[int64]map[int64]struct{}, len(items))
	for _, logEntry := range attackLogs {
		metrics.TotalAttackCount++
		if logEntry.IsSuccess {
			metrics.SuccessfulAttackCount++
		} else {
			metrics.FailedAttackCount++
		}
		switch contestdomain.NormalizeAWDAttackSource(logEntry.Source) {
		case model.AWDAttackSourceSubmission:
			metrics.SubmissionAttackCount++
		case model.AWDAttackSourceManual:
			metrics.ManualAttackLogCount++
		default:
			metrics.LegacyAttackLogCount++
		}

		if item := items[logEntry.AttackerTeamID]; item != nil {
			if logEntry.IsSuccess {
				item.SuccessfulAttackCount++
			}
			item.AttackScore += logEntry.ScoreGained
		}
		if !logEntry.IsSuccess {
			continue
		}
		if item := items[logEntry.VictimTeamID]; item != nil {
			item.SuccessfulBreachCount++
		}
		if uniqueAttackersAgainst[logEntry.VictimTeamID] == nil {
			uniqueAttackersAgainst[logEntry.VictimTeamID] = make(map[int64]struct{})
		}
		uniqueAttackersAgainst[logEntry.VictimTeamID][logEntry.AttackerTeamID] = struct{}{}
	}
	return uniqueAttackersAgainst
}
