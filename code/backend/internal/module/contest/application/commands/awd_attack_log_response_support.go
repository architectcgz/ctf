package commands

import (
	"context"

	contestentity "ctf-platform/internal/module/contest/entity"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) buildAttackLogResponse(
	ctx context.Context,
	contestID, roundID int64,
	req CreateAttackLogInput,
	logRecord *contestentity.AWDAttackLog,
	teams map[int64]*contestentity.Team,
) (*AWDAttackLogResp, error) {
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
		req.VictimTeamID,
		logRecord.ServiceID,
		contestentity.AWDServiceStatusCompromised,
	); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return awdAttackLogRespFromModel(logRecord, teams[req.AttackerTeamID].Name, teams[req.VictimTeamID].Name), nil
}
