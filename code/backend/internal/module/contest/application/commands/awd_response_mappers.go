package commands

import (
	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
)

func awdTeamServiceRespFromModel(record *model.AWDTeamService, teamName string, serviceName ...string) *dto.AWDTeamServiceResp {
	resp := contestResponseMapperInst.ToAWDTeamServiceRespBasePtr(record)
	if resp == nil {
		return nil
	}
	resp.TeamName = teamName
	resp.CheckResult = contestdomain.ParseAWDCheckResult(record.CheckResult)
	if len(serviceName) > 0 {
		resp.ServiceName = serviceName[0]
		resp.AWDChallengeTitle = serviceName[0]
	}
	return resp
}

func awdAttackLogRespFromModel(record *model.AWDAttackLog, attackerTeam, victimTeam string) *dto.AWDAttackLogResp {
	resp := contestResponseMapperInst.ToAWDAttackLogRespBasePtr(record)
	if resp == nil {
		return nil
	}
	resp.AttackerTeam = attackerTeam
	resp.VictimTeam = victimTeam
	resp.Source = contestdomain.NormalizeAWDAttackSource(record.Source)
	return resp
}
