package commands

import (
	"context"

	"ctf-platform/internal/apperror"
	contestentity "ctf-platform/internal/module/contest/entity"
)

func (s *AWDService) buildCheckerRunResp(ctx context.Context, contestID int64, round *contestentity.AWDRound) (*AWDCheckerRunResp, error) {
	services, err := s.listRoundServices(ctx, contestID, round.ID)
	if err != nil {
		return nil, err
	}
	roundResp := contestResponseMapperInst.ToAWDRoundRespBasePtr(round)
	if roundResp == nil {
		return nil, nil
	}
	return &AWDCheckerRunResp{
		Round: &AWDRoundResp{
			ID:           roundResp.ID,
			ContestID:    roundResp.ContestID,
			RoundNumber:  roundResp.RoundNumber,
			Status:       roundResp.Status,
			StartedAt:    roundResp.StartedAt,
			EndedAt:      roundResp.EndedAt,
			AttackScore:  roundResp.AttackScore,
			DefenseScore: roundResp.DefenseScore,
			CreatedAt:    roundResp.CreatedAt,
			UpdatedAt:    roundResp.UpdatedAt,
		},
		Services: services,
	}, nil
}

func (s *AWDService) listRoundServices(ctx context.Context, contestID, roundID int64) ([]*AWDTeamServiceResp, error) {
	if _, err := s.ensureAWDRound(ctx, contestID, roundID); err != nil {
		return nil, err
	}

	records, err := s.repo.ListServicesByRound(ctx, roundID)
	if err != nil {
		return nil, apperror.ErrInternal.WithCause(err)
	}

	teams, err := s.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}

	resp := make([]*AWDTeamServiceResp, 0, len(records))
	for _, record := range records {
		recordCopy := record
		teamName := ""
		if team := teams[record.TeamID]; team != nil {
			teamName = team.Name
		}
		serviceResp := awdTeamServiceRespFromModel(&recordCopy, teamName)
		if serviceResp == nil {
			resp = append(resp, nil)
			continue
		}
		resp = append(resp, &AWDTeamServiceResp{
			ID:                serviceResp.ID,
			RoundID:           serviceResp.RoundID,
			TeamID:            serviceResp.TeamID,
			TeamName:          serviceResp.TeamName,
			ServiceID:         serviceResp.ServiceID,
			ServiceName:       serviceResp.ServiceName,
			AWDChallengeID:    serviceResp.AWDChallengeID,
			AWDChallengeTitle: serviceResp.AWDChallengeTitle,
			ServiceStatus:     serviceResp.ServiceStatus,
			CheckResult:       serviceResp.CheckResult,
			CheckerType:       serviceResp.CheckerType,
			AttackReceived:    serviceResp.AttackReceived,
			SLAScore:          serviceResp.SLAScore,
			DefenseScore:      serviceResp.DefenseScore,
			AttackScore:       serviceResp.AttackScore,
			UpdatedAt:         serviceResp.UpdatedAt,
		})
	}
	return resp, nil
}
