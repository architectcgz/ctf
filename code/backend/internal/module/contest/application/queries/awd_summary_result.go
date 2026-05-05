package queries

import "sort"

type AWDRoundSummaryItemResult struct {
	TeamID                  int64
	TeamName                string
	ServiceUpCount          int
	ServiceDownCount        int
	ServiceCompromisedCount int
	AttackScore             int
	DefenseScore            int
	SLAScore                int
	TotalScore              int
	SuccessfulAttackCount   int
	SuccessfulBreachCount   int
	UniqueAttackersAgainst  int
}

type AWDRoundMetricsResult struct {
	TotalServiceCount         int
	ServiceUpCount            int
	ServiceDownCount          int
	ServiceCompromisedCount   int
	AttackedServiceCount      int
	DefenseSuccessCount       int
	TotalAttackCount          int
	SuccessfulAttackCount     int
	FailedAttackCount         int
	SchedulerCheckCount       int
	ManualCurrentRoundChecks  int
	ManualSelectedRoundChecks int
	ManualServiceCheckCount   int
	SubmissionAttackCount     int
	ManualAttackLogCount      int
	LegacyAttackLogCount      int
}

type AWDRoundSummaryResult struct {
	Round   *AWDRoundResult
	Metrics *AWDRoundMetricsResult
	Items   []*AWDRoundSummaryItemResult
}

func sortAWDRoundSummaryItems(items []*AWDRoundSummaryItemResult) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].TotalScore != items[j].TotalScore {
			return items[i].TotalScore > items[j].TotalScore
		}
		return items[i].TeamID < items[j].TeamID
	})
}
