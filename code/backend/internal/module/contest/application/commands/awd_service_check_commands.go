package commands

import (
	"context"
	"errors"
	"time"

	"ctf-platform/internal/dto"
	"ctf-platform/internal/model"
	contestdomain "ctf-platform/internal/module/contest/domain"
	contestports "ctf-platform/internal/module/contest/ports"
	"ctf-platform/pkg/errcode"
)

func (s *AWDService) RunCurrentRoundChecks(ctx context.Context, contestID int64) (*dto.AWDCheckerRunResp, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	if !now.Before(contest.EndTime) {
		return nil, errcode.ErrContestEnded
	}
	if contest.Status != model.ContestStatusRunning && contest.Status != model.ContestStatusFrozen {
		return nil, errcode.ErrContestNotRunning
	}
	round, err := s.resolveCurrentRoundForContest(ctx, contest)
	if err != nil {
		return nil, err
	}

	if s.roundManager == nil {
		return nil, errcode.ErrInternal.WithCause(errors.New("awd round manager is nil"))
	}
	if err := s.roundManager.RunRoundServiceChecks(ctx, contest, round, contestdomain.AWDCheckSourceManualCurrent); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return s.buildCheckerRunResp(ctx, contestID, round)
}

func (s *AWDService) RunRoundChecks(ctx context.Context, contestID, roundID int64) (*dto.AWDCheckerRunResp, error) {
	contest, err := s.ensureAWDContest(ctx, contestID)
	if err != nil {
		return nil, err
	}
	round, err := s.ensureAWDRound(ctx, contestID, roundID)
	if err != nil {
		return nil, err
	}

	if s.roundManager == nil {
		return nil, errcode.ErrInternal.WithCause(errors.New("awd round manager is nil"))
	}
	if err := s.roundManager.RunRoundServiceChecks(ctx, contest, round, contestdomain.AWDCheckSourceManualSelected); err != nil {
		return nil, errcode.ErrInternal.WithCause(err)
	}
	return s.buildCheckerRunResp(ctx, contestID, round)
}

func (s *AWDService) buildCheckerRunResp(ctx context.Context, contestID int64, round *model.AWDRound) (*dto.AWDCheckerRunResp, error) {
	services, err := s.listServices(ctx, contestID, round.ID)
	if err != nil {
		return nil, err
	}
	return &dto.AWDCheckerRunResp{
		Round:    contestdomain.AWDRoundRespFromModel(round),
		Services: services,
	}, nil
}

func (s *AWDService) listServices(ctx context.Context, contestID, roundID int64) ([]*dto.AWDTeamServiceResp, error) {
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
		resp = append(resp, contestdomain.AWDTeamServiceRespFromModel(&recordCopy, teamName))
	}
	return resp, nil
}

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
