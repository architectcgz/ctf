package queries

import (
	"context"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) ListAttackLogs(ctx context.Context, contestID, roundID int64) ([]*dto.AWDAttackLogResp, error) {
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

	resp := make([]*dto.AWDAttackLogResp, 0, len(logs))
	for _, item := range logs {
		logCopy := item
		attackerName := ""
		victimName := ""
		if team := teams[item.AttackerTeamID]; team != nil {
			attackerName = team.Name
		}
		if team := teams[item.VictimTeamID]; team != nil {
			victimName = team.Name
		}
		resp = append(resp, contestdomain.AWDAttackLogRespFromModel(&logCopy, attackerName, victimName))
	}
	return resp, nil
}
