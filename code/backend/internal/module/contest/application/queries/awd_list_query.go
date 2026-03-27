package queries

import (
	"context"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) ListRounds(ctx context.Context, contestID int64) ([]*dto.AWDRoundResp, error) {
	if _, err := s.ensureAWDContest(ctx, contestID); err != nil {
		return nil, err
	}

	rounds, err := s.repo.ListRoundsByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	resp := make([]*dto.AWDRoundResp, 0, len(rounds))
	for _, round := range rounds {
		roundCopy := round
		resp = append(resp, contestdomain.AWDRoundRespFromModel(&roundCopy))
	}
	return resp, nil
}

func (s *AWDService) ListServices(ctx context.Context, contestID, roundID int64) ([]*dto.AWDTeamServiceResp, error) {
	if _, err := s.ensureAWDRound(ctx, contestID, roundID); err != nil {
		return nil, err
	}

	records, err := s.repo.ListServicesByRound(ctx, roundID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	teams, err := s.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}

	resp := make([]*dto.AWDTeamServiceResp, 0, len(records))
	for _, record := range records {
		recordCopy := record
		teamName := ""
		if team := teams[record.TeamID]; team != nil {
			teamName = team.Name
		}
		resp = append(resp, contestdomain.AWDTeamServiceRespFromModel(&recordCopy, teamName))
	}
	return resp, nil
}

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
