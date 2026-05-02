package commands

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func awdRoundRespFromModel(round *model.AWDRound) *dto.AWDRoundResp {
	if round == nil {
		return nil
	}
	mapped := contestResponseMapperInst.ToAWDRoundRespBase(*round)
	return &mapped
}

func awdTeamServiceRespFromModel(record *model.AWDTeamService, teamName string, serviceName ...string) *dto.AWDTeamServiceResp {
	if record == nil {
		return nil
	}
	mapped := contestResponseMapperInst.ToAWDTeamServiceRespBase(*record)
	resp := &mapped
	resp.TeamName = teamName
	resp.CheckResult = contestdomain.ParseAWDCheckResult(record.CheckResult)
	if len(serviceName) > 0 {
		resp.ServiceName = serviceName[0]
		resp.AWDChallengeTitle = serviceName[0]
	}
	return resp
}

func awdAttackLogRespFromModel(record *model.AWDAttackLog, attackerTeam, victimTeam string) *dto.AWDAttackLogResp {
	if record == nil {
		return nil
	}
	mapped := contestResponseMapperInst.ToAWDAttackLogRespBase(*record)
	mapped.AttackerTeam = attackerTeam
	mapped.VictimTeam = victimTeam
	mapped.Source = contestdomain.NormalizeAWDAttackSource(record.Source)
	return &mapped
}
