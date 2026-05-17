package commands

import (
	"context"

	"ctf-platform/internal/model"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) buildUpsertServiceCheckResp(
	ctx context.Context,
	contestID, roundID int64,
	runtimeService *model.ContestAWDService,
	req UpsertServiceCheckInput,
	team *model.Team,
	record *model.AWDTeamService,
) (*AWDTeamServiceResp, error) {
	if s.scoreboardCache != nil {
		if err := s.scoreboardCache.RebuildContestScoreboard(ctx, contestID); err != nil {
			return nil, errcode.ErrInternal.WithCause(err)
		}
	}
	currentRoundID, err := s.resolveCurrentRoundID(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if err := syncAWDServiceStatusField(
		ctx,
		s.stateStore,
		contestID,
		roundID,
		currentRoundID,
		req.TeamID,
		runtimeService.ID,
		req.ServiceStatus,
	); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	serviceResp := awdTeamServiceRespFromModel(record, team.Name)
	if serviceResp == nil {
		return nil, nil
	}
	return &AWDTeamServiceResp{
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
	}, nil
}
