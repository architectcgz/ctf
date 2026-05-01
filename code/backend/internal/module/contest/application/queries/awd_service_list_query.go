package queries

import (
	"context"

	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) ListServices(ctx context.Context, contestID, roundID int64) ([]AWDTeamServiceResult, error) {
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

	resp := make([]AWDTeamServiceResult, 0, len(records))
	for _, record := range records {
		teamName := ""
		if team := teams[record.TeamID]; team != nil {
			teamName = team.Name
		}
		serviceName := serviceNames[record.ServiceID]
		resp = append(resp, AWDTeamServiceResult{
			ID:                record.ID,
			RoundID:           record.RoundID,
			TeamID:            record.TeamID,
			TeamName:          teamName,
			ServiceID:         record.ServiceID,
			ServiceName:       serviceName,
			AWDChallengeID:    record.AWDChallengeID,
			AWDChallengeTitle: serviceName,
			ServiceStatus:     record.ServiceStatus,
			CheckResult:       contestdomain.ParseAWDCheckResult(record.CheckResult),
			CheckerType:       string(record.CheckerType),
			AttackReceived:    record.AttackReceived,
			SLAScore:          record.SLAScore,
			DefenseScore:      record.DefenseScore,
			AttackScore:       record.AttackScore,
			UpdatedAt:         record.UpdatedAt,
		})
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
