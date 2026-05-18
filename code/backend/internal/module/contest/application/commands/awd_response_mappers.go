package commands

import (
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestentity "ctf-platform/internal/module/contest/entity"
)

func awdTeamServiceRespFromModel(record *contestentity.AWDTeamService, teamName string, serviceName ...string) *AWDTeamServiceResp {
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

func awdAttackLogRespFromModel(record *contestentity.AWDAttackLog, attackerTeam, victimTeam string) *AWDAttackLogResp {
	resp := contestResponseMapperInst.ToAWDAttackLogRespBasePtr(record)
	if resp == nil {
		return nil
	}
	resp.AttackerTeam = attackerTeam
	resp.VictimTeam = victimTeam
	resp.Source = contestdomain.NormalizeAWDAttackSource(record.Source)
	return resp
}
