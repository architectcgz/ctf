package commands

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) buildUpsertServiceCheckResp(
	ctx context.Context,
	contestID, roundID int64,
	runtimeService *model.ContestAWDService,
	req *dto.UpsertAWDServiceCheckReq,
	team *model.Team,
	record *model.AWDTeamService,
) (*dto.AWDTeamServiceResp, error) {
	if err := s.repo.RebuildContestScoreboardCache(ctx, s.redis, contestID); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	currentRoundID, err := s.resolveCurrentRoundID(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if err := syncAWDServiceStatusField(
		ctx,
		s.redis,
		contestID,
		roundID,
		currentRoundID,
		req.TeamID,
		runtimeService.ID,
		req.ServiceStatus,
	); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return contestdomain.AWDTeamServiceRespFromModel(record, team.Name), nil
}
