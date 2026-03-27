package domain

import (
	"sort"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
)

func AWDRoundRespFromModel(round *model.AWDRound) *dto.AWDRoundResp {
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

func AWDTeamServiceRespFromModel(record *model.AWDTeamService, teamName string) *dto.AWDTeamServiceResp {
	if record == nil {
		return nil
	}
	return &dto.AWDTeamServiceResp{
		ID:             record.ID,
		RoundID:        record.RoundID,
		TeamID:         record.TeamID,
		TeamName:       teamName,
		ChallengeID:    record.ChallengeID,
		ServiceStatus:  record.ServiceStatus,
		CheckResult:    ParseAWDCheckResult(record.CheckResult),
		AttackReceived: record.AttackReceived,
		DefenseScore:   record.DefenseScore,
		AttackScore:    record.AttackScore,
		UpdatedAt:      record.UpdatedAt,
	}
}

func AWDAttackLogRespFromModel(record *model.AWDAttackLog, attackerTeam, victimTeam string) *dto.AWDAttackLogResp {
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
		ChallengeID:    record.ChallengeID,
		AttackType:     record.AttackType,
		Source:         NormalizeAWDAttackSource(record.Source),
		SubmittedFlag:  record.SubmittedFlag,
		IsSuccess:      record.IsSuccess,
		ScoreGained:    record.ScoreGained,
		CreatedAt:      record.CreatedAt,
	}
}

func SortAWDSummaryItems(items []*dto.AWDRoundSummaryItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].TotalScore != items[j].TotalScore {
			return items[i].TotalScore > items[j].TotalScore
		}
		return items[i].TeamID < items[j].TeamID
	})
}
