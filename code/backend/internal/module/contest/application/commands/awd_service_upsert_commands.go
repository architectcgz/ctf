package commands

import (
	"context"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
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
	if err := s.ensureContestChallenge(ctx, contestID, req.ChallengeID); err != nil {
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

	now := time.Now()
	var record *model.AWDTeamService
	if err := s.repo.WithinTransaction(ctx, func(txRepo contestports.AWDRepository) error {
		var txErr error
		record, txErr = txRepo.UpsertServiceCheck(ctx, roundID, req.TeamID, req.ChallengeID, req.ServiceStatus, checkResult, defenseScore, now)
		if txErr != nil {
			return txErr
		}
		return txRepo.RecalculateContestTeamScores(ctx, contestID)
	}); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	if err := s.repo.RebuildContestScoreboardCache(ctx, s.redis, contestID); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	currentRoundID, err := s.resolveCurrentRoundID(ctx, contestID)
	if err != nil {
		return nil, err
	}
	if err := syncAWDServiceStatusField(ctx, s.redis, contestID, roundID, currentRoundID, req.TeamID, req.ChallengeID, req.ServiceStatus); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}

	return contestdomain.AWDTeamServiceRespFromModel(record, team.Name), nil
}
