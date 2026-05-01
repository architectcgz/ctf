package commands

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) buildCheckerRunResp(ctx context.Context, contestID int64, round *model.AWDRound) (*dto.AWDCheckerRunResp, error) {
	services, err := s.listRoundServices(ctx, contestID, round.ID)
	if err != nil {
		return nil, err
	}
	return &dto.AWDCheckerRunResp{
		Round:    awdRoundRespFromModel(round),
		Services: services,
	}, nil
}

func (s *AWDService) listRoundServices(ctx context.Context, contestID, roundID int64) ([]*dto.AWDTeamServiceResp, error) {
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
		resp = append(resp, awdTeamServiceRespFromModel(&recordCopy, teamName))
	}
	return resp, nil
}
