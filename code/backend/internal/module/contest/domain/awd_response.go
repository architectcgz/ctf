package domain

import (
	"sort"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func AWDRoundRespFromModel(round *model.AWDRound) *dto.AWDRoundResp {
	return contestResponseMapperInst.ToAWDRoundResp(round)
}

func AWDTeamServiceRespFromModel(record *model.AWDTeamService, teamName string, serviceName ...string) *dto.AWDTeamServiceResp {
	if record == nil {
		return nil
	}
	resp := contestResponseMapperInst.ToAWDTeamServiceResp(record)
	resp.TeamName = teamName
	if len(serviceName) > 0 {
		resp.ServiceName = serviceName[0]
		resp.AWDChallengeTitle = serviceName[0]
	}
	return resp
}

func AWDAttackLogRespFromModel(record *model.AWDAttackLog, attackerTeam, victimTeam string) *dto.AWDAttackLogResp {
	if record == nil {
		return nil
	}
	resp := contestResponseMapperInst.ToAWDAttackLogResp(record)
	resp.AttackerTeam = attackerTeam
	resp.VictimTeam = victimTeam
	resp.Source = NormalizeAWDAttackSource(record.Source)
	return resp
}

func SortAWDSummaryItems(items []*dto.AWDRoundSummaryItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].TotalScore != items[j].TotalScore {
			return items[i].TotalScore > items[j].TotalScore
		}
		return items[i].TeamID < items[j].TeamID
	})
}
