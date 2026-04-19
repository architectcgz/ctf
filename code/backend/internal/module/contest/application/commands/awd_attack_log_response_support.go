package commands

import (
	"context"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) buildAttackLogResponse(
	ctx context.Context,
	contestID, roundID int64,
	req *dto.CreateAWDAttackLogReq,
	logRecord *model.AWDAttackLog,
	teams map[int64]*model.Team,
) (*dto.AWDAttackLogResp, error) {
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
		req.VictimTeamID,
		logRecord.ServiceID,
		model.AWDServiceStatusCompromised,
	); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return contestdomain.AWDAttackLogRespFromModel(logRecord, teams[req.AttackerTeamID].Name, teams[req.VictimTeamID].Name), nil
}
