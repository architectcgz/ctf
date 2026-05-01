package queries

import (
	"context"

	"ctf-platform/pkg/errcode"
)

func (s *AWDService) ListAttackLogs(ctx context.Context, contestID, roundID int64) ([]AWDAttackLogResult, error) {
	if _, err := s.ensureAWDRound(ctx, contestID, roundID); err != nil {
		return nil, err
	}

	logs, err := s.repo.ListAttackLogsByRound(ctx, roundID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teams, err := s.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}

	resp := make([]AWDAttackLogResult, 0, len(logs))
	for _, item := range logs {
		attackerName := ""
		victimName := ""
		if team := teams[item.AttackerTeamID]; team != nil {
			attackerName = team.Name
		}
		if team := teams[item.VictimTeamID]; team != nil {
			victimName = team.Name
		}
		resp = append(resp, AWDAttackLogResult{
			ID:             item.ID,
			RoundID:        item.RoundID,
			AttackerTeamID: item.AttackerTeamID,
			AttackerTeam:   attackerName,
			VictimTeamID:   item.VictimTeamID,
			VictimTeam:     victimName,
			ServiceID:      item.ServiceID,
			AWDChallengeID: item.AWDChallengeID,
			AttackType:     item.AttackType,
			Source:         item.Source,
			SubmittedFlag:  item.SubmittedFlag,
			IsSuccess:      item.IsSuccess,
			ScoreGained:    item.ScoreGained,
			CreatedAt:      item.CreatedAt,
		})
	}
	return resp, nil
}
