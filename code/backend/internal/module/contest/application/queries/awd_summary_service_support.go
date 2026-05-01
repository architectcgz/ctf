package queries

import (
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func accumulateAWDRoundServiceSummary(
	items map[int64]*AWDRoundSummaryItemResult,
	metrics *AWDRoundMetricsResult,
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
		item.SLAScore += service.SLAScore
		item.DefenseScore += service.DefenseScore
	}
}
