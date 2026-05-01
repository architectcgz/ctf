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
	return &dto.AWDRoundResp{
		ID:           round.ID,
		ContestID:    round.ContestID,
		RoundNumber:  round.RoundNumber,
		Status:       round.Status,
		StartedAt:    round.StartedAt,
		EndedAt:      round.EndedAt,
		AttackScore:  round.AttackScore,
		DefenseScore: round.DefenseScore,
		CreatedAt:    round.CreatedAt,
		UpdatedAt:    round.UpdatedAt,
	}
}

func awdTeamServiceRespFromModel(record *model.AWDTeamService, teamName string, serviceName ...string) *dto.AWDTeamServiceResp {
	if record == nil {
		return nil
	}
	resp := &dto.AWDTeamServiceResp{
		ID:             record.ID,
		RoundID:        record.RoundID,
		TeamID:         record.TeamID,
		TeamName:       teamName,
		ServiceID:      record.ServiceID,
		AWDChallengeID: record.AWDChallengeID,
		ServiceStatus:  record.ServiceStatus,
		CheckResult:    contestdomain.ParseAWDCheckResult(record.CheckResult),
		CheckerType:    record.CheckerType,
		AttackReceived: record.AttackReceived,
		SLAScore:       record.SLAScore,
		DefenseScore:   record.DefenseScore,
		AttackScore:    record.AttackScore,
		UpdatedAt:      record.UpdatedAt,
	}
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
	return &dto.AWDAttackLogResp{
		ID:             record.ID,
		RoundID:        record.RoundID,
		AttackerTeamID: record.AttackerTeamID,
		AttackerTeam:   attackerTeam,
		VictimTeamID:   record.VictimTeamID,
		VictimTeam:     victimTeam,
		ServiceID:      record.ServiceID,
		AWDChallengeID: record.AWDChallengeID,
		AttackType:     record.AttackType,
		Source:         contestdomain.NormalizeAWDAttackSource(record.Source),
		SubmittedFlag:  record.SubmittedFlag,
		IsSuccess:      record.IsSuccess,
		ScoreGained:    record.ScoreGained,
		CreatedAt:      record.CreatedAt,
	}
}
