package commands

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) UpsertServiceCheck(ctx context.Context, contestID, roundID int64, req *dto.UpsertAWDServiceCheckReq) (*dto.AWDTeamServiceResp, error) {
	round, err := s.ensureAWDRound(ctx, contestID, roundID)
	if err != nil {
		return nil, err
	}
	teamMap, err := s.loadContestTeams(ctx, contestID)
	if err != nil {
		return nil, err
	}
	team, ok := teamMap[req.TeamID]
	if !ok {
		return nil, errcode.ErrNotFound
	}
	runtimeService, err := s.resolveContestRuntimeService(ctx, contestID, req.ServiceID)
	if err != nil {
		return nil, err
	}

	normalizedCheckResult := contestdomain.NormalizeManualAWDCheckResult(req.CheckResult)
	checkResult, err := contestdomain.MarshalAWDCheckResult(normalizedCheckResult)
	if err != nil {
		return nil, errcode.ErrInvalidParams
	}
	defenseScore := 0
	if req.ServiceStatus == model.AWDServiceStatusUp {
		defenseScore = round.DefenseScore
	}

	record, err := s.upsertServiceCheckAndRecalculate(
		ctx,
		contestID,
		roundID,
		runtimeService,
		req,
		checkResult,
		defenseScore,
		time.Now(),
	)
	if err != nil {
		return nil, err
	}
	return s.buildUpsertServiceCheckResp(ctx, contestID, roundID, runtimeService, req, team, record)
}
