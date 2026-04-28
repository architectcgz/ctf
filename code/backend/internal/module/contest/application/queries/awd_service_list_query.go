package queries

import (
	"context"

	"ctf-platform/internal/dto"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

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

	serviceNames, err := s.loadServiceNames(ctx, contestID)
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
		resp = append(resp, contestdomain.AWDTeamServiceRespFromModel(&recordCopy, teamName, serviceNames[record.ServiceID]))
	}
	return resp, nil
}

func (s *AWDService) loadServiceNames(ctx context.Context, contestID int64) (map[int64]string, error) {
	definitions, err := s.repo.ListServiceDefinitionsByContest(ctx, contestID)
	if err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	names := make(map[int64]string, len(definitions))
	for _, definition := range definitions {
		names[definition.ServiceID] = definition.ServiceName
	}
	return names, nil
}
