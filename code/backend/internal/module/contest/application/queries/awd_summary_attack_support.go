package queries

import (
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
)

func accumulateAWDRoundAttackSummary(
	items map[int64]*AWDRoundSummaryItemResult,
	metrics *AWDRoundMetricsResult,
	attackLogs []contestentity.AWDAttackLog,
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
		case contestentity.AWDAttackSourceSubmission:
			metrics.SubmissionAttackCount++
		case contestentity.AWDAttackSourceManual:
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
